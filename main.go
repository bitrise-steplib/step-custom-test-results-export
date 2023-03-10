package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/go-steputils/stepconf"
	"github.com/bitrise-io/go-steputils/testresultexport"
	"github.com/bitrise-io/go-utils/log"
	"github.com/ryanuber/go-glob"
)

func failf(format string, args ...interface{}) {
	log.Errorf(format, args...)
	os.Exit(1)
}

func main() {
	var stepConf config
	if err := stepconf.Parse(&stepConf); err != nil {
		failf("Issue with input: %s", err)
	}
	stepconf.Print(stepConf)

	log.SetEnableDebugLog(stepConf.VerboseLog)

	log.Infof("Searching for test results")

	var matches []string
	basePath := strings.Split(stepConf.BasePath, "*")[0]
	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if glob.Glob(stepConf.SearchPattern, path) {
			matches = append(matches, path)
		}

		return nil
	})

	if err != nil {
		failf("Invalid base path %s: %s", stepConf.BasePath, err)
	}

	if len(matches) < 1 {
		failf("Provided search pattern (%s) did not match any files within %s", stepConf.SearchPattern, stepConf.BasePath)
	}

	if len(matches) > 1 {
		warnMessage := multipleMatchesWarning(matches)
		log.Warnf(warnMessage)
	}

	match := matches[0]

	log.Donef("Exporting test result: %s", match)

	exporter := testresultexport.NewExporter(stepConf.TestResultsDir)

	if err := exporter.ExportTest(stepConf.TestName, match); err != nil {
		failf("Failed to export test result: %s", err)
	}
}

func multipleMatchesWarning(matches []string) string {
	warnMessage := fmt.Sprintf("Provided search pattern matches %d files:\n", len(matches))
	for i := 0; i < 5; i++ {
		if i == len(matches) {
			break
		}

		warnMessage += fmt.Sprintf("- %s\n", matches[i])
	}
	if len(matches) > 5 {
		warnMessage += "...\n"
	}
	return warnMessage
}
