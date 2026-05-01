package routes

import (
	controllers "github.com/PaleBlueDot1990/magic-stream-movies/controllers"
	middlewares "github.com/PaleBlueDot1990/magic-stream-movies/middlewares"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"github.com/gin-gonic/gin"
)

func SetupProtectedRoutes(router *gin.Engine, client *mongo.Client) {
	router.Use(middlewares.AuthMiddleWare())

	router.GET("/movie/:imdb_id", controllers.GetMovie(client))
	router.POST("/addmovie", controllers.AddMovie(client))
	router.GET("/recommendedmovies", controllers.GetRecommendedMovies(client))
	router.PATCH("/updatereview/:imdb_id", controllers.AdminReviewUpdate(client))
}


