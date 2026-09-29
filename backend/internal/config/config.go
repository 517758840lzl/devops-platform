package config

import (
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Port         string
	JWTSecret    string
	DataDir      string
	DBPath       string
	UploadDir    string
	WorkspaceDir string
	ArtifactDir  string
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

// RelDataPath stores a portable key under DataDir. Absolute leftovers stay as-is only if outside DataDir.
func RelDataPath(dataDir, absPath string) string {
	absPath = strings.TrimSpace(absPath)
	if absPath == "" {
		return ""
	}
	rel, err := filepath.Rel(dataDir, absPath)
	if err != nil || rel == "." || strings.HasPrefix(rel, "..") {
		return absPath
	}
	return rel
}

// ResolveUnderDataDir maps a stored key (relative or leftover CWD/absolute) onto DataDir.
func ResolveUnderDataDir(dataDir, stored string) string {
	stored = strings.TrimSpace(stored)
	if stored == "" {
		return ""
	}
	if filepath.IsAbs(stored) {
		if _, err := os.Stat(stored); err == nil {
			return stored
		}
		slash := filepath.ToSlash(stored)
		for _, mark := range []string{"/artifacts/", "/uploads/", "/workspaces/"} {
			if i := strings.Index(slash, mark); i >= 0 {
				return filepath.Join(dataDir, filepath.FromSlash(slash[i+1:]))
			}
		}
		return stored
	}
	stripped := strings.TrimPrefix(filepath.ToSlash(stored), "data/")
	return filepath.Join(dataDir, filepath.FromSlash(stripped))
}

func Load() Config {
	dataDir := envOr("DATA_DIR", "data")
	return Config{
		Port:         envOr("PORT", "8080"),
		JWTSecret:    envOr("JWT_SECRET", "devops-platform-dev-secret-change-me"),
		DataDir:      dataDir,
		DBPath:       envOr("DB_PATH", filepath.Join(dataDir, "devops.db")),
		UploadDir:    envOr("UPLOAD_DIR", filepath.Join(dataDir, "uploads")),
		WorkspaceDir: envOr("WORKSPACE_DIR", filepath.Join(dataDir, "workspaces")),
		ArtifactDir:  envOr("ARTIFACT_DIR", filepath.Join(dataDir, "artifacts")),
	}
}
