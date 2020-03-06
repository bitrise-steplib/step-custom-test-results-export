package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/go-steputils/stepconf"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-steplib/bitrise-step-custom-test-results-export/testresultexport"
	"github.com/ryanuber/go-glob"
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

	matches := []string{}
	basePath := strings.Split(stepConf.BasePath, "*")[0]
	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if glob.Glob(stepConf.Pattern, path) {
			matches = append(matches, path)
		}

		return nil
	})

	if err != nil {
		failf("Invalid base path %s: %s", stepConf.BasePath, err)
	}

	if len(matches) < 1 {
		failf("Pattern %s did not match any files within path %s", stepConf.Pattern, stepConf.BasePath)
	}

	if len(matches) > 1 {
		log.Warnf("Pattern matched more than one file, will use the first match:")
		for i, m := range matches {
			template := fmt.Sprintf("- %s\n", m)
			if i == 0 {
				template = colorstring.Green(template)
			}
			log.Warnf(template)
		}
	}

	match := matches[0]
	exporter := testresultexport.NewExporter(stepConf.TestResultsDir)

	if err := exporter.ExportTest(stepConf.TestName, match); err != nil {
		failf("Failed to export test result: %s", err)
	}
}
