package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/go-android/v2/testresult/junitxml"
	"github.com/bitrise-io/go-steputils/v2/testattachment"
	"github.com/bitrise-io/go-steputils/v2/testreport"
	"github.com/bitrise-io/go-utils/v2/log"
)

// exportConventionAttachments copies the files under root that are named after a test case of the
// JUnit XMLs at junitPaths into reportDir, where the deploy step links them to their test case.
// The XMLs are indexed together, the same way the deploy step reads a report folder. Problems are
// logged as warnings: the test results themselves are already exported.
func exportConventionAttachments(logger log.Logger, collector testattachment.Collector, root, deployDir string, junitPaths []string, reportDir string, referenced []string) {
	report, err := readJUnitReport(junitPaths...)
	if err != nil {
		logger.Warnf("Failed to read test cases, attachments are not matched by file name: %s", err)
		return
	}

	if deployDir == "" {
		logger.Warnf("BITRISE_TEST_DEPLOY_DIR is not set, attachments exported by earlier steps are not filtered out")
	}
	result, err := collector.Collect(root, deployDir, testattachment.NewIndex(&report))
	if err != nil {
		logger.Warnf("Failed to collect attachments matched by file name: %s", err)
		return
	}
	if result.GitCheckErr != nil {
		logger.Warnf("Failed to check which files are tracked by git, committed files are not filtered out: %s", result.GitCheckErr)
	}
	logSkippedAttachments(logger, result.Skipped)

	candidates := withoutReferenced(result.Candidates, referenced)
	if len(candidates) == 0 {
		return
	}

	logger.Donef("Exporting %d attachments matched by file name.", len(candidates))
	for _, failed := range collector.CopyToReport(reportDir, candidates) {
		logger.Warnf("Failed to export attachment %s: %s", failed.Path, failed.Reason)
	}
}

func attachmentRoot(basePath string) string {
	root := basePath
	if info, err := os.Stat(basePath); err == nil && !info.IsDir() {
		root = filepath.Dir(basePath)
	}
	return absPath(root)
}

func readJUnitReport(paths ...string) (testreport.TestReport, error) {
	var converter junitxml.Converter
	if !converter.Detect(paths) {
		return testreport.TestReport{}, fmt.Errorf("%s is not a JUnit XML", strings.Join(paths, ", "))
	}
	return converter.Convert()
}

// Files of other test reports under the same root are expected, so those reasons are only debug logs.
func logSkippedAttachments(logger log.Logger, skipped []testattachment.Skipped) {
	for _, s := range skipped {
		switch {
		case errors.Is(s.Reason, testattachment.ErrDuplicateName),
			errors.Is(s.Reason, testattachment.ErrAmbiguousTest),
			errors.Is(s.Reason, testattachment.ErrUnknownRun),
			errors.Is(s.Reason, testattachment.ErrMissingLabel):
			logger.Warnf("Skipping attachment %s: %s", s.Path, s.Reason)
		default:
			logger.Debugf("Skipping attachment %s: %s", s.Path, s.Reason)
		}
	}
}

// referencedPaths resolves the attachment_* values of the JUnit XML at junitPath, which are
// relative to the XML's folder.
func referencedPaths(junitPath string, values []string) []string {
	paths := make([]string, 0, len(values))
	for _, value := range values {
		paths = append(paths, filepath.Join(filepath.Dir(junitPath), value))
	}
	return paths
}

// withoutReferenced leaves out the files that an XML already references with an attachment_*
// property. They are exported under their relative path, and a second copy in the report folder
// would be linked again by the deploy step.
func withoutReferenced(candidates []testattachment.Candidate, referenced []string) []testattachment.Candidate {
	if len(referenced) == 0 {
		return candidates
	}
	paths := map[string]bool{}
	for _, path := range referenced {
		paths[absPath(path)] = true
	}

	var kept []testattachment.Candidate
	for _, candidate := range candidates {
		if !paths[absPath(candidate.Path)] {
			kept = append(kept, candidate)
		}
	}
	return kept
}

func absPath(path string) string {
	if abs, err := filepath.Abs(path); err == nil {
		return abs
	}
	return filepath.Clean(path)
}
