package routes

import (
	"Truyen_BE/controllers"
	middlewares "Truyen_BE/middleware"

	"github.com/gin-gonic/gin"
)

func ChapterRoutes(router *gin.Engine) {
	chapterGroup := router.Group("/stories/chapters")
	chapterGroup.Use(middlewares.LoggingMiddleware)
	{
		// ✅ Public – đọc chương
		chapterGroup.GET("/:story_id/:number", controllers.GetChapterByStoryAndNumber)
		chapterGroup.GET("/id/:id", controllers.GetChapterByID)
		chapterGroup.GET("/newest", controllers.GetNewestChapters)
		// ✅ Yêu cầu đăng nhập + phân quyền tác giả của truyện/chương
		chapterGroup.POST("",
			middlewares.AuthMiddleware(),
			middlewares.IsAuthorOfStory(),
			controllers.InsertChapter)

		chapterGroup.PUT("/:id",
			middlewares.AuthMiddleware(),
			middlewares.IsAuthorOfChapter(), // kiểm tra chương thuộc truyện của user
			controllers.UpdateChapter)

		chapterGroup.DELETE("/:id",
			middlewares.AuthMiddleware(),
			middlewares.IsAuthorOfChapter(),
			controllers.DeleteChapter)
		chapterGroup.POST("/comment", middlewares.AuthMiddleware(), controllers.InsertComment)
		chapterGroup.GET("/comments/:chapter_id", controllers.GetCommentsByChapterID)
	}
}
