package main

import (
	"context"
	"log"
	"net"

	"github.com/google/uuid"
	"github.com/kalpovskii/checklist/internal/app/pb"
	"github.com/kalpovskii/checklist/internal/app/repositories"
	"github.com/kalpovskii/checklist/internal/app/services"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type TaskServer struct {
	pb.UnimplementedTaskServiceServer
	service *services.TaskService
}

func (s *TaskServer) Create(ctx context.Context, req *pb.CreateTaskRequest) (*pb.TaskResponse, error) {
	task, err := s.service.Create(req.Title, req.Content)
	if err != nil {
		return nil, err
	}
	return &pb.TaskResponse{
		Task: &pb.Task{
			Id:        task.ID.String(),
			Title:     task.Title,
			Content:   task.Content,
			Done:      task.Done,
			CreatedAt: timestamppb.New(task.CreatedAt),
		},
	}, nil
}

func (s *TaskServer) List(ctx context.Context, req *emptypb.Empty) (*pb.TaskListResponse, error) {
    tasks, err := s.service.List()
    if err != nil {
        return nil, err
    }

    resp := &pb.TaskListResponse{}
    for _, t := range tasks {
        resp.Tasks = append(resp.Tasks, &pb.Task{
            Id:        t.ID.String(),
            Title:     t.Title,
            Content:   t.Content,
            Done:      t.Done,
            CreatedAt: timestamppb.New(t.CreatedAt),
        })
    }
    return resp, nil
}

func (s *TaskServer) Delete(ctx context.Context, req *pb.TaskIDRequest) (*pb.StatusResponse, error) {
	id, err := uuid.Parse(req.Id)
  if err != nil {
      return nil, err
	}
	
	err = s.service.Delete(id)
	if err != nil {
		return nil, err
	}
	return &pb.StatusResponse{Status: "deleted"}, nil
}

func (s *TaskServer) MarkDone(ctx context.Context, req *pb.TaskIDRequest) (*pb.StatusResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, err
	}
	err = s.service.MarkDone(id)
	if err != nil {
		return nil, err
	}
	return &pb.StatusResponse{Status: "done"}, nil
}

func main() {
	viper.SetEnvPrefix("CHECKLIST")
	viper.AutomaticEnv()

	port := viper.GetString("DB_GRPC_PORT")
  dsn  := viper.GetString("DB_POSTGRES_DSN")
	if dsn == "" || port == "" {
		log.Fatal("DB_POSTGRES_DSN or DB_GRPC_PORT is not configured")
	}

	repo, err := repositories.NewPostgresTaskRepo(dsn)
	if err != nil {
		log.Fatal(err)
	}

	service := services.NewTaskService(repo)
	server := &TaskServer{service: service}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterTaskServiceServer(grpcServer, server)

	log.Printf("gRPC server listening on %s", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}