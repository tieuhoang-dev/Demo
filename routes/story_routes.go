package routes

import (
	"Truyen_BE/controllers"
	middlewares "Truyen_BE/middleware"

	"github.com/gin-gonic/gin"
)

func StoryRoutes(router *gin.Engine) {

	storyGroup := router.Group("/stories") 
	{
		storyGroup.GET("", controllers.GetStories)
		storyGroup.GET("/search", controllers.SearchStoriesByName)
		storyGroup.GET("/:id/chapters", controllers.GetChaptersByStoryID)
		storyGroup.GET("/filter", controllers.FilterStories)
		storyGroup.GET("/ranking", controllers.GetTopRankedStories)
		storyGroup.GET("/:id/export", controllers.ExportStoryChapters)
		storyGroup.GET("/featured", controllers.GetFeaturedStories)
		storyGroup.GET("/:id", controllers.GetStoryContent)
		storyGroup.GET("/genre", controllers.GetGenresWithCount)
		storyGroup.GET("/genre/:genre", controllers.GetAllStoriesOfGenre)
		storyGroup.GET("/newest", controllers.GetNewestUpdatedStoryList)
		storyGroup.POST("", middlewares.AuthMiddleware(), middlewares.RequireRole("author"), controllers.InsertStory)
		storyGroup.PUT("/:id", middlewares.AuthMiddleware(), middlewares.AuthorOnly(), controllers.UpdateStory)
		storyGroup.DELETE("/:id", middlewares.AuthMiddleware(), middlewares.AuthorOnly(), controllers.DeleteStory)
	}

	author := router.Group("/my-stories")
	author.Use(middlewares.AuthMiddleware(), middlewares.AuthorOnly()) 
	{
		author.DELETE("/:title", controllers.DeleteStoryByAuthor)
	}

	admin := router.Group("/admin/stories")
	admin.Use(middlewares.AuthMiddleware(), middlewares.AdminOnly())
	{
		admin.PUT("/ban/:title", controllers.BanStory)
		admin.PUT("/unban/:title", controllers.UnbanStory)
	}
}
