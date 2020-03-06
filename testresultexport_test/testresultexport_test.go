package testresultexport_test

import (
	"os"
	"testing"

	"github.com/bitrise-steplib/step-custom-test-results-export/testresultexport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockFuncs struct {
	mock.Mock
}

func (m *MockFuncs) MkdirAll(path string, perm os.FileMode) error {
	args := m.Called(path, perm)
	return args.Error(0)
}

func (m *MockFuncs) Copy(src string, dst string) error {
	args := m.Called(src, dst)
	return args.Error(0)
}

func (m *MockFuncs) GenerateTestInfoFile(dir string, data []byte) error {
	args := m.Called(dir, data)
	return args.Error(0)
}

const testJSON = `{
	"test_name":"test_name",
}`

func TestExportTestWritesTestResults(t *testing.T) {
	// Arrange
	testName := "testName"
	testResultPath := "testResultPath/Result.xml"
	testExportPath := "testExportPath"
	expectedTestFolder := "testExportPath/testName"
	expectedCopyPath := "testExportPath/testName/Result.xml"

	mockFuncs := new(MockFuncs)
	mockFuncs.On("MkdirAll", expectedTestFolder, mock.Anything).Return(nil)
	mockFuncs.On("Copy", testResultPath, expectedCopyPath).Return(nil)
	mockFuncs.On("GenerateTestInfoFile", expectedTestFolder, mock.Anything).Return(nil)

	testSubject := testresultexport.Exporter{
		ExportPath:           testExportPath,
		Copy:                 mockFuncs.Copy,
		GenerateTestInfoFile: mockFuncs.GenerateTestInfoFile,
		MkdirAll:             mockFuncs.MkdirAll,
	}

	// Act
	err := testSubject.ExportTest(testName, testResultPath)

	// Assert
	assert.Nil(t, err, "error should be nil")
	mockFuncs.AssertExpectations(t)
}
