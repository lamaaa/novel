package router

import (
	"github.com/gin-gonic/gin"
	"novel-service/handlers"
)

func Setup(r *gin.Engine) {
	api := r.Group("/api")
	{
		// 小说
		api.GET("/novels", handlers.GetNovels)
		api.GET("/novels/:id", handlers.GetNovel)
		api.POST("/novels", handlers.CreateNovel)
		api.PUT("/novels/:id", handlers.UpdateNovel)
		api.DELETE("/novels/:id", handlers.DeleteNovel)

		// 章节
		api.GET("/novels/:id/chapters", handlers.GetChapters)
		api.GET("/chapters/:id", handlers.GetChapter)
		api.POST("/novels/:id/chapters", handlers.CreateChapter)
		api.PUT("/chapters/:id", handlers.UpdateChapter)
		api.DELETE("/chapters/:id", handlers.DeleteChapter)

		// 人物
		api.GET("/novels/:id/characters", handlers.GetCharacters)
		api.GET("/characters/:id", handlers.GetCharacter)
		api.POST("/novels/:id/characters", handlers.CreateCharacter)
		api.PUT("/characters/:id", handlers.UpdateCharacter)
		api.DELETE("/characters/:id", handlers.DeleteCharacter)

		// 世界观
		api.GET("/novels/:id/worldviews", handlers.GetWorldviews)
		api.GET("/worldviews/:id", handlers.GetWorldview)
		api.POST("/novels/:id/worldviews", handlers.CreateWorldview)
		api.PUT("/worldviews/:id", handlers.UpdateWorldview)
		api.DELETE("/worldviews/:id", handlers.DeleteWorldview)

		// 伏笔
		api.GET("/novels/:id/foreshadowings", handlers.GetForeshadowings)
		api.GET("/foreshadowings/:id", handlers.GetForeshadowing)
		api.POST("/novels/:id/foreshadowings", handlers.CreateForeshadowing)
		api.PUT("/foreshadowings/:id", handlers.UpdateForeshadowing)
		api.DELETE("/foreshadowings/:id", handlers.DeleteForeshadowing)

		// 版本管理
		api.GET("/versions/:entityType/:entityId", handlers.GetVersions)
		api.GET("/versions/detail/:id", handlers.GetVersion)
		api.POST("/versions/:id/rollback", handlers.RollbackVersion)
	}
}
