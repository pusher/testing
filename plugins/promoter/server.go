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
	"os/exec"
	"strings"
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
	tokenGenerator   func() []byte
	botName          string
	botPassGenerator func() []byte
	email            string

	gc *git.Client
	// Used for unit testing
	push func(org, repo, dir, branch string) error
	ghc  githubClient
	git  string
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
	author := pr.User.Login

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

	var errs []string
	// Create a new PR for each target branch
	for _, target := range targets {
		err = s.createPromotionPR(l, org, repo, target, body, author, title, num)
		if err != nil {
			errs = append(errs, err.Error())
		}
	}

	if len(errs) > 0 {
		errors := strings.Join(errs, ", ")
		return fmt.Errorf("error(s) encountered creating promotion PR: %s", errors)
	}

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
		cErr := s.createComment(org, repo, prNumber, resp)
		if cErr != nil {
			return cErr
		}
		return err
	}
	s.log.WithFields(l.Data).WithField("duration", time.Since(startClone)).Info("Cloned and checked out source branch: ", baseBranch)

	newBranch := fmt.Sprintf("pr-%d", prNumber)
	err = r.CheckoutNewBranch(newBranch)
	if err != nil {
		resp := fmt.Sprintf("cannot create new branch %s: %v", newBranch, err)
		s.log.WithFields(l.Data).Info(resp)
		cErr := s.createComment(org, repo, prNumber, resp)
		if cErr != nil {
			return cErr
		}
		return err
	}
	s.log.WithFields(l.Data).Info("Checked out promotion branch: ", newBranch)

	push := s.gitPush
	if s.push != nil {
		push = s.push
	}

	// Push the new branch back to the origin
	if err := push(org, repo, r.Dir, newBranch); err != nil {
		resp := fmt.Sprintf("failed to push promotion branch: %v", err)
		s.log.WithFields(l.Data).Info(resp)
		cErr := s.createComment(org, repo, prNumber, resp)
		if cErr != nil {
			return cErr
		}
		return err
	}
	s.log.WithFields(l.Data).Info("Pushed promotion branch to remote: ", newBranch)

	return nil
}

func (s *Server) createPromotionPR(l *logrus.Entry, org, repo, targetBranch, prBody, prAuthor, prTitle string, prNumber int) error {
	promotionTitle := fmt.Sprintf("Promote to %s: %s", targetBranch, prTitle)
	// Construct PR Body
	promotionBody := fmt.Sprintf("This is an automated promotion of PR #%d", prNumber)
	// Append original PR body to PR description
	promotionBody = fmt.Sprintf("%s\n\n---\n%s", promotionBody, prBody)
	// Assign original author and request their review
	promotionBody = fmt.Sprintf("%s\n\n---\n/assign @%s\n/cc @%s", promotionBody, prAuthor, prAuthor)

	// Create the promotion PR
	promotionBranch := fmt.Sprintf("pr-%d", prNumber)
	createdNum, err := s.ghc.CreatePullRequest(org, repo, promotionTitle, promotionBody, promotionBranch, targetBranch, true)
	if err != nil {
		resp := fmt.Sprintf("new pull request could not be created: %v", err)
		s.log.WithFields(l.Data).Info(resp)
		cErr := s.createComment(org, repo, prNumber, resp)
		if cErr != nil {
			return cErr
		}
		return err
	}

	// Comment on the source PR that we have created a promotion PR
	resp := fmt.Sprintf("new pull request created: #%d", createdNum)
	s.log.WithFields(l.Data).Info(resp)
	err = s.ghc.CreateComment(org, repo, prNumber, fmt.Sprintf("Automated promotion PR created #%d", createdNum))
	if err != nil {
		return err
	}

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

// Push pushes over https to the provided owner/repo#branch using a password
// for basic auth.
func (s *Server) gitPush(org, repo, dir, branch string) error {
	if s.botName == "" || s.getBotPass() == "" {
		return fmt.Errorf("cannot push without credentials - configure your git client")
	}
	s.log.Infof("Pushing to '%s/%s (branch: %s)'.", org, repo, branch)
	remote := fmt.Sprintf("https://%s:%s@%s/%s/%s", s.botName, s.getBotPass(), "github.com", org, repo)
	co := exec.Command(s.git, "push", remote, branch)
	co.Dir = dir
	out, err := co.CombinedOutput()
	if err != nil {
		s.log.Errorf("Pushing failed with error: %v and output: %q", err, string(out))
		return fmt.Errorf("pushing failed with error: %v and output: %q", err, string(out))
	}
	return nil
}

func (s *Server) getBotPass() string {
	return string(s.botPassGenerator())
}
