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
type Exporter struct {
	ExportPath string

	MkdirAll             func(path string, perm os.FileMode) error
	GenerateTestInfoFile func(dir string, data *TestInfo) error
	Copy                 func(src, dst string) error
}

// NewExporter instantiates a new exporter
func NewExporter(exportPath string) *Exporter {
	e := Exporter{
		ExportPath: exportPath,
		MkdirAll:   os.MkdirAll,
		GenerateTestInfoFile: func(dir string, data *TestInfo) error {
			pth := filepath.Join(dir, ResultDescriptorFileName)
			return fileutil.WriteJSONToFile(pth, data)
		},
		Copy: func(src, dst string) error {
			return command.CopyDir(src, dst, false)
		},
	}

	return &e
}

// ExportTest exports a test result with a given name
func (e *Exporter) ExportTest(name string, testResultPath string) error {
	testInfo := &TestInfo{
		Name: name,
	}

	testResultName := filepath.Base(testResultPath)
	exportDir := filepath.Join(e.ExportPath, name)
	testResultExportName := filepath.Join(exportDir, testResultName)

	if err := e.MkdirAll(exportDir, os.ModePerm); err != nil {
		return fmt.Errorf("skipping test result (%s): could not ensure unique export dir (%s): %s", testResultPath, exportDir, err)
	}

	if err := e.GenerateTestInfoFile(exportDir, testInfo); err != nil {
		return err
	}

	return e.Copy(testResultPath, testResultExportName)
}
