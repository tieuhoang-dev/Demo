package routes

import (
	"Truyen_BE/controllers"
	middlewares "Truyen_BE/middleware"

	"github.com/gin-gonic/gin"
)

func UserRoutes(router *gin.Engine) {
	// Không cần đăng nhập
	public := router.Group("/users")
	{
		public.POST("/register", controllers.RegisterUser)
		public.POST("/auth/login", controllers.LoginUser)
	}

	// Cần đăng nhập
	protected := router.Group("/users")
	protected.Use(middlewares.AuthMiddleware())
	{
		protected.GET("/me", controllers.GetCurrentUser)
		//protected.GET("/:id", controllers.GetUserByID)
		//protected.PUT("/:id", controllers.UpdateUser)
		//protected.DELETE("/:id", controllers.DeleteUser)
	}
}
