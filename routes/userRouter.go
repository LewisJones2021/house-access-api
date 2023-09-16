package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/lewisjones2021/house-access-api/controllers"
)

// userRoutes function
func UserRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.POST("/users/signup", controllers.SignUp())
	incomingRoutes.POST("/users/login", controllers.Login())
}
