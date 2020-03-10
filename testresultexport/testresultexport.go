package testresultexport

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/fileutil"
)

const (
	// ResultDescriptorFileName is the name of the test result descriptor file.
	ResultDescriptorFileName = "test-info.json"
)

// TestInfo ...
type TestInfo struct {
	Name string `json:"test_name" yaml:"test_name"` // Test name
}

// ExporterInterface is an interface for exporting a test result
type ExporterInterface interface {
	ExportTest(name string, testResultPath string) error
}

// Exporter is an implementation of the ExporterInterface
type exporter struct {
	ExportPath string

	mkdirAll             func(path string, perm os.FileMode) error
	generateTestInfoFile func(dir string, data *TestInfo) error
	copy                 func(src, dst string) error
}

// NewExporter instantiates a new exporter
func NewExporter(exportPath string) ExporterInterface {
	e := exporter{
		ExportPath: exportPath,
		mkdirAll:   os.MkdirAll,
		generateTestInfoFile: func(dir string, data *TestInfo) error {
			pth := filepath.Join(dir, ResultDescriptorFileName)
			return fileutil.WriteJSONToFile(pth, data)
		},
		copy: func(src, dst string) error {
			return command.CopyDir(src, dst, false)
		},
	}

	return &e
}

// ExportTest exports a test result with a given name
func (e *exporter) ExportTest(name string, testResultPath string) error {
	testInfo := &TestInfo{
		Name: name,
	}

	exportDir := filepath.Join(e.ExportPath, name)

	if err := e.mkdirAll(exportDir, os.ModePerm); err != nil {
		return fmt.Errorf("skipping test result (%s): could not ensure unique export dir (%s): %s", testResultPath, exportDir, err)
	}

	if err := e.generateTestInfoFile(exportDir, testInfo); err != nil {
		return err
	}

	return e.copy(testResultPath, exportDir)
}
