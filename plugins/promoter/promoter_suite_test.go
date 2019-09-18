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
	"flag"
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/ginkgo/types"
	. "github.com/onsi/gomega"
)

var (
	// reportDir is used to set the output directory for JUnit artifacts
	reportDir string
)

func init() {
	flag.StringVar(&reportDir, "report-dir", "", "Set report directory for artifact output")
}

// Reporters creates the ginkgo reporters for the test suites
func getReporters() []Reporter {
	now, _ := time.Now().MarshalText()
	reps := []Reporter{NewlineReporter{}}
	if reportDir != "" {
		reps = append(reps, reporters.NewJUnitReporter(fmt.Sprintf("%s/junit_%s_%d.xml", reportDir, string(now), config.GinkgoConfig.ParallelNode)))
	}
	return reps
}

func TestMain(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "Promoter Suite", getReporters())
}

// Print a newline after the default Reporter output so that the results are correctly parsed
// by test automation.
// See issue https://github.com/jstemmer/go-junit-report/issues/31
type NewlineReporter struct{}

func (NewlineReporter) SpecSuiteWillBegin(cfg config.GinkgoConfigType, summary *SuiteSummary) {}

func (NewlineReporter) BeforeSuiteDidRun(setupSummary *SetupSummary) {}

func (NewlineReporter) AfterSuiteDidRun(setupSummary *SetupSummary) {}

func (NewlineReporter) SpecWillRun(specSummary *SpecSummary) {}

func (NewlineReporter) SpecDidComplete(specSummary *SpecSummary) {}

// SpecSuiteDidEnd Prints a newline between "35 Passed | 0 Failed | 0 Pending | 0 Skipped" and "--- PASS:"
func (NewlineReporter) SpecSuiteDidEnd(summary *SuiteSummary) { fmt.Printf("\n") }
