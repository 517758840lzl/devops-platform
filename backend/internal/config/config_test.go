package config

import (
	"path/filepath"
	"testing"
)

func TestLoad_dataDirCoversChildren(t *testing.T) {
	t.Setenv("DATA_DIR", "/tmp/zzyd-data")
	t.Setenv("DB_PATH", "")
	t.Setenv("UPLOAD_DIR", "")
	t.Setenv("WORKSPACE_DIR", "")

	cfg := Load()
	if cfg.DataDir != "/tmp/zzyd-data" {
		t.Fatalf("DataDir=%q", cfg.DataDir)
	}
	if cfg.DBPath != filepath.Join("/tmp/zzyd-data", "devops.db") {
		t.Fatalf("DBPath=%q", cfg.DBPath)
	}
	if cfg.UploadDir != filepath.Join("/tmp/zzyd-data", "uploads") {
		t.Fatalf("UploadDir=%q", cfg.UploadDir)
	}
	if cfg.WorkspaceDir != filepath.Join("/tmp/zzyd-data", "workspaces") {
		t.Fatalf("WorkspaceDir=%q", cfg.WorkspaceDir)
	}
	if cfg.ArtifactDir != filepath.Join("/tmp/zzyd-data", "artifacts") {
		t.Fatalf("ArtifactDir=%q", cfg.ArtifactDir)
	}
}

func TestResolveUnderDataDir_legacyRelative(t *testing.T) {
	got := ResolveUnderDataDir("/data", "data/uploads/config/project_2/i18n_source/content.dart")
	want := filepath.Join("/data", "uploads/config/project_2/i18n_source/content.dart")
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestRelDataPath_staysPortable(t *testing.T) {
	got := RelDataPath("/data", "/data/artifacts/MO/build_12")
	if got != filepath.Join("artifacts", "MO", "build_12") {
		t.Fatalf("got %q", got)
	}
}

func TestLoad_explicitOverrideWins(t *testing.T) {
	t.Setenv("DATA_DIR", "/tmp/zzyd-data")
	t.Setenv("UPLOAD_DIR", "/custom/uploads")
	t.Setenv("DB_PATH", "")
	t.Setenv("WORKSPACE_DIR", "")

	cfg := Load()
	if cfg.UploadDir != "/custom/uploads" {
		t.Fatalf("UploadDir=%q, want override", cfg.UploadDir)
	}
	if cfg.DBPath != filepath.Join("/tmp/zzyd-data", "devops.db") {
		t.Fatalf("DBPath=%q", cfg.DBPath)
	}
}
