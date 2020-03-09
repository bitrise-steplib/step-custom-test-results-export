package testresultexport_test

import (
	"fmt"
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

func (m *MockFuncs) GenerateTestInfoFile(dir string, data *testresultexport.TestInfo) error {
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

func TestExportTestMkdirFails(t *testing.T) {
	// Arrange
	testName := "testName"
	testResultPath := "testResultPath/Result.xml"
	testExportPath := "testExportPath"
	expectedTestFolder := "testExportPath/testName"
	testError := fmt.Errorf("test error")

	mockFuncs := new(MockFuncs)
	mockFuncs.On("MkdirAll", expectedTestFolder, mock.Anything).Return(testError)

	testSubject := testresultexport.Exporter{
		ExportPath:           testExportPath,
		Copy:                 mockFuncs.Copy,
		GenerateTestInfoFile: mockFuncs.GenerateTestInfoFile,
		MkdirAll:             mockFuncs.MkdirAll,
	}

	// Act
	err := testSubject.ExportTest(testName, testResultPath)

	// Assert
	assert.EqualErrorf(t,
		err,
		fmt.Sprintf("skipping test result (%s): could not ensure unique export dir (%s): %s",
			testResultPath,
			expectedTestFolder,
			testError),
		"should throw the expected error",
	)
	mockFuncs.AssertExpectations(t)
}

func TestExportTestGenerateTestInfoFileFails(t *testing.T) {
	// Arrange
	testName := "testName"
	testResultPath := "testResultPath/Result.xml"
	testExportPath := "testExportPath"
	expectedTestFolder := "testExportPath/testName"
	testError := fmt.Errorf("test error")

	mockFuncs := new(MockFuncs)
	mockFuncs.On("MkdirAll", expectedTestFolder, mock.Anything).Return(nil)
	mockFuncs.On("GenerateTestInfoFile", expectedTestFolder, mock.Anything).Return(testError)

	testSubject := testresultexport.Exporter{
		ExportPath:           testExportPath,
		Copy:                 mockFuncs.Copy,
		GenerateTestInfoFile: mockFuncs.GenerateTestInfoFile,
		MkdirAll:             mockFuncs.MkdirAll,
	}

	// Act
	err := testSubject.ExportTest(testName, testResultPath)

	// Assert
	assert.Equal(t, err, testError, "should throw the expected error")
	mockFuncs.AssertExpectations(t)
}
