package routes

import (
	"Truyen_BE/controllers"
	middlewares "Truyen_BE/middleware"

	"github.com/gin-gonic/gin"
)

// StoryRoutes định nghĩa các route cho truyện
// Các route này bao gồm cả các route công khai và các route yêu cầu xác thực
func StoryRoutes(router *gin.Engine) {

	storyGroup := router.Group("/stories")
	{
		// ✅ Public APIs
		storyGroup.GET("", controllers.GetStories)
		storyGroup.GET("/search", controllers.SearchStoriesByName)
		storyGroup.GET("/:id/chapters", controllers.GetChaptersByStoryID)
		storyGroup.GET("/filter", controllers.FilterStories)
		storyGroup.GET("/ranking", controllers.GetTopRankedStories)
		storyGroup.GET("/:id/export", controllers.ExportStoryChapters)
		storyGroup.GET("/featured", controllers.GetFeaturedStories)

		// ✅ APIs yêu cầu xác thực và vai trò
		storyGroup.POST("", middlewares.AuthMiddleware(), middlewares.RequireRole("author"), controllers.InsertStory)
		storyGroup.PUT("/:id", middlewares.AuthMiddleware(), middlewares.AuthorOnly(), controllers.UpdateStory)
		storyGroup.DELETE("/:id", middlewares.AuthMiddleware(), middlewares.AuthorOnly(), controllers.DeleteStory)
	}

	// ✅ Tác giả tự xoá truyện đã bị ẩn
	author := router.Group("/my-stories")
	author.Use(middlewares.AuthMiddleware(), middlewares.AuthorOnly()) // Kiểm tra là tác giả
	{
		author.DELETE("/:title", controllers.DeleteStoryByAuthor)
	}

	// ✅ Quản trị viên: ban truyện
	admin := router.Group("/admin/stories")
	admin.Use(middlewares.AuthMiddleware(), middlewares.AdminOnly())
	{
		admin.PUT("/ban/:title", controllers.BanStory)
		admin.PUT("/unban/:title", controllers.UnbanStory)
	}
}
