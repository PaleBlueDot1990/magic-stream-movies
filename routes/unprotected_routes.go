package routes

import (
	controllers "github.com/PaleBlueDot1990/magic-stream-movies/controllers"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"github.com/gin-gonic/gin"
)

func SetupUnprotectedRoutes(router *gin.Engine, client *mongo.Client) {
	router.GET("/movies", controllers.GetMovies(client))
	router.POST("/register", controllers.RegisterUser(client))
	router.POST("/login", controllers.LoginUser(client))
}