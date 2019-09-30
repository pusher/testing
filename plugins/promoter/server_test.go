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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gtypes "github.com/onsi/gomega/types"
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

		push := func(org, repo, dir, newBranch string) error {
			pushedBranches = append(pushedBranches, newBranch)
			return nil
		}

		server = &Server{
			tokenGenerator: getSecret,
			botName:        "ci-bot",

			gc:   c,
			push: push,
			ghc:  ghc,
			log:  logrus.StandardLogger().WithField("client", "promoter"),
			wg:   &sync.WaitGroup{},
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
			server.wg.Wait()
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

			It("creates a PR to promote the merged Pull Request", func() {
				Expect(ghc.FakeClient.PullRequests).To(ContainElement(SatisfyAll(
					withField("Title", Equal("Promote to target: PR Title")),
					withField("Body", Equal(expectedPRBody)),
					withField("Base", Equal(github.PullRequestBranch{
						Ref: "target",
						Repo: github.Repo{
							Name: "bar",
							Owner: github.User{
								Login: "foo",
							},
						},
					})),
					withField("Head", Equal(github.PullRequestBranch{
						Ref: "pr-123",
						Repo: github.Repo{
							Name: "bar",
							Owner: github.User{
								Login: "foo",
							},
						},
					})),
				)))
			})

			It("comments on the merged PR", func() {
				Expect(ghc.FakeClient.IssueComments).To(HaveKeyWithValue(123, ContainElement(
					withField("Body", Equal("Automated promotion PR created #0")),
				)))
			})
		})
	})

	Context("ServeHTTP", func() {
		var w *httptest.ResponseRecorder
		var r *http.Request
		var event *github.PullRequestEvent

		BeforeEach(func() {
			event = &github.PullRequestEvent{
				Action:      github.PullRequestActionClosed,
				Number:      123,
				PullRequest: *ghc.FakeClient.PullRequests[123],
				Repo: github.Repo{
					Owner: github.User{Login: "foo"},
					Name:  "bar",
				},
			}

			payload, err := json.Marshal(event)
			Expect(err).ToNot(HaveOccurred())

			buf := &bytes.Buffer{}
			buf.Write(payload)

			w = httptest.NewRecorder()
			r = httptest.NewRequest("POST", "/", buf)

			// Add "hook" required headers
			// Use `push` so we don't actually create the PRs
			r.Header.Add("X-GitHub-Event", "push")
			r.Header.Add("X-GitHub-Delivery", "abcdef")
			r.Header.Add("X-Hub-Signature", github.PayloadSignature(payload, server.tokenGenerator()))
			r.Header.Add("content-type", "application/json")
		})

		JustBeforeEach(func() {
			server.ServeHTTP(w, r)
		})

		Context("without a source or target paramenter", func() {
			It("should return a 400 status code", func() {
				Expect(w.Result().StatusCode).To(Equal(400))
			})

			It("should return an error in the response body", func() {
				body, _ := ioutil.ReadAll(w.Result().Body)
				Expect(string(body)).To(Equal("400 Bad Request: Missing source parameter\n"))
			})
		})

		Context("with a source, but without a target paramenter", func() {
			BeforeEach(func() {
				r.URL.RawQuery = "source=foo"
			})

			It("should return a 400 status code", func() {
				Expect(w.Result().StatusCode).To(Equal(400))
			})

			It("should return an error in the response body", func() {
				body, _ := ioutil.ReadAll(w.Result().Body)
				Expect(string(body)).To(Equal("400 Bad Request: Missing target parameter\n"))
			})
		})

		Context("with both a source and target parameter", func() {
			BeforeEach(func() {
				r.URL.RawQuery = "source=foo&target=bar"
			})

			It("should return a 200 status code", func() {
				Expect(w.Result().StatusCode).To(Equal(200))
			})

			It("should return an empty response body", func() {
				body, _ := ioutil.ReadAll(w.Result().Body)
				Expect(body).To(BeEmpty())
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

var expectedPRBody = `This is an automated promotion of PR #123

---
PR Body

---
/assign @baz
/cc @baz`

// withField gets the value of the named field from the object
func withField(field string, matcher gtypes.GomegaMatcher) gtypes.GomegaMatcher {
	// Addressing Field by <struct>.<field> can be recursed
	fields := strings.SplitN(field, ".", 2)
	if len(fields) == 2 {
		matcher = withField(fields[1], matcher)
	}

	return WithTransform(func(obj interface{}) interface{} {
		r := reflect.ValueOf(obj)
		f := reflect.Indirect(r).FieldByName(fields[0])
		if !f.IsValid() {
			panic(fmt.Sprintf("Object '%s' does not have a field '%s'", reflect.TypeOf(obj), fields[0]))
		}
		return f.Interface()
	}, matcher)
}
