package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kalpovskii/checklist/internal/app/pb"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type taskClientStub struct {
	createFn   func(ctx context.Context, in *pb.CreateTaskRequest, opts ...grpc.CallOption) (*pb.TaskResponse, error)
	listFn     func(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*pb.TaskListResponse, error)
	deleteFn   func(ctx context.Context, in *pb.TaskIDRequest, opts ...grpc.CallOption) (*pb.StatusResponse, error)
	markDoneFn func(ctx context.Context, in *pb.TaskIDRequest, opts ...grpc.CallOption) (*pb.StatusResponse, error)
}

func (s *taskClientStub) Create(ctx context.Context, in *pb.CreateTaskRequest, opts ...grpc.CallOption) (*pb.TaskResponse, error) {
	return s.createFn(ctx, in, opts...)
}

func (s *taskClientStub) List(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*pb.TaskListResponse, error) {
	return s.listFn(ctx, in, opts...)
}

func (s *taskClientStub) Delete(ctx context.Context, in *pb.TaskIDRequest, opts ...grpc.CallOption) (*pb.StatusResponse, error) {
	return s.deleteFn(ctx, in, opts...)
}

func (s *taskClientStub) MarkDone(ctx context.Context, in *pb.TaskIDRequest, opts ...grpc.CallOption) (*pb.StatusResponse, error) {
	return s.markDoneFn(ctx, in, opts...)
}

func setupTestRouter(stub *taskClientStub) (*gin.Engine, func()) {
	gin.SetMode(gin.TestMode)

	prevClient := taskClient
	taskClient = stub

	router := gin.Default()
	router.POST("/create", func(c *gin.Context) { createHandler(c, nil) })
	router.GET("/list", func(c *gin.Context) { listHandler(c, nil) })
	router.DELETE("/delete", func(c *gin.Context) { deleteHandler(c, nil) })
	router.PUT("/done", func(c *gin.Context) { doneHandler(c, nil) })

	cleanup := func() {
		taskClient = prevClient
	}

	return router, cleanup
}

func TestCreateHandlerSuccess(t *testing.T) {
	stub := &taskClientStub{
		createFn: func(ctx context.Context, in *pb.CreateTaskRequest, _ ...grpc.CallOption) (*pb.TaskResponse, error) {
			if in.Title != "title" || in.Content != "content" {
				t.Fatalf("unexpected payload: %+v", in)
			}
			return &pb.TaskResponse{Task: &pb.Task{Id: "1", Title: in.Title, Content: in.Content}}, nil
		},
	}

	router, cleanup := setupTestRouter(stub)
	defer cleanup()

	body := `{"title":"title","content":"content"}`
	req := httptest.NewRequest(http.MethodPost, "/create", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.Code)
	}

	var got pb.Task
	if err := json.Unmarshal(resp.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if got.Id != "1" || got.Title != "title" || got.Content != "content" {
		t.Fatalf("unexpected task response: %+v", got)
	}
}

func TestCreateHandlerBadJSON(t *testing.T) {
	stub := &taskClientStub{
		createFn: func(ctx context.Context, in *pb.CreateTaskRequest, _ ...grpc.CallOption) (*pb.TaskResponse, error) {
			return nil, nil
		},
	}

	router, cleanup := setupTestRouter(stub)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/create", strings.NewReader("{invalid"))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.Code)
	}
}

func TestCreateHandlerGrpcError(t *testing.T) {
	stub := &taskClientStub{
		createFn: func(ctx context.Context, in *pb.CreateTaskRequest, _ ...grpc.CallOption) (*pb.TaskResponse, error) {
			return nil, errors.New("grpc failure")
		},
	}

	router, cleanup := setupTestRouter(stub)
	defer cleanup()

	body := `{"title":"title","content":"content"}`
	req := httptest.NewRequest(http.MethodPost, "/create", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", resp.Code)
	}
}

func TestListHandlerSuccess(t *testing.T) {
	stub := &taskClientStub{
		listFn: func(ctx context.Context, in *emptypb.Empty, _ ...grpc.CallOption) (*pb.TaskListResponse, error) {
			return &pb.TaskListResponse{
				Tasks: []*pb.Task{
					{Id: "1", Title: "t1"},
					{Id: "2", Title: "t2"},
				},
			}, nil
		},
	}

	router, cleanup := setupTestRouter(stub)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/list", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.Code)
	}

	var got []*pb.Task
	if err := json.Unmarshal(resp.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(got) != 2 || got[0].Id != "1" || got[1].Id != "2" {
		t.Fatalf("unexpected list response: %+v", got)
	}
}

func TestDeleteHandlerSuccess(t *testing.T) {
	stub := &taskClientStub{
		deleteFn: func(ctx context.Context, in *pb.TaskIDRequest, _ ...grpc.CallOption) (*pb.StatusResponse, error) {
			if in.Id != "123" {
				t.Fatalf("unexpected id: %s", in.Id)
			}
			return &pb.StatusResponse{Status: "ok"}, nil
		},
	}

	router, cleanup := setupTestRouter(stub)
	defer cleanup()

	body := `{"id":"123"}`
	req := httptest.NewRequest(http.MethodDelete, "/delete", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.Code)
	}

	var got pb.StatusResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if got.Status != "ok" {
		t.Fatalf("unexpected status response: %+v", got)
	}
}

func TestDoneHandlerSuccess(t *testing.T) {
	stub := &taskClientStub{
		markDoneFn: func(ctx context.Context, in *pb.TaskIDRequest, _ ...grpc.CallOption) (*pb.StatusResponse, error) {
			if in.Id != "555" {
				t.Fatalf("unexpected id: %s", in.Id)
			}
			return &pb.StatusResponse{Status: "done"}, nil
		},
	}

	router, cleanup := setupTestRouter(stub)
	defer cleanup()

	body := `{"id":"555"}`
	req := httptest.NewRequest(http.MethodPut, "/done", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.Code)
	}

	var got pb.StatusResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if got.Status != "done" {
		t.Fatalf("unexpected status response: %+v", got)
	}
}
