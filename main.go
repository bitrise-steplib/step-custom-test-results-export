package main

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/go-steputils/v2/stepconf"
	"github.com/bitrise-io/go-steputils/v2/testattachment"
	"github.com/bitrise-io/go-steputils/v2/testresultexport" //nolint:staticcheck // deprecated, but kept for this step's migration
	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/env"
	"github.com/bitrise-io/go-utils/v2/fileutil"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/ryanuber/go-glob"
)

func failf(logger log.Logger, format string, args ...interface{}) {
	logger.Errorf(format, args...)
	os.Exit(1)
}

func main() {
	logger := log.NewLogger()
	fileManager := fileutil.NewFileManager()
	envRepo := env.NewRepository()

	var stepConf config
	if err := stepconf.NewInputParser(envRepo).Parse(&stepConf); err != nil {
		failf(logger, "Issue with input: %s", err)
	}
	stepconf.Print(stepConf)

	logger.EnableDebugLog(stepConf.VerboseLog)

	fmt.Println()
	logger.Infof("Searching for test results")

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
		failf(logger, "Invalid base path %s: %s", stepConf.BasePath, err)
	}

	if len(matches) < 1 {
		failf(logger, "Provided search pattern (%s) did not match any files within %s", stepConf.SearchPattern, stepConf.BasePath)
	}

	if stepConf.ExportAllMatches && len(matches) > 1 {
		if results := junitResults(logger, matches); len(results) > 0 {
			exportJUnitResults(logger, fileManager, envRepo, stepConf, basePath, results)
			return
		}
	}

	if len(matches) > 1 {
		logger.Warnf("%s", multipleMatchesWarning(matches))
	}

	match := matches[0]

	logger.Donef("Exporting test result: %s", match)

	exporter := testresultexport.NewExporter(stepConf.TestResultsDir, fileManager)
	if err := exporter.ExportTest(stepConf.TestName, match); err != nil {
		failf(logger, "Failed to export test result: %s", err)
	}

	if strings.HasSuffix(strings.ToLower(match), ".xml") {
		attachments, err := examineAttachmentsInJUnitXML(logger, match)
		if err != nil {
			failf(logger, "Failed to examine attachments in JUnit XML: %s", err)
		}

		if len(attachments) > 0 {
			logger.Donef("Exporting %d attachments found in JUnit XML.", len(attachments))
			junitDir := filepath.Dir(match)
			if err := exportAttachmentsFromJUnitXML(logger, fileManager, attachments, junitDir, stepConf); err != nil {
				failf(logger, "Failed to export attachments from JUnit XML: %s", err)
			}
		}

		collector := testattachment.NewCollector(command.NewFactory(envRepo), fileManager)
		reportDir := filepath.Join(stepConf.TestResultsDir, stepConf.TestName)
		exportConventionAttachments(logger, collector, attachmentRoot(basePath), envRepo.Get("BITRISE_TEST_DEPLOY_DIR"), []string{match}, reportDir, referencedPaths(match, attachments))
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

func examineAttachmentsInJUnitXML(logger log.Logger, junitXmlPath string) ([]string, error) {
	f, err := os.Open(junitXmlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open JUnit XML: %w", err)
	}
	defer f.Close()

	var root xmlNode
	if err := xml.NewDecoder(f).Decode(&root); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	var attachments []string
	seen := make(map[string]bool)
	collectAttachments(logger, &root, &attachments, seen)
	return attachments, nil
}

type xmlNode struct {
	XMLName  xml.Name
	Attrs    []xml.Attr `xml:",any,attr"`
	Children []xmlNode  `xml:",any"`
	Content  string     `xml:",chardata"`
}

func collectAttachments(logger log.Logger, n *xmlNode, attachments *[]string, seen map[string]bool) {
	if n.XMLName.Local == "property" {
		var name, value string
		for _, attr := range n.Attrs {
			switch attr.Name.Local {
			case "name":
				name = attr.Value
			case "value":
				value = attr.Value
			}
		}
		if strings.HasPrefix(name, "attachment_") && value != "" && !filepath.IsAbs(value) && !seen[value] {
			seen[value] = true
			*attachments = append(*attachments, value)
		} else if filepath.IsAbs(value) {
			logger.Warnf("Skipping absolute path attachment for security reasons: %s", value)
		}
	}

	for i := range n.Children {
		collectAttachments(logger, &n.Children[i], attachments, seen)
	}
}

func exportAttachmentsFromJUnitXML(logger log.Logger, fileManager fileutil.FileManager, files []string, junitDir string, stepConf config) error {
	for _, file := range files {
		srcPath := filepath.Join(junitDir, file)

		if _, err := os.Stat(srcPath); err != nil {
			logger.Warnf("Attachment file not found, skipping: %s", srcPath)
			continue
		}

		dstPath := filepath.Join(stepConf.TestResultsDir, stepConf.TestName, file)
		dstDir := filepath.Dir(dstPath)

		logger.Debugf("Exporting attachment from %s to %s", srcPath, dstPath)

		if err := os.MkdirAll(dstDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dstDir, err)
		}
		if err := fileManager.CopyFile(srcPath, dstPath, &fileutil.CopyOptions{Overwrite: true}); err != nil {
			return fmt.Errorf("failed to copy attachment %s: %w", file, err)
		}
	}

	return nil
}
