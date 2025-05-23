package routes

import (
	"Truyen_BE/controllers"
	middlewares "Truyen_BE/middleware"

	"github.com/gin-gonic/gin"
)

func BookshelfRoutes(router *gin.Engine) {
	bookshelfGroup := router.Group("/bookshelf")
	bookshelfGroup.Use(middlewares.AuthMiddleware())
	{

		bookshelfGroup.GET("", controllers.GetBookshelf)                     // GET /bookshelf
		bookshelfGroup.DELETE("/:story_id", controllers.RemoveFromBookshelf) // DELETE /bookshelf/:story_id	}
		bookshelfGroup.POST("", controllers.UpdateLastChapter)              // PUT /bookshelf/last-chapter
	}
}
