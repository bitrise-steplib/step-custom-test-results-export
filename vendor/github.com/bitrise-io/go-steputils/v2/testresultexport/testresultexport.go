// Package testresultexport exports a step's test result file or directory to
// $BITRISE_TEST_DEPLOY_DIR with a sidecar test-info.json.
//
// Deprecated: The v2 testreport package is the preferred abstraction for
// new callers. This package exists to keep two legacy consumers
// (bitrise-step-flutter-test, step-custom-test-results-export) working
// against go-steputils/v2 without a larger rewrite. New steps should
// produce a testreport.TestReport directly.
package testresultexport

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/go-utils/v2/fileutil"
)

// ResultDescriptorFileName is the name of the test result descriptor file
// written next to the copied test result file or directory.
const ResultDescriptorFileName = "test-info.json"

// TestInfo is the payload serialized into test-info.json.
type TestInfo struct {
	Name string `json:"test-name" yaml:"test-name"`
}

// Exporter writes test result directories into a deploy-dir layout that
// Bitrise can pick up. Callers that want to mock test-result export should
// define a minimal interface at the call site (typically one method:
// ExportTest(name, testResultPath string) error) and depend on that
// instead of this concrete type.
type Exporter struct {
	exportPath  string
	fileManager fileutil.FileManager
}

// NewExporter returns an Exporter that writes under exportPath using the
// given FileManager for file operations.
func NewExporter(exportPath string, fileManager fileutil.FileManager) *Exporter {
	return &Exporter{exportPath: exportPath, fileManager: fileManager}
}

// ExportTest copies the test result at testResultPath, a file (e.g. a JUnit XML) or a directory
// (e.g. an .xcresult bundle), into <exportPath>/<name>/ under its own name, and writes a sidecar
// test-info.json describing it. A directory path ending in a separator is exported by its
// contents instead. An earlier export under the same name is overwritten.
func (e *Exporter) ExportTest(name, testResultPath string) error {
	exportDir := filepath.Join(e.exportPath, name)

	if err := os.MkdirAll(exportDir, os.ModePerm); err != nil {
		return fmt.Errorf("skipping test result (%s): ensure export dir (%s): %w", testResultPath, exportDir, err)
	}

	infoBytes, err := json.Marshal(&TestInfo{Name: name})
	if err != nil {
		return fmt.Errorf("marshal test info: %w", err)
	}
	infoPath := filepath.Join(exportDir, ResultDescriptorFileName)
	if err := e.fileManager.WriteBytes(infoPath, infoBytes); err != nil {
		return fmt.Errorf("write %s: %w", infoPath, err)
	}

	info, err := os.Stat(testResultPath)
	if err != nil {
		return fmt.Errorf("skipping test result (%s): %w", testResultPath, err)
	}
	// The v1 exporter ran `rsync -ar <testResultPath> <exportDir>`, and consumers rely on both of its
	// cases: without a trailing slash it copies the path itself, so a .xcresult stays a bundle and is
	// detected; with one it copies the contents, so the JUnit XMLs of a results folder land where the
	// deploy step looks for them, one level deep.
	dst := filepath.Join(exportDir, filepath.Base(testResultPath))
	opts := &fileutil.CopyOptions{Overwrite: true}
	if info.IsDir() && strings.HasSuffix(testResultPath, string(filepath.Separator)) {
		dst = exportDir
	}
	if info.IsDir() {
		return e.fileManager.CopyDir(testResultPath, dst, opts)
	}
	return e.fileManager.CopyFile(testResultPath, dst, opts)
}
