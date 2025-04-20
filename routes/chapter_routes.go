package routes

import (
	"Truyen_BE/controllers"

	"github.com/gin-gonic/gin"
)

func ChapterRoutes(router *gin.Engine) {
	chapterGroup := router.Group("/stories/chapters")
	{
		chapterGroup.POST("", controllers.InsertChapter)                               // POST /chapters
		chapterGroup.PUT("/:id", controllers.UpdateChapter)                            // PUT /chapters/:chapterId
		chapterGroup.DELETE("/:id", controllers.DeleteChapter)                         // DELETE /chapters/:chapterId
		chapterGroup.GET("/:story_id/:number", controllers.GetChapterByStoryAndNumber) // GET /stories/:story_id/:number
		chapterGroup.GET("/id/:id", controllers.GetChapterByID)                        // GET /chapters/id/:id
	}
}
