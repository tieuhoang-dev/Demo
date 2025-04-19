package routes

import (
	"Truyen_BE/controllers"

	"github.com/gin-gonic/gin"
)

func StoryRoutes(router *gin.Engine) {

	storyGroup := router.Group("/stories")
	{
		storyGroup.GET("", controllers.GetStories)                 // GET /stories
		storyGroup.GET("/search", controllers.SearchStoriesByName) // GET /stories/search?name=abc
		storyGroup.POST("", controllers.InsertStory)               // POST /stories
		storyGroup.PUT("/:name", controllers.UpdateStory)          // PUT /stories/:name
		storyGroup.DELETE("/:name", controllers.DeleteStory)       // DELETE /stories/:name
		storyGroup.GET("/chapter", controllers.GetChaptersByStoryName)
	}
}
