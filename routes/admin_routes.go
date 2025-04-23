package routes

import (
	"Truyen_BE/controllers"
	 "Truyen_BE/middleware"

	"github.com/gin-gonic/gin"
)

func AdminRoutes(r *gin.Engine) {
	admin := r.Group("/admin")
	admin.Use(middlewares.AuthMiddleware(), middlewares.AdminOnly())
	{
		admin.PUT("/stories/:title/ban", controllers.BanStory)
		admin.PUT("/stories/:title/unban", controllers.UnbanStory)
	}
}
