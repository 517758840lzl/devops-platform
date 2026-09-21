package handler

import (
	"devops-platform/internal/config"
	"devops-platform/internal/model"
	"devops-platform/internal/pkg/jwtutil"
	"devops-platform/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db  *gorm.DB
	cfg config.Config
}

func NewAuthHandler(db *gorm.DB, cfg config.Config) *AuthHandler {
	return &AuthHandler{db: db, cfg: cfg}
}

type loginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "invalid request")
		return
	}
	var user model.User
	if err := h.db.Where("username = ? AND status = ?", req.Username, "active").First(&user).Error; err != nil {
		response.Fail(c, 401, "invalid username or password")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		response.Fail(c, 401, "invalid username or password")
		return
	}
	token, err := jwtutil.Sign(user.ID, user.Username, user.Role, h.cfg.JWTSecret)
	if err != nil {
		response.Fail(c, 500, "token error")
		return
	}
	response.OK(c, gin.H{
		"token": token,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"name":     user.Name,
			"role":     user.Role,
			"email":    user.Email,
		},
	})
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var user model.User
	if err := h.db.First(&user, userID).Error; err != nil {
		response.Fail(c, 404, "user not found")
		return
	}
	response.OK(c, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"name":     user.Name,
		"role":     user.Role,
		"email":    user.Email,
	})
}
