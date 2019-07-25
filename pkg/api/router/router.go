package router

import (
	_ "zeus/docs"
	"zeus/pkg/api/controllers"
	"zeus/pkg/api/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
)

func Init(e *gin.Engine) {
	e.Use(
		gin.Recovery(),
	)
	e.Use(cors.Default()) // CORS
	e.Use(middleware.SetLangVer())
	e.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	e.GET("/test", controllers.Healthy)
	//version fragment
	v1 := e.Group("/v1")

	//auth handlers
	auth := v1.Group("/auth", gin.BasicAuth(gin.Accounts{
		"zeus": "2019@win",
	}))
	jwtAuth := middleware.JwtAuth()
	auth.POST("/token", jwtAuth.LoginHandler)
	auth.GET("/refresh_token", jwtAuth.RefreshHandler)

	//api handlers
	api := v1.Group("/api")
	api.Use(jwtAuth.MiddlewareFunc(), middleware.JwtPrepare)

	userController := controllers.UserController{}

	//login
	api.GET("/login/info", userController.Info)
	//user
	api.GET("/users", userController.List)
	api.GET("/users/:id", userController.Get)
	api.PATCH("/users/:id", userController.Edit)
	api.DELETE("/users/:id", userController.Delete)

	roleController := controllers.RoleController{}
	//role
	api.GET("/roles", roleController.List)
	api.GET("/roles/:id", roleController.Get)
}
