package testresultexport

import "os"

// NewExporterMocked instantiates a new exporter for mocking
func NewExporterMocked(exportPath string,
	mkdirAll func(path string, perm os.FileMode) error,
	generateTestInfoFile func(dir string, data *TestInfo) error,
	copy func(src, dst string) error) ExporterInterface {
	e := exporter{
		ExportPath:           exportPath,
		mkdirAll:             mkdirAll,
		generateTestInfoFile: generateTestInfoFile,
		copy:                 copy,
	}

	return &e
}
