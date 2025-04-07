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
		// storyGroup.GET("/:id", controllers.GetStoryByID)          // GET /stories/:id (nếu cần)
		storyGroup.POST("", controllers.InsertStory) // POST /stories
		// storyGroup.PUT("/:id", controllers.UpdateStory)           // PUT /stories/:id
		// storyGroup.DELETE("/:id", controllers.DeleteStory)        // DELETE /stories/:id
	}
}
