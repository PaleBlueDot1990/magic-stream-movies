package controllers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	database "github.com/PaleBlueDot1990/magic-stream-movies/Server/MagicStreamMoviesServer/database"
	models "github.com/PaleBlueDot1990/magic-stream-movies/Server/MagicStreamMoviesServer/models"
	"github.com/PaleBlueDot1990/magic-stream-movies/Server/MagicStreamMoviesServer/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/tmc/langchaingo/llms/openai"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func GetMovies(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100 * time.Second)
		defer cancel()

		var movieCollection *mongo.Collection = database.OpenCollection("movies", client)
		cursor, err := movieCollection.Find(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch movies"})
			return 
		}
		defer cursor.Close(ctx)

		var movies []models.Movie
		if err = cursor.All(ctx, &movies); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode movies"})
			return 
		}

		c.JSON(http.StatusOK, movies)
	}
}

func GetMovie(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100 * time.Second)
		defer cancel()

		movieID := c.Param("imdb_id")
		if movieID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Movie ID is required"})
			return 
		}

		var movie models.Movie
		var movieCollection *mongo.Collection = database.OpenCollection("movies", client)
		err := movieCollection.FindOne(ctx, bson.M{"imdb_id": movieID}).Decode(&movie)
		
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Movie not found"})
			return 
		}

		c.JSON(http.StatusOK, movie)
	}
}

func AddMovie(client *mongo.Client) gin.HandlerFunc {
	return func(c * gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100 * time.Second)
		defer cancel()

		var movie models.Movie 
		if err := c.ShouldBindJSON(&movie); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return 
		}

		var validate *validator.Validate = validator.New()
		if err := validate.Struct(movie); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
			return 
		}

		var movieCollection *mongo.Collection = database.OpenCollection("movies", client)
		result, err := movieCollection.InsertOne(ctx, movie)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add movie"})
			return 
		}

		c.JSON(http.StatusCreated, result);
	}
}

