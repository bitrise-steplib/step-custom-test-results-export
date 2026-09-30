package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/go-utils/v2/fileutil"
	"github.com/bitrise-io/go-utils/v2/log"
)

func Test_multipleMatchesWarning(t *testing.T) {
	tests := []struct {
		name    string
		matches []string
		want    string
	}{
		{
			name:    "Five or less matches",
			matches: []string{"match_1", "match_2", "match_3", "match_4", "match_5"},
			want: `Provided search pattern matches 5 files:
- match_1
- match_2
- match_3
- match_4
- match_5
`,
		},
		{
			name:    "More than five matches",
			matches: []string{"match_1", "match_2", "match_3", "match_4", "match_5", "match_6"},
			want: `Provided search pattern matches 6 files:
- match_1
- match_2
- match_3
- match_4
- match_5
...
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := multipleMatchesWarning(tt.matches); got != tt.want {
				t.Errorf("multipleMatchesWarning() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_exportAttachmentsFromJUnitXML(t *testing.T) {
	junitDir := t.TempDir()
	writeTestFile(t, filepath.Join(junitDir, "shots", "login.png"), "png")
	stepConf := config{TestName: "tests", TestResultsDir: t.TempDir()}

	err := exportAttachmentsFromJUnitXML(log.NewLogger(), fileutil.NewFileManager(), []string{"shots/login.png", "missing.png"}, junitDir, stepConf)
	if err != nil {
		t.Fatalf("export attachments: %s", err)
	}

	assertTestFile(t, filepath.Join(stepConf.TestResultsDir, "tests", "shots", "login.png"), "png")
	if _, err := os.Stat(filepath.Join(stepConf.TestResultsDir, "tests", "missing.png")); !os.IsNotExist(err) {
		t.Errorf("missing attachment should be skipped, got %v", err)
	}
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func assertTestFile(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Errorf("read %s: %s", path, err)
		return
	}
	if string(got) != want {
		t.Errorf("%s = %q, want %q", path, got, want)
	}
}
