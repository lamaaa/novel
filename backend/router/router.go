package router

import (
	"github.com/gin-gonic/gin"
	"novel-service/handlers"
)

func Setup(r *gin.Engine) {
	api := r.Group("/api")
	{
		// 小说（只读）
		api.GET("/novels", handlers.GetNovels)
		api.GET("/novels/:id", handlers.GetNovel)

		// 章节（只读）
		api.GET("/novels/:id/chapters", handlers.GetChapters)
		api.GET("/chapters/:id", handlers.GetChapter)

		// 人物（只读）
		api.GET("/novels/:id/characters", handlers.GetCharacters)
		api.GET("/characters/:id", handlers.GetCharacter)

		// 世界观（只读）
		api.GET("/novels/:id/worldviews", handlers.GetWorldviews)
		api.GET("/worldviews/:id", handlers.GetWorldview)

		// 伏笔（只读）
		api.GET("/novels/:id/foreshadowings", handlers.GetForeshadowings)
		api.GET("/foreshadowings/:id", handlers.GetForeshadowing)

		// 长期记忆（只读）
		api.GET("/novels/:id/memory", handlers.GetNovelMemory)
		api.GET("/novels/:id/memory/search", handlers.SearchNovelMemory)

		// 版本管理（只读）
		api.GET("/versions/:entityType/:entityId", handlers.GetVersions)
		api.GET("/versions/detail/:id", handlers.GetVersion)
	}
}
