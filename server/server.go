package server

import (
	"api-test/handlers"
	"github.com/gin-contrib/cors"
	"log"

	"github.com/gin-gonic/gin"
)

func NewServer() {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:5173"},
		AllowMethods: []string{
			"GET",
			"POST",
			"OPTIONS",
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Length",
			"Content-Type",
			"Authorization",
		},
	}))

	api := r.Group("/api")
	{
		api.POST("/register", handlers.RegisterUser)
		api.POST("/login", handlers.LoginUser)

		api.POST("/interact", handlers.CreateInteraction)

		task := api.Group("/task")
		{
			task.POST("/generate", handlers.GenerateTask)
			task.POST("/save", handlers.SaveTask)
			task.POST("/evaluate", handlers.EvaluateTask)
		}

		chat := api.Group("/chat")
		{
			chat.POST("/task-question", handlers.TaskSendChat)
		}

		user := api.Group("/user")
		{
			user.GET("/tasks", handlers.GetUserTasks)
			user.GET("/task/:task_id", handlers.GetSingleTask)

			stats := user.Group("/stats")
			{
				stats.GET("/general", handlers.GetUserStats)
				stats.GET("/full", handlers.GetUserStatsFull)
				stats.GET("/language", handlers.GetUserStatsLanguage)
			}

			settings := user.Group("/settings")
			{
				settings.POST("/change-username", handlers.ChangeUsername)
				settings.POST("/change-password", handlers.ChangePassword)
				settings.POST("/delete-account", handlers.DeleteAccount)
			}
		}
	}

	err := r.Run(":8888")
	if err != nil {
		log.Fatal(err)
	}
}
