package middleware

import (
	"strings"

	"devops-platform/internal/config"
	"devops-platform/internal/pkg/jwtutil"
	"devops-platform/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

const UserIDKey = "user_id"
const UsernameKey = "username"
const RoleKey = "role"

func Auth(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			response.Fail(c, 401, "unauthorized")
			c.Abort()
			return
		}
		token := strings.TrimPrefix(header, "Bearer ")
		claims, err := jwtutil.Parse(token, cfg.JWTSecret)
		if err != nil {
			response.Fail(c, 401, "invalid token")
			c.Abort()
			return
		}
		c.Set(UserIDKey, claims.UserID)
		c.Set(UsernameKey, claims.Username)
		c.Set(RoleKey, claims.Role)
		c.Next()
	}
}

func CurrentUserID(c *gin.Context) uint {
	v, _ := c.Get(UserIDKey)
	id, _ := v.(uint)
	return id
}
