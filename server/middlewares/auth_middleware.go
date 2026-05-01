package middlewares

import (
	"net/http"

	utils "github.com/PaleBlueDot1990/magic-stream-movies/Server/MagicStreamMoviesServer/utils"

	"github.com/gin-gonic/gin"
)

func AuthMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := utils.GetAccessToken(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error":err.Error()})
			c.Abort()
			return 
		}

		claims, err := utils.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"err":"Invalid token"})
			c.Abort()
			return 
		}

		c.Set("userId", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}

