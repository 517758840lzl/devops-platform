package handler

import (
	"devops-platform/internal/model"
	"devops-platform/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserHandler struct {
	db *gorm.DB
}

func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{db: db}
}

func (h *UserHandler) List(c *gin.Context) {
	var users []model.User
	if err := h.db.Where("status = ?", "active").Select("id, username, name, email, role").Find(&users).Error; err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, users)
}
