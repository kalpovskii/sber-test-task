package main

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kalpovskii/checklist/internal/app/pb"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

var taskClient pb.TaskServiceClient

func initConfig() {
	viper.SetEnvPrefix("CHECKLIST")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
}

func main() {
	initConfig()

  grpcURL := viper.GetString("DB_GRPC_URL")
  apiPort := viper.GetString("API_PORT")

	if apiPort == "" || grpcURL == "" {
		log.Fatal("API_PORT or DB_GRPC_URL is not configured")
	}

	// connect to gRPC server
	conn, err := grpc.Dial(grpcURL, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	taskClient = pb.NewTaskServiceClient(conn)

	log.Printf("API started on :%s", apiPort)
	log.Printf("Connected to gRPC DB at %s", grpcURL)

	r := gin.Default()

	r.POST("/create", createHandler)
	r.GET("/list", listHandler)
	r.DELETE("/delete", deleteHandler)
	r.PUT("/done", doneHandler)

	log.Fatal(r.Run(":" + apiPort))
}

func createHandler(c *gin.Context) {
	var req struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := taskClient.Create(ctx, &pb.CreateTaskRequest{
		Title:   req.Title,
		Content: req.Content,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res.Task)
}

func listHandler(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := taskClient.List(ctx, &emptypb.Empty{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res.Tasks)
}

func deleteHandler(c *gin.Context) {
	var req struct {
		ID string `json:"id"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := taskClient.Delete(ctx, &pb.TaskIDRequest{Id: req.ID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func doneHandler(c *gin.Context) {
	var req struct {
		ID string `json:"id"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := taskClient.MarkDone(ctx, &pb.TaskIDRequest{Id: req.ID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}