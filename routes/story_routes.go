package routes

import (
	"Truyen_BE/controllers"

	"github.com/gin-gonic/gin"
)

func StoryRoutes(router *gin.Engine) {

	storyGroup := router.Group("/stories")
	{
		storyGroup.GET("", controllers.GetStories)
		storyGroup.GET("/search", controllers.SearchStoriesByName)
		storyGroup.POST("", controllers.InsertStory)
		storyGroup.PUT("/:id", controllers.UpdateStory)
		storyGroup.DELETE("/:id", controllers.DeleteStory)
		storyGroup.GET("/:id/chapters", controllers.GetChaptersByStoryID)
		storyGroup.GET("/filter", controllers.FilterStories)
		storyGroup.GET("/ranking", controllers.GetTopRankedStories)
		storyGroup.GET("/:id/export", controllers.ExportStoryChapters)
		storyGroup.GET("featured", controllers.GetFeaturedStories)
	}
}
