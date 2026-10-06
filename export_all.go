package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/go-steputils/v2/testattachment"
	"github.com/bitrise-io/go-steputils/v2/testresultexport" //nolint:staticcheck // deprecated, but kept for this step's migration
	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/env"
	"github.com/bitrise-io/go-utils/v2/fileutil"
	"github.com/bitrise-io/go-utils/v2/log"
)

// junitResults returns the matches that are JUnit XMLs, if every match is an XML file. It returns
// nil when the Step should export only the first match instead: when a match is a folder or another
// kind of file, or when none of them is a JUnit XML.
func junitResults(logger log.Logger, matches []string) []string {
	for _, match := range matches {
		if info, err := os.Stat(match); err != nil || info.IsDir() || !strings.HasSuffix(strings.ToLower(match), ".xml") {
			logger.Warnf("Not every match is a JUnit XML file (%s), so only the first match is exported. To export every match, make the search pattern match JUnit XML files only, for example TEST-*.xml.", match)
			return nil
		}
	}

	var results []string
	for _, match := range matches {
		// A file that isn't a JUnit XML would make Deploy to Bitrise.io drop every report of the build.
		if _, err := readJUnitReport(match); err != nil {
			logger.Warnf("Skipping %s: not a JUnit XML", match)
			logger.Debugf("%s", err)
			continue
		}
		results = append(results, match)
	}
	if len(results) == 0 {
		logger.Warnf("None of the matches is a JUnit XML, so only the first match is exported.")
	}
	return results
}

// exportJUnitResults exports every JUnit XML into the same report folder, so they show up as one
// report in the Test Reports, together with their attachments.
func exportJUnitResults(logger log.Logger, fileManager fileutil.FileManager, envRepo env.Repository, stepConf config, basePath string, results []string) {
	logger.Donef("Exporting %d test results:", len(results))
	reportDir := filepath.Join(stepConf.TestResultsDir, stepConf.TestName)
	root := attachmentRoot(basePath)

	exporter := testresultexport.NewExporter(stepConf.TestResultsDir, fileManager)
	exportedNames := map[string]bool{}
	var referenced []string
	for _, result := range results {
		name := filepath.Base(result)
		if exportedNames[name] {
			name = flattenedPath(root, result)
		}
		exportedNames[name] = true
		logger.Printf("- %s => %s", result, name)

		if name == filepath.Base(result) {
			if err := exporter.ExportTest(stepConf.TestName, result); err != nil {
				failf(logger, "Failed to export test result: %s", err)
			}
		} else if err := fileManager.CopyFile(result, filepath.Join(reportDir, name), &fileutil.CopyOptions{Overwrite: true}); err != nil {
			failf(logger, "Failed to export test result: %s", err)
		}

		attachments, err := examineAttachmentsInJUnitXML(logger, result)
		if err != nil {
			failf(logger, "Failed to examine attachments in JUnit XML: %s", err)
		}
		if len(attachments) > 0 {
			logger.Donef("Exporting %d attachments found in %s.", len(attachments), result)
			if err := exportAttachmentsFromJUnitXML(logger, fileManager, attachments, filepath.Dir(result), stepConf); err != nil {
				failf(logger, "Failed to export attachments from JUnit XML: %s", err)
			}
		}
		referenced = append(referenced, referencedPaths(result, attachments)...)
	}

	collector := testattachment.NewCollector(command.NewFactory(envRepo), fileManager)
	exportConventionAttachments(logger, collector, root, envRepo.Get("BITRISE_TEST_DEPLOY_DIR"), results, reportDir, referenced)
}

// flattenedPath names a test result after its path under root, for example
// app-build-test-results-testReleaseUnitTest-TEST-LoginTest.xml, when its own name is already taken
// in the report folder.
func flattenedPath(root, path string) string {
	rel, err := filepath.Rel(root, absPath(path))
	if err != nil || strings.HasPrefix(rel, "..") {
		rel = path
	}
	return strings.ReplaceAll(strings.TrimPrefix(filepath.ToSlash(rel), "/"), "/", "-")
}
