package testresultexport

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	GenerateTestInfoFile func(dir string, data []byte) error
	Copy                 func(src, dst string) error
}

// NewExporter instantiates a new exporter
func NewExporter(exportPath string) *Exporter {
	e := Exporter{
		ExportPath: exportPath,
		MkdirAll:   os.MkdirAll,
		GenerateTestInfoFile: func(dir string, data []byte) error {
			return generateTestInfoFile(dir, data)
		},
		Copy: func(src, dst string) error {
			return copy(src, dst)
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

	testInfoData, err := json.Marshal(testInfo)
	if err != nil {
		return err
	}

	if err := e.MkdirAll(exportDir, os.ModePerm); err != nil {
		return fmt.Errorf("skipping test result (%s): could not ensure unique export dir (%s): %s", testResultPath, exportDir, err)
	}

	if err := e.GenerateTestInfoFile(exportDir, testInfoData); err != nil {
		return err
	}

	err = e.Copy(testResultPath, testResultExportName)

	return err
}

func copy(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	return err
}

func generateTestInfoFile(dir string, data []byte) error {
	f, err := os.Create(filepath.Join(dir, ResultDescriptorFileName))
	if err != nil {
		return err
	}

	if _, err := f.Write(data); err != nil {
		return err
	}

	if err := f.Sync(); err != nil {
		return err
	}

	if err := f.Close(); err != nil {
		return err
	}

	return nil
}
