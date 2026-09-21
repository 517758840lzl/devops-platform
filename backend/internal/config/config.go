package config

import "os"

type Config struct {
	Port          string
	JWTSecret     string
	DBPath        string
	UploadDir     string
	WorkspaceDir  string
}

func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "devops-platform-dev-secret-change-me"
	}
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "data/devops.db"
	}
	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "data/uploads"
	}
	workspaceDir := os.Getenv("WORKSPACE_DIR")
	if workspaceDir == "" {
		workspaceDir = "data/workspaces"
	}
	return Config{Port: port, JWTSecret: secret, DBPath: dbPath, UploadDir: uploadDir, WorkspaceDir: workspaceDir}
}
