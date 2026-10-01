package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bitrise-io/go-steputils/v2/testattachment"
	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/env"
	"github.com/bitrise-io/go-utils/v2/fileutil"
	"github.com/bitrise-io/go-utils/v2/log"
)

const loginJUnitXML = `<testsuites>
  <testsuite name="LoginTest">
    <testcase classname="com.example.LoginTest" name="emptyState"/>
    <testcase classname="com.example.LoginTest" name="wrongPassword">
      <properties>
        <property name="attachment_0" value="shots/com.example.LoginTest__wrongPassword__1.png"/>
      </properties>
    </testcase>
  </testsuite>
</testsuites>`

func Test_exportConventionAttachments(t *testing.T) {
	root := t.TempDir()
	junitPath := filepath.Join(root, "results", "junit.xml")
	writeTestFile(t, junitPath, loginJUnitXML)
	writeTestFile(t, filepath.Join(root, "build", "shots", "com.example.LoginTest__emptyState__1.png"), "screenshot")
	reportDir := newReportDir(t)

	exportConventionAttachments(log.NewLogger(), newTestCollector(), root, t.TempDir(), junitPath, reportDir, nil)

	assertTestFile(t, filepath.Join(reportDir, "com.example.LoginTest__emptyState__1.png"), "screenshot")
}

func Test_exportConventionAttachments_skipsReferencedFiles(t *testing.T) {
	root := t.TempDir()
	junitPath := filepath.Join(root, "junit.xml")
	writeTestFile(t, junitPath, loginJUnitXML)
	writeTestFile(t, filepath.Join(root, "shots", "com.example.LoginTest__wrongPassword__1.png"), "screenshot")
	reportDir := newReportDir(t)

	exportConventionAttachments(log.NewLogger(), newTestCollector(), root, t.TempDir(), junitPath, reportDir, []string{"shots/com.example.LoginTest__wrongPassword__1.png"})

	if _, err := os.Stat(filepath.Join(reportDir, "com.example.LoginTest__wrongPassword__1.png")); !os.IsNotExist(err) {
		t.Errorf("referenced attachment should not be copied again, got %v", err)
	}
}

func Test_exportConventionAttachments_invalidXML(t *testing.T) {
	root := t.TempDir()
	junitPath := filepath.Join(root, "junit.xml")
	writeTestFile(t, junitPath, "not xml")
	writeTestFile(t, filepath.Join(root, "com.example.LoginTest__emptyState__1.png"), "screenshot")
	reportDir := newReportDir(t)
	var logs bytes.Buffer

	exportConventionAttachments(log.NewLogger(log.WithOutput(&logs)), newTestCollector(), root, t.TempDir(), junitPath, reportDir, nil)

	if !strings.Contains(logs.String(), "Failed to read test cases") {
		t.Errorf("expected a warning about the XML, got logs: %s", logs.String())
	}
	if _, err := os.Stat(filepath.Join(reportDir, "com.example.LoginTest__emptyState__1.png")); !os.IsNotExist(err) {
		t.Errorf("nothing should be exported, got %v", err)
	}
}

func Test_exportConventionAttachments_missingDeployDir(t *testing.T) {
	root := t.TempDir()
	junitPath := filepath.Join(root, "junit.xml")
	writeTestFile(t, junitPath, loginJUnitXML)
	writeTestFile(t, filepath.Join(root, "com.example.LoginTest__emptyState__1.png"), "screenshot")
	reportDir := newReportDir(t)
	var logs bytes.Buffer

	exportConventionAttachments(log.NewLogger(log.WithOutput(&logs)), newTestCollector(), root, "", junitPath, reportDir, nil)

	if !strings.Contains(logs.String(), "BITRISE_TEST_DEPLOY_DIR is not set") {
		t.Errorf("expected a warning about the deploy dir, got logs: %s", logs.String())
	}
	assertTestFile(t, filepath.Join(reportDir, "com.example.LoginTest__emptyState__1.png"), "screenshot")
}

func Test_attachmentRoot(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "junit.xml")
	writeTestFile(t, file, loginJUnitXML)

	if got := attachmentRoot(file); got != dir {
		t.Errorf("attachmentRoot(file) = %s, want %s", got, dir)
	}
	if got := attachmentRoot(dir); got != dir {
		t.Errorf("attachmentRoot(dir) = %s, want %s", got, dir)
	}
}

func newReportDir(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "tests")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func newTestCollector() testattachment.Collector {
	return testattachment.NewCollector(command.NewFactory(env.NewRepository()), fileutil.NewFileManager())
}