func AdminReviewUpdate(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, err := utils.GetUserRoleFromContext(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error":"Role not found in context"})
			return 
		}

		if role != "ADMIN" {
			c.JSON(http.StatusUnauthorized, gin.H{"error":"User must be part of the ADMIN role"})
			return 
		}

		movieID := c.Param("imdb_id")
		if movieID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error":"Movie Id required"})
			return 
		}

		var req struct {
			AdminReview string `json:"admin_review"`
		}

		var res struct {
			RankingName string `json:"ranking_name"`
			AdminReview string `json:"admin_review"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error":"Invalid request body"})
			return 
		}

		sentiment, rankVal, err := GetReviewRanking(req.AdminReview, client)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error":"Error getting review ranking", "details": err.Error()})
			return 
		}

		filter := bson.M{"imdb_id":movieID}
		update := bson.M {
			"$set": bson.M {
				"admin_review": req.AdminReview,
				"ranking": bson.M {
					"ranking_value": rankVal,
					"ranking_name": sentiment,
				},
			},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 100 * time.Second)
		defer cancel()

		var movieCollection *mongo.Collection = database.OpenCollection("movies", client)
		result, err := movieCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error":"Error updating movie"})
			return 
		}

		if result.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error":"Movie not found"})
			return 
		}

		res.RankingName = sentiment
		res.AdminReview = req.AdminReview
		c.JSON(http.StatusOK, res)
	}
}

func GetReviewRanking(admin_review string, client *mongo.Client) (string, int, error) {
	rankings, err := GetRankings(client)
	if err != nil {
		return "", 0, err
	}

	sentimentDelimited := ""
	for _, ranking := range rankings {
		if ranking.RankingValue != 999 {
			sentimentDelimited = sentimentDelimited + ranking.RankingName + ","
		}
	}

	sentimentDelimited = strings.Trim(sentimentDelimited, ",")
	err = godotenv.Load(".env")
	if err != nil {
		log.Println("Warning: .env file not found")
	}

	openAiApiKey := os.Getenv("OPENAI_API_KEY")
	if openAiApiKey == "" {
		return "", 0, errors.New("could not read OPENAI API KEY")
	}

	llm, err := openai.New(openai.WithToken(openAiApiKey))
	if err != nil {
		return "", 0, err
	}

	basePromptTemplate := os.Getenv("BASE_PROMPT_TEMPLATE")
	basePrompt := strings.Replace(basePromptTemplate, "{rankings}", sentimentDelimited, 1)

	response, err := llm.Call(context.Background(), basePrompt + admin_review)
	if err != nil {
		return "", 0, err 
	}

	rankVal := 0
	for _, ranking := range rankings {
		if ranking.RankingName == response {
			rankVal = ranking.RankingValue
			break
		}
	}

	return response, rankVal, nil 
}

func GetRankings(client *mongo.Client) ([]models.Ranking, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 100 * time.Second)
	defer cancel()

	var rankingCollection *mongo.Collection = database.OpenCollection("rankings", client)
	cursor, err := rankingCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)
	var rankings []models.Ranking
	if err := cursor.All(ctx, &rankings); err != nil {
		return nil, err
	}

	return rankings, nil 
}

func GetRecommendedMovies(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId, err := utils.GetUserIdFromContext(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error":"User Id not found in context"})
			return 
		}

		favourite_genres, err := GetUsersFavouriteGenres(userId, client)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error":err.Error()})
			return 
		}

		err = godotenv.Load(".env")
		if err != nil {
			log.Println("Warning: .env file not found")
		}

		var recommendedMoviesLimitVal int64 = 5 
		recommendedMoviesLimitStr := os.Getenv("RECOMMENDED_MOVIE_LIMIT")
		if recommendedMoviesLimitStr != "" {
			recommendedMoviesLimitVal, _ = strconv.ParseInt(recommendedMoviesLimitStr, 10, 64)
		}

		findOptions := options.Find()
		findOptions.SetSort(bson.D{{Key: "ranking.ranking_value", Value:1}})
		findOptions.SetLimit(recommendedMoviesLimitVal)

		filter := bson.M{"genre.genre_name": bson.M{"$in":favourite_genres}}
		ctx, cancel := context.WithTimeout(context.Background(), 100 * time.Second)
		defer cancel()

		var movieCollection *mongo.Collection = database.OpenCollection("movies", client)
		cursor, err := movieCollection.Find(ctx, filter, findOptions)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching recommended movies"})
			return 
		}
		defer cursor.Close(c)

		var recommendedMovies []models.Movie 
		if err := cursor.All(ctx, &recommendedMovies); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return 
		}

		c.JSON(http.StatusOK, recommendedMovies)
	}
}

func GetUsersFavouriteGenres(userId string, client *mongo.Client) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 100 * time.Second)
	defer cancel()

	filter := bson.M{"user_id":userId}
	projections := bson.M{
		"favourite_genres.genre_name": 1,
		"_id": 0,
	}

	opts := options.FindOne().SetProjection(projections)
	var results bson.M

	var userCollection *mongo.Collection = database.OpenCollection("users", client)
	err := userCollection.FindOne(ctx, filter, opts).Decode(&results)
	if err != nil {
		if err == mongo.ErrNoDocuments  {
			return []string{}, nil 
		}
		return []string{}, err
	}

	favGenresArray, ok := results["favourite_genres"].(bson.A)
	if !ok {
		return []string{}, errors.New("unable to retrieve favorite genres for user")
	}

	var genreNames []string 
	for _, item := range favGenresArray {
		if genreMap, ok := item.(bson.D); ok {
			for _, elem := range genreMap {
				if elem.Key == "genre_name" {
					if name, ok := elem.Value.(string); ok {
						genreNames = append(genreNames, name)
					}
				}
			}
		}
	}

	return genreNames, nil 
}

