package handler

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"devops-platform/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type dirEntry struct {
	Name  string `json:"name"`
	Path  string `json:"path"`
	IsDir bool   `json:"is_dir"`
}

func (h *ProjectHandler) BrowseDirs(c *gin.Context) {
	reqPath := strings.TrimSpace(c.Query("path"))
	if reqPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			response.Fail(c, 500, err.Error())
			return
		}
		reqPath = home
	}

	abs, err := filepath.Abs(reqPath)
	if err != nil {
		response.Fail(c, 400, "invalid path")
		return
	}
	info, err := os.Stat(abs)
	if err != nil || !info.IsDir() {
		response.Fail(c, 404, "directory not found")
		return
	}

	entries, err := os.ReadDir(abs)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}

	dirs := make([]dirEntry, 0)
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		dirs = append(dirs, dirEntry{
			Name:  e.Name(),
			Path:  filepath.Join(abs, e.Name()),
			IsDir: true,
		})
	}
	sort.Slice(dirs, func(i, j int) bool { return dirs[i].Name < dirs[j].Name })

	parent := filepath.Dir(abs)
	if parent == abs {
		parent = ""
	}

	response.OK(c, gin.H{
		"current": abs,
		"parent":  parent,
		"entries": dirs,
	})
}
