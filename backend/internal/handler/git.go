package handler

import (
	"devops-platform/internal/pkg/response"
	"devops-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type GitHandler struct{}

func NewGitHandler() *GitHandler {
	return &GitHandler{}
}

func (h *GitHandler) ListBranches(c *gin.Context) {
	gitURL := c.Query("url")
	branches, err := service.ListRemoteBranches(gitURL)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, branches)
}
