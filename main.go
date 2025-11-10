package main

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/go-steputils/stepconf"
	"github.com/bitrise-io/go-steputils/testresultexport"
	"github.com/bitrise-io/go-utils/command"
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

	fmt.Println()
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

	if strings.HasSuffix(strings.ToLower(match), ".xml") {
		attachments, err := examineAttachmentsInJUnitXML(match)
		if err != nil {
			failf("Failed to examine attachments in JUnit XML: %s", err)
		}

		if len(attachments) > 0 {
			log.Donef("Exporting %d attachments found in JUnit XML.", len(attachments))
			junitDir := filepath.Dir(match)
			if err := exportAttachmentsFromJUnitXML(attachments, junitDir, stepConf); err != nil {
				failf("Failed to export attachments from JUnit XML: %s", err)
			}
		}
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

func examineAttachmentsInJUnitXML(junitXmlPath string) ([]string, error) {
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
	collectAttachments(&root, &attachments, seen)
	return attachments, nil
}

type xmlNode struct {
	XMLName  xml.Name
	Attrs    []xml.Attr `xml:",any,attr"`
	Children []xmlNode  `xml:",any"`
	Content  string     `xml:",chardata"`
}

func collectAttachments(n *xmlNode, attachments *[]string, seen map[string]bool) {
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
			log.Warnf("Skipping absolute path attachment for security reasons: %s", value)
		}
	}

	for i := range n.Children {
		collectAttachments(&n.Children[i], attachments, seen)
	}
}

func exportAttachmentsFromJUnitXML(files []string, junitDir string, stepConf config) error {
	for _, file := range files {
		srcPath := filepath.Join(junitDir, file)

		if _, err := os.Stat(srcPath); err != nil {
			log.Warnf("Attachment file not found, skipping: %s", srcPath)
			continue
		}

		dstPath := filepath.Join(stepConf.TestResultsDir, stepConf.TestName, file)
		dstDir := filepath.Dir(dstPath)

		log.Debugf("Exporting attachment from %s to %s", srcPath, dstPath)

		if err := os.MkdirAll(dstDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dstDir, err)
		}
		if err := command.CopyFile(srcPath, dstPath); err != nil {
			return fmt.Errorf("failed to copy attachment %s: %w", file, err)
		}
	}

	return nil
}
