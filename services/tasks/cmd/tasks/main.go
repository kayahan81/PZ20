package main

import (
	"net/http"
	"os"

	"tech-ip-sem2/services/tasks/internal/client/authgrpc"
	"tech-ip-sem2/services/tasks/internal/handler"
	"tech-ip-sem2/services/tasks/internal/middleware"
	"tech-ip-sem2/services/tasks/internal/storage"
	"tech-ip-sem2/shared/logger"
	sharedMiddleware "tech-ip-sem2/shared/middleware"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func main() {
	log := logger.New("tasks")
	defer log.Sync()

	port := os.Getenv("TASKS_PORT")
	if port == "" {
		port = "8082"
	}

	authGRPCAddr := os.Getenv("AUTH_GRPC_ADDR")
	if authGRPCAddr == "" {
		authGRPCAddr = "localhost:50051"
	}

	log.Info("starting tasks service",
		zap.String("http_port", port),
		zap.String("auth_grpc_addr", authGRPCAddr),
	)

	// Создаем gRPC клиент
	authClient, err := authgrpc.NewClient(authGRPCAddr, log)
	if err != nil {
		log.Fatal("failed to connect to auth", zap.Error(err))
	}
	defer authClient.Close()

	// Инициализация
	taskStorage := storage.NewMemoryStorage()
	tasksHandler := handler.NewTasksHandler(taskStorage, log)
	authMiddleware := middleware.NewAuthMiddleware(authClient, log)

	// Роутер
	mux := http.NewServeMux()

	// Защищенные маршруты
	mux.Handle("POST /v1/tasks", authMiddleware.RequireAuth(http.HandlerFunc(tasksHandler.CreateTask)))
	mux.Handle("GET /v1/tasks", authMiddleware.RequireAuth(http.HandlerFunc(tasksHandler.GetTasks)))
	mux.Handle("GET /v1/tasks/{id}", authMiddleware.RequireAuth(http.HandlerFunc(tasksHandler.GetTask)))
	mux.Handle("PATCH /v1/tasks/{id}", authMiddleware.RequireAuth(http.HandlerFunc(tasksHandler.UpdateTask)))
	mux.Handle("DELETE /v1/tasks/{id}", authMiddleware.RequireAuth(http.HandlerFunc(tasksHandler.DeleteTask)))

	// Метрики (без авторизации)
	mux.Handle("GET /metrics", promhttp.Handler())

	// Цепочка middleware: request-id → access log → метрики
	handler := sharedMiddleware.RequestIDMiddleware(
		sharedMiddleware.AccessLogMiddleware(log)(
			middleware.MetricsMiddleware(mux),
		),
	)

	log.Info("HTTP server listening", zap.String("port", port))
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal("failed to serve", zap.Error(err))
	}
}
