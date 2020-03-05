package main

import (
	"os"
	"path/filepath"

	"github.com/bitrise-io/go-steputils/stepconf"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-steplib/bitrise-step-custom-test-results-export/testresultexport"
)

func failf(format string, args ...interface{}) {
	log.Errorf(format, args...)
	os.Exit(1)
}

func main() {
	var stepConf Config
	if err := stepconf.Parse(&stepConf); err != nil {
		failf("Config: %s", err)
	}
	stepconf.Print(stepConf)

	log.SetEnableDebugLog(stepConf.VerboseLog)

	matches, err := filepath.Glob(stepConf.SearchPath)
	if err != nil {
		failf("Invalid search path %s: %s", stepConf.SearchPath, err)
	}

	if len(matches) < 1 {
		failf("Search path %s did not match any files", stepConf.SearchPath)
	}

	if len(matches) > 1 {
		log.Warnf("Search path matched more than one file, will use the first match, matches: %s", matches)
	}

	match := matches[0]
	exporter := testresultexport.NewExporter(stepConf.TestResultsDir)

	if err := exporter.ExportTest(stepConf.TestName, match); err != nil {
		failf("Failed to export test result: %s", err)
	}
}
