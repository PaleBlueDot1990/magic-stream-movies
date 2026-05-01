package main 

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
	
	routes "github.com/PaleBlueDot1990/magic-stream-movies/routes"
	database "github.com/PaleBlueDot1990/magic-stream-movies/database"
)

func main() {
	router := gin.Default()

	router.GET("/hello", func(c *gin.Context) {
		c.String(200, "Hello, Magic Stream Movies!")
	})

	var client *mongo.Client = database.Connect()
	routes.SetupUnprotectedRoutes(router, client)
	routes.SetupProtectedRoutes(router, client)
	
	if err := router.Run(":8080"); err != nil {
		fmt.Println("Failed to start server: ", err)
	}
}