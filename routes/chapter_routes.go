package routes

import (
	"Truyen_BE/controllers"

	"github.com/gin-gonic/gin"
)

func ChapterRoutes(router *gin.Engine) {
	chapterGroup := router.Group("/stories/chapters")
	{
		chapterGroup.POST("", controllers.InsertChapter) // POST /chapters
		//chapterGroup.PUT("/:chapterId", controllers.UpdateChapter) // PUT /chapters/:chapterId
		//chapterGroup.DELETE("/:chapterId", controllers.DeleteChapter) // DELETE /chapters/:chapterId
	}
}
