package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kalpovskii/checklist/internal/app/repositories"
	"github.com/kalpovskii/checklist/internal/app/services"
	"github.com/spf13/viper"
)

func initConfig() {
	// if _, err := os.Stat(".env"); err == nil {
	// 	if err := godotenv.Load(); err != nil {
	// 		log.Fatalf("failed to load .env: %v", err)
	// 	}
	// 	log.Println("Config loaded from .env")
	// } else {
	// 	log.Fatal(".env file is missing")
	// }

	viper.SetEnvPrefix("CHECKLIST")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
}

func main() {
	initConfig()

	dsn := viper.GetString("DB_POSTGRES_DSN")
	port := viper.GetString("DB_PORT") 

	if dsn == "" || port == "" {
		log.Fatal("db.postgres.dsn or db.port is not configured")
	}

	repo, err := repositories.NewPostgresTaskRepo(dsn)
	if err != nil {
		log.Fatal(err)
	}

	service := services.NewTaskService(repo)

	r := gin.Default()

	r.POST("/create", func(c *gin.Context) {
		var req struct {
			Title   string `json:"title"`
			Content string `json:"content"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		task, err := service.Create(req.Title, req.Content)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, task)
	})

	r.GET("/list", func(c *gin.Context) {
		tasks, err := service.List()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, tasks)
	})

	r.DELETE("/delete", func(c *gin.Context) {
		var req struct {
			ID string `json:"id"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		id, err := uuid.Parse(req.ID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
			return
		}
		err = service.Delete(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "deleted"})
	})

	r.PUT("/done", func(c *gin.Context) {
		var req struct {
			ID string `json:"id"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		id, err := uuid.Parse(req.ID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
			return
		}
		err = service.MarkDone(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "done"})
	})

	log.Fatal(r.Run(":" + port))
}