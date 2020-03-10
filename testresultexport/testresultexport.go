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
	Name string `json:"test-name" yaml:"test-name"` // Test name
}

// Exporter is an implementation of the ExporterInterface
type Exporter struct {
	exportPath string

	mkdirAll             func(path string, perm os.FileMode) error
	generateTestInfoFile func(dir string, data *TestInfo) error
	copy                 func(src, dst string) error
}

// NewExporter instantiates a new exporter
func NewExporter(exportPath string) *Exporter {
	e := Exporter{
		exportPath: exportPath,
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

// SetMkdirAll sets mkdirAll
func (e *Exporter) SetMkdirAll(mkdirAll func(path string, perm os.FileMode) error) {
	e.mkdirAll = mkdirAll
}

// SetGenerateTestInfoFile sets generateTestInfoFile
func (e *Exporter) SetGenerateTestInfoFile(generateTestInfoFile func(dir string, data *TestInfo) error) {
	e.generateTestInfoFile = generateTestInfoFile
}

// SetCopy sets generateTestInfoFile
func (e *Exporter) SetCopy(copy func(src, dst string) error) {
	e.copy = copy
}

// ExportTest exports a test result with a given name
func (e *Exporter) ExportTest(name string, testResultPath string) error {
	testInfo := &TestInfo{
		Name: name,
	}

	exportDir := filepath.Join(e.exportPath, name)

	if err := e.mkdirAll(exportDir, os.ModePerm); err != nil {
		return fmt.Errorf("skipping test result (%s): could not ensure unique export dir (%s): %s", testResultPath, exportDir, err)
	}

	if err := e.generateTestInfoFile(exportDir, testInfo); err != nil {
		return err
	}

	return e.copy(testResultPath, exportDir)
}
