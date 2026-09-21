package router

import (
	"devops-platform/internal/config"
	"devops-platform/internal/handler"
	"devops-platform/internal/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func New(db *gorm.DB, cfg config.Config) *gin.Engine {
	r := gin.Default()
	r.Use(cors())

	authH := handler.NewAuthHandler(db, cfg)
	dashH := handler.NewDashboardHandler(db)
	projectH := handler.NewProjectHandler(db)
	gitH := handler.NewGitHandler()
	reqH := handler.NewRequirementHandler(db)
	iterH := handler.NewIterationHandler(db)
	issueH := handler.NewIssueHandler(db)
	releaseH := handler.NewReleaseHandler(db)
	commentH := handler.NewCommentHandler(db)
	activityH := handler.NewActivityHandler(db)
	memberH := handler.NewMemberHandler(db)
	userH := handler.NewUserHandler(db)
	attachH := handler.NewAttachmentHandler(db, cfg)
	buildH := handler.NewBuildHandler(db, cfg)
	releaseAssetH := handler.NewReleaseAssetHandler(db, cfg)
	projectStoreAssetH := handler.NewProjectStoreAssetHandler(db, cfg)
	configFileH := handler.NewConfigFileHandler(db, cfg)
	progressH := handler.NewProgressHandler(db)

	r.Static("/uploads", cfg.UploadDir)

	api := r.Group("/api")
	{
		api.POST("/auth/login", authH.Login)
		// 微信推送免登录下载
		api.GET("/builds/share/:token/download", buildH.DownloadByShareToken)
	}

	auth := api.Group("")
	auth.Use(middleware.Auth(cfg))
	{
		auth.GET("/auth/me", authH.Me)
		auth.GET("/dashboard/summary", dashH.Summary)
		auth.GET("/users", userH.List)

		auth.GET("/projects", projectH.List)
		auth.GET("/projects/:id", projectH.Get)
		auth.POST("/projects", projectH.Create)
		auth.PUT("/projects/:id", projectH.Update)
		auth.GET("/fs/browse", projectH.BrowseDirs)
		auth.GET("/git/branches", gitH.ListBranches)
		auth.GET("/projects/:id/builds", buildH.List)
		auth.POST("/projects/:id/builds", buildH.Trigger)
		auth.GET("/projects/:id/members", memberH.List)
		auth.POST("/projects/:id/members", memberH.Add)
		auth.PUT("/projects/:id/members/:memberId", memberH.Update)
		auth.DELETE("/projects/:id/members/:memberId", memberH.Remove)
		auth.GET("/projects/:id/settings", projectH.GetSettings)
		auth.PUT("/projects/:id/settings", projectH.SaveSettings)
		auth.POST("/projects/:id/notify/test", projectH.TestNotify)
		auth.GET("/projects/:id/config-files", configFileH.List)
		auth.GET("/projects/:id/config-files/:category/content", configFileH.GetContent)
		auth.POST("/projects/:id/config-files/:category/upload", configFileH.Upload)
		auth.POST("/projects/:id/config-files/:category/paste", configFileH.Paste)
		auth.GET("/projects/:id/config-files/:category/preview", configFileH.Preview)
		auth.GET("/projects/:id/config-files/i18n/diff", configFileH.I18nDiff)
		auth.GET("/projects/:id/config-files/sms/default-template", configFileH.DefaultSmsTemplate)
		auth.GET("/projects/:id/store-assets", projectStoreAssetH.List)
		auth.GET("/projects/:id/store-assets/download", projectStoreAssetH.DownloadZip)
		auth.GET("/projects/:id/store-assets/:assetId/download", projectStoreAssetH.DownloadOne)
		auth.POST("/projects/:id/store-assets", projectStoreAssetH.Upload)
		auth.DELETE("/projects/:id/store-assets/:assetId", projectStoreAssetH.Delete)

		auth.GET("/projects/:id/progress", progressH.Get)
		auth.PATCH("/projects/:id/checklist/:itemId", progressH.PatchChecklist)
		auth.POST("/projects/:id/progress/apply-hints", progressH.ApplyHints)

		auth.GET("/requirements", reqH.List)
		auth.GET("/requirements/:id", reqH.Get)
		auth.POST("/requirements", reqH.Create)
		auth.PUT("/requirements/:id", reqH.Update)

		auth.GET("/iterations", iterH.List)
		auth.GET("/iterations/:id", iterH.Get)
		auth.GET("/iterations/:id/kanban", iterH.Kanban)
		auth.POST("/iterations", iterH.Create)
		auth.PUT("/iterations/:id", iterH.Update)

		auth.GET("/issues", issueH.List)
		auth.GET("/issues/:id", issueH.Get)
		auth.POST("/issues", issueH.Create)
		auth.PUT("/issues/:id", issueH.Update)
		auth.PATCH("/issues/:id/status", issueH.PatchStatus)

		auth.GET("/releases", releaseH.List)
		auth.GET("/releases/:id", releaseH.Get)
		auth.POST("/releases", releaseH.Create)
		auth.PUT("/releases/:id", releaseH.Update)
		auth.POST("/releases/:id/submit", releaseH.Submit)
		auth.POST("/releases/:id/approve", releaseH.Approve)
		auth.POST("/releases/:id/reject", releaseH.Reject)
		auth.POST("/releases/:id/publish", releaseH.Publish)
		auth.GET("/releases/:id/approvals", releaseH.Approvals)
		auth.GET("/releases/:id/store-assets", releaseAssetH.List)
		auth.POST("/releases/:id/store-assets", releaseAssetH.Upload)
		auth.DELETE("/releases/:id/store-assets/:assetId", releaseAssetH.Delete)

		auth.GET("/comments", commentH.List)
		auth.POST("/comments", commentH.Create)
		auth.DELETE("/comments/:id", commentH.Delete)

		auth.GET("/activities", activityH.List)

		auth.GET("/attachments", attachH.List)
		auth.POST("/attachments/upload", attachH.Upload)
		auth.DELETE("/attachments/:id", attachH.Delete)

		auth.GET("/builds", buildH.List)
		auth.GET("/builds/:id", buildH.Get)
		auth.GET("/builds/:id/artifacts", buildH.ListArtifactFiles)
		auth.GET("/builds/:id/download", buildH.DownloadArtifacts)
		auth.GET("/builds/:id/download-file", buildH.DownloadArtifactFile)
		auth.DELETE("/builds/:id/artifacts", buildH.DeleteArtifacts)
	}

	return r
}

func cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
