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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"k8s.io/test-infra/prow/git"
	"k8s.io/test-infra/prow/git/localgit"
	"k8s.io/test-infra/prow/github"
	"k8s.io/test-infra/prow/github/fakegithub"
)

var _ = Describe("Promoter suite", func() {
	var server *Server
	var lg *localgit.LocalGit
	var c *git.Client
	var ghc *fakeClient
	var pushedBranches []string

	var AssertDoesNotCreatePR = func() {
		It("does not create a new PR", func() {
			Expect(ghc.FakeClient.PullRequests).To(SatisfyAll(HaveLen(1), HaveKey(123)))
		})

		It("does not push any new branches", func() {
			Expect(pushedBranches).To(BeEmpty())
		})
	}

	BeforeEach(func() {
		var err error
		lg, c, err = localgit.New()
		Expect(err).ToNot(HaveOccurred())
		setupFakeRepo(lg)

		getSecret := func() []byte {
			return []byte("sha=abcdefg")
		}

		pushedBranches = []string{}

		ghc = &fakeClient{FakeClient: &fakegithub.FakeClient{
			PullRequests:  make(map[int]*github.PullRequest),
			IssueComments: make(map[int][]github.IssueComment),
		}}
		Expect(ghc.FakeClient.PullRequests).ToNot(BeNil())
		Expect(ghc.FakeClient.IssueComments).ToNot(BeNil())

		mergeSHA := "abc1234"
		ghc.FakeClient.PullRequests[123] = &github.PullRequest{
			Number: 123,
			Base: github.PullRequestBranch{
				Ref: "source",
				Repo: github.Repo{
					Owner: github.User{Login: "foo"},
					Name:  "bar",
				},
			},
			Title:    "PR Title",
			Body:     "PR Body",
			User:     github.User{Login: "baz"},
			Merged:   true,
			MergeSHA: &mergeSHA,
		}

		push := func(repo, newBranch string) error {
			pushedBranches = append(pushedBranches, newBranch)
			return nil
		}

		server = &Server{
			tokenGenerator: getSecret,
			botName:        "ci-bot",
			email:          "ci-bot@foo",

			gc:   c,
			push: push,
			ghc:  ghc,
			log:  logrus.StandardLogger().WithField("client", "promoter"),
		}

	})

	AfterEach(func() {
		Expect(lg.Clean()).ToNot(HaveOccurred())
		Expect(c.Clean()).ToNot(HaveOccurred())
	})

	Context("handleEvent", func() {
		var eventType, eventGUID string
		var event *github.PullRequestEvent
		var sources, targets []string
		var handleErr error

		BeforeEach(func() {
			eventType = "pull_request"
			eventGUID = "abcdef"
			sources = []string{"source"}
			targets = []string{"target"}

			event = &github.PullRequestEvent{
				Action:      github.PullRequestActionClosed,
				Number:      123,
				PullRequest: *ghc.FakeClient.PullRequests[123],
				Repo: github.Repo{
					Owner: github.User{Login: "foo"},
					Name:  "bar",
				},
			}
		})

		JustBeforeEach(func() {
			payload, err := json.Marshal(event)
			Expect(err).ToNot(HaveOccurred())
			handleErr = server.handleEvent(eventType, eventGUID, payload, sources, targets)
		})

		Context("when the eventType is not pull_request", func() {
			BeforeEach(func() {
				eventType = "not_pull_request"
			})

			It("does not return an error", func() {
				Expect(handleErr).ToNot(HaveOccurred())
			})

			AssertDoesNotCreatePR()
		})

		Context("when the event is not a closure event", func() {
			BeforeEach(func() {
				event.Action = github.PullRequestActionOpened
			})

			It("does not return an error", func() {
				Expect(handleErr).ToNot(HaveOccurred())
			})

			AssertDoesNotCreatePR()
		})

		Context("when the PR has not yet been merged", func() {
			BeforeEach(func() {
				event.PullRequest.Merged = false
			})

			It("does not return an error", func() {
				Expect(handleErr).ToNot(HaveOccurred())
			})

			AssertDoesNotCreatePR()
		})

		Context("when the PR is merged into a branch not in sources", func() {
			BeforeEach(func() {
				event.PullRequest.Base.Ref = "not-in-sources"
			})

			It("does not return an error", func() {
				Expect(handleErr).ToNot(HaveOccurred())
			})

			AssertDoesNotCreatePR()
		})

		Context("when the PR should be promoted", func() {
			It("does not return an error", func() {
				Expect(handleErr).ToNot(HaveOccurred())
			})

			It("pushes a new branch for the merged Pull Request", func() {
				Expect(pushedBranches).Should(ContainElement(Equal("pr-123")))
			})
		})
	})

})

func setupFakeRepo(lg *localgit.LocalGit) {
	Expect(lg.MakeFakeRepo("foo", "bar")).ToNot(HaveOccurred())
	Expect(lg.CheckoutNewBranch("foo", "bar", "source")).ToNot(HaveOccurred())
	Expect(lg.AddCommit("foo", "bar", initialFiles)).ToNot(HaveOccurred())
}

var initialFiles = map[string][]byte{
	"bar.go": []byte(`// Package bar does an interesting thing.
package bar
// Foo does a thing.
func Foo(wow int) int {
	return 42 + wow
}
`),
}
