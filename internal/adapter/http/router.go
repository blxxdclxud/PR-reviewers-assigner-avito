package http

import (
	"github.com/blxxdclxud/PR-reviewers-assigner-avito/internal/adapter/http/handler"
	"github.com/blxxdclxud/PR-reviewers-assigner-avito/internal/usecase"
	"github.com/gin-gonic/gin"
)

func SetupRouter(
	teamUC *usecase.TeamUseCase,
	userUC *usecase.UserUseCase,
	prUC *usecase.PRUseCase,
	statsUC *usecase.StatsUseCase) *gin.Engine {

	router := gin.Default()

	// Handlers
	teamHandler := handler.NewTeamHandler(teamUC)
	userHandler := handler.NewUserHandler(userUC)
	prHandler := handler.NewPRHandler(prUC)

	statsHandler := handler.NewStatsHandler(statsUC)

	// Health check endpoint
	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Add middlewares
	//router.Use()

	// User endpoints
	user := router.Group("/users")
	{
		user.POST("/setIsActive", userHandler.SetIsActive)
		user.GET("/getReview", userHandler.GetReview)
	}

	// Team endpoints
	team := router.Group("/team")
	{
		team.POST("/add", teamHandler.Add)
		team.GET("/get", teamHandler.Get)
	}

	// Pull Request endpoints
	pr := router.Group("/pullRequest")
	{
		pr.POST("/create", prHandler.Create)
		pr.POST("/merge", prHandler.Merge)
		pr.POST("/reassign", prHandler.Reassign)
	}

	router.GET("/stats", statsHandler.GetStats)

	return router
}
