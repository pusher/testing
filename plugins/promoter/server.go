/*
Copyright 2017 The Kubernetes Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"k8s.io/test-infra/prow/git"
	"k8s.io/test-infra/prow/github"
	"k8s.io/test-infra/prow/pluginhelp"
)

type githubClient interface {
	CreatePullRequest(org, repo, title, body, head, base string, canModify bool) (int, error)
	CreateComment(org, repo string, number int, comment string) error
}

// HelpProvider construct the pluginhelp.PluginHelp for this plugin.
func HelpProvider(enabledRepos []string) (*pluginhelp.PluginHelp, error) {
	pluginHelp := &pluginhelp.PluginHelp{
		Description: `The promoter plugin is used for promoting PRs across branches. For every successful promoter invocation a new PR is opened against the target branch and assigned to the requester.`,
	}
	pluginHelp.AddCommand(pluginhelp.Command{
		Description: "Promote changes to a different branch. This plugin automatically promotes PRs once merged to the configured target branch.",
	})
	return pluginHelp, nil
}

// Server implements http.Handler. It validates incoming GitHub webhooks and
// then dispatches them to the appropriate plugins.
type Server struct {
	tokenGenerator func() []byte
	botName        string
	email          string

	gc *git.Client
	// Used for unit testing
	push func(repo, newBranch string) error
	ghc  githubClient
	log  *logrus.Entry
	wg   *sync.WaitGroup
}

// ServeHTTP validates an incoming webhook and puts it into the event channel.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	eventType, eventGUID, payload, ok, _ := github.ValidateWebhook(w, r, s.tokenGenerator())
	if !ok {
		return
	}

	sources, targets, ok := validateParams(w, r)
	if !ok {
		return
	}

	if err := s.handleEvent(eventType, eventGUID, payload, sources, targets); err != nil {
		s.log.WithError(err).Error("Error parsing event.")
	}
}

func validateParams(w http.ResponseWriter, r *http.Request) ([]string, []string, bool) {
	params := r.URL.Query()

	sources, ok := params["source"]
	if !ok || len(sources) == 0 {
		responseHTTPError(w, http.StatusBadRequest, "400 Bad Request: Missing source parameter")
		return []string{}, []string{}, false
	}

	targets, ok := params["target"]
	if !ok || len(targets) == 0 {
		responseHTTPError(w, http.StatusBadRequest, "400 Bad Request: Missing target parameter")
		return []string{}, []string{}, false
	}

	return sources, targets, true
}

func responseHTTPError(w http.ResponseWriter, statusCode int, response string) {
	logrus.WithFields(logrus.Fields{
		"response":    response,
		"status-code": statusCode,
	}).Debug(response)
	http.Error(w, response, statusCode)
}

func (s *Server) handleEvent(eventType, eventGUID string, payload []byte, sources, targets []string) error {
	l := s.log.WithFields(logrus.Fields{
		"event-type":     eventType,
		github.EventGUID: eventGUID,
	})
	switch eventType {
	case "pull_request":
		var pr github.PullRequestEvent
		if err := json.Unmarshal(payload, &pr); err != nil {
			return err
		}
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			if err := s.handlePullRequest(l, pr, sources, targets); err != nil {
				s.log.WithError(err).WithFields(l.Data).Info("Promote failed.")
			}
		}()
	default:
		l.Debugf("skipping event of type %q", eventType)
	}
	return nil
}

func (s *Server) handlePullRequest(l *logrus.Entry, pre github.PullRequestEvent, sources, targets []string) error {
	// Only consider newly merged PRs
	if pre.Action != github.PullRequestActionClosed {
		return nil
	}

	pr := pre.PullRequest
	if !pr.Merged || pr.MergeSHA == nil {
		return nil
	}

	org := pr.Base.Repo.Owner.Login
	repo := pr.Base.Repo.Name
	baseBranch := pr.Base.Ref
	num := pr.Number
	title := pr.Title
	body := pr.Body

	l = l.WithFields(logrus.Fields{
		github.OrgLogField:  org,
		github.RepoLogField: repo,
		github.PrLogField:   num,
	})

	if !contains(sources, baseBranch) {
		l.Debugf("skipping PR %d as base branch (%s) is not one of %v", num, baseBranch, sources)
		return nil
	}

	// Create a new branch at the head of the base branch and push it
	err := s.createPromotionBranch(l, org, repo, baseBranch, num)
	if err != nil {
		return fmt.Errorf("error creating branch: %v", err)
	}

	// Make sure it compiles before we implement the behaviour
	l.Info(baseBranch, title, body)

	//TODO: Implement handling logic
	return nil
}

func (s *Server) createPromotionBranch(l *logrus.Entry, org, repo, baseBranch string, prNumber int) error {
	// Clone the repo, checkout the base branch.
	startClone := time.Now()
	r, err := s.gc.Clone(org + "/" + repo)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Clean(); err != nil {
			s.log.WithError(err).WithFields(l.Data).Error("Error cleaning up repo.")
		}
	}()

	err = r.Checkout(baseBranch)
	if err != nil {
		resp := fmt.Sprintf("cannot checkout %s: %v", baseBranch, err)
		s.log.WithFields(l.Data).Info(resp)
		return s.createComment(org, repo, prNumber, resp)
	}
	s.log.WithFields(l.Data).WithField("duration", time.Since(startClone)).Info("Cloned and checked out source branch: ", baseBranch)

	newBranch := fmt.Sprintf("pr-%d", prNumber)
	err = r.CheckoutNewBranch(newBranch)
	if err != nil {
		resp := fmt.Sprintf("cannot create new branch %s: %v", newBranch, err)
		s.log.WithFields(l.Data).Info(resp)
		return s.createComment(org, repo, prNumber, resp)
	}
	s.log.WithFields(l.Data).Info("Checked out promotion branch: ", newBranch)

	push := r.Push
	if s.push != nil {
		push = s.push
	}

	// Push the new branch back to the origin
	if err := push(repo, newBranch); err != nil {
		resp := fmt.Sprintf("failed to push promotion branch: %v", err)
		s.log.WithFields(l.Data).Info(resp)
		return s.createComment(org, repo, prNumber, resp)
	}
	s.log.WithFields(l.Data).Info("Pushed promotion branch to remote: ", newBranch)

	return nil
}

func (s *Server) createComment(org, repo string, num int, resp string) error {
	return s.ghc.CreateComment(org, repo, num, fmt.Sprintf("Error occurred promoting branch: %s", resp))
}

func contains(list []string, toFind string) bool {
	for _, item := range list {
		if item == toFind {
			return true
		}
	}
	return false
}
