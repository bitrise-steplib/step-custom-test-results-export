package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/bitrise-io/go-utils/v2/env"
	"github.com/bitrise-io/go-utils/v2/fileutil"
	"github.com/bitrise-io/go-utils/v2/log"
)

const homeJUnitXML = `<testsuite name="HomeTest">
  <testcase classname="com.example.HomeTest" name="feedLoaded"/>
</testsuite>`

func Test_junitResults(t *testing.T) {
	dir := t.TempDir()
	login := filepath.Join(dir, "TEST-com.example.LoginTest.xml")
	home := filepath.Join(dir, "TEST-com.example.HomeTest.xml")
	manifest := filepath.Join(dir, "AndroidManifest.xml")
	binary := filepath.Join(dir, "binary", "output.bin")
	releaseLogin := filepath.Join(dir, "testReleaseUnitTest", "TEST-com.example.LoginTest.xml")
	writeTestFile(t, login, loginJUnitXML)
	writeTestFile(t, releaseLogin, loginJUnitXML)
	writeTestFile(t, home, homeJUnitXML)
	writeTestFile(t, manifest, `<manifest package="com.example"/>`)
	writeTestFile(t, binary, "bin")

	tests := []struct {
		name    string
		matches []string
		want    []string
	}{
		{name: "every match is a JUnit XML", matches: []string{home, login}, want: []string{home, login}},
		{name: "XML that isn't JUnit is skipped", matches: []string{manifest, home, login}, want: []string{home, login}},
		{name: "a file name that is already exported is skipped", matches: []string{home, login, releaseLogin}, want: []string{home, login}},
		{name: "a folder falls back to the first match", matches: []string{filepath.Dir(binary), home, login}, want: nil},
		{name: "another kind of file falls back to the first match", matches: []string{binary, home, login}, want: nil},
		{name: "no JUnit XML falls back to the first match", matches: []string{manifest, manifest}, want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := junitResults(log.NewLogger(), tt.matches); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("junitResults() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_exportJUnitResults(t *testing.T) {
	root := t.TempDir()
	debugLogin := filepath.Join(root, "app", "build", "test-results", "testDebugUnitTest", "TEST-com.example.LoginTest.xml")
	debugHome := filepath.Join(root, "app", "build", "test-results", "testDebugUnitTest", "TEST-com.example.HomeTest.xml")
	writeTestFile(t, debugLogin, loginJUnitXML)
	writeTestFile(t, debugHome, homeJUnitXML)
	writeTestFile(t, filepath.Join(root, "app", "build", "test-results", "testDebugUnitTest", "shots", "com.example.LoginTest__wrongPassword__1.png"), "referenced")
	writeTestFile(t, filepath.Join(root, "app", "build", "outputs", "roborazzi", "com.example.HomeTest__feedLoaded__1.png"), "convention")
	stepConf := config{TestName: "tests", TestResultsDir: t.TempDir()}
	reportDir := filepath.Join(stepConf.TestResultsDir, "tests")

	exportJUnitResults(log.NewLogger(), fileutil.NewFileManager(), env.NewRepository(), stepConf, root, []string{debugHome, debugLogin})

	assertTestFile(t, filepath.Join(reportDir, "TEST-com.example.HomeTest.xml"), homeJUnitXML)
	assertTestFile(t, filepath.Join(reportDir, "TEST-com.example.LoginTest.xml"), loginJUnitXML)
	assertTestFile(t, filepath.Join(reportDir, "shots", "com.example.LoginTest__wrongPassword__1.png"), "referenced")
	assertTestFile(t, filepath.Join(reportDir, "com.example.HomeTest__feedLoaded__1.png"), "convention")
	if _, err := os.Stat(filepath.Join(reportDir, "test-info.json")); err != nil {
		t.Errorf("test-info.json should be written: %s", err)
	}
	if _, err := os.Stat(filepath.Join(reportDir, "com.example.LoginTest__wrongPassword__1.png")); !os.IsNotExist(err) {
		t.Errorf("referenced attachment should not be copied again, got %v", err)
	}
}
