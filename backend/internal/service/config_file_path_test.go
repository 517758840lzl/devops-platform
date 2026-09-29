package service

import (
	"os"
	"path/filepath"
	"testing"

	"devops-platform/internal/model"
)

func TestResolveConfigFilePath_legacyCWDRelative(t *testing.T) {
	root := t.TempDir()
	uploadDir := filepath.Join(root, "uploads")
	want := filepath.Join(uploadDir, "config", "project_2", "i18n_source", "content.dart")
	if err := os.MkdirAll(filepath.Dir(want), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(want, []byte("const x = 'ok';"), 0o644); err != nil {
		t.Fatal(err)
	}

	got := ResolveConfigFilePath(uploadDir, &model.ProjectConfigFile{
		ProjectID:    2,
		Category:     "i18n_source",
		FilePath:     "data/uploads/config/project_2/i18n_source/content.dart",
		OriginalName: "app_strings.dart",
	})
	if got != want {
		t.Fatalf("resolved %q, want %q", got, want)
	}
}

func TestReadConfigFileContent_missingIsEmpty(t *testing.T) {
	content, err := ReadConfigFileContent(t.TempDir(), &model.ProjectConfigFile{
		ProjectID: 9,
		Category:  "i18n_source",
		FilePath:  "data/uploads/config/project_9/i18n_source/content.dart",
	})
	if err != nil {
		t.Fatalf("missing file should not error, got %v", err)
	}
	if content != "" {
		t.Fatalf("want empty content, got %q", content)
	}
}
