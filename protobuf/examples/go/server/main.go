package main

import (
	"context"
	"log"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"

	user "github.com/engineering-mastery/protobuf/examples/go/pkg"
)

const (
	grpcPort    = ":50051"
	httpPort    = ":8080"
	grpcAddress = "localhost:50051"
)

// server - реализация UserServiceServer
type server struct {
	user.UnimplementedUserServiceServer
}

// GetUser - получение пользователя по ID
func (s *server) GetUser(ctx context.Context, req *user.GetUserRequest) (*user.User, error) {
	log.Printf("Received request for user ID: %s", req.GetId())

	// В реальном приложении здесь был бы запрос к базе данных
	// Для примера возвращаем фиктивные данные
	return &user.User{
		Id:    req.GetId(),
		Name:  "John Doe",
		Email: "john.doe@example.com",
		Age:   30,
	}, nil
}

func main() {
	// Создаем gRPC сервер
	grpcServer := grpc.NewServer()

	// Регистрируем сервис
	user.RegisterUserServiceServer(grpcServer, &server{})

	// Запускаем gRPC сервер в отдельной горутине
	log.Printf("Starting gRPC server on %s", grpcPort)
	go func() {
		lis, err := net.Listen("tcp", grpcPort)
		if err != nil {
			log.Fatalf("Failed to listen on gRPC port: %v", err)
		}
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// Создаем HTTP сервер для gRPC Gateway
	mux := runtime.NewServeMux()
	conn, err := grpc.NewClient(
		grpcAddress,
		grpc.WithTransportCredentials(nil), // Для localhost можно без TLS
	)
	if err != nil {
		log.Fatalf("Failed to create gRPC connection: %v", err)
	}
	defer conn.Close()

	// Регистрируем HTTP handler для UserService
	if err := user.RegisterUserServiceHandler(context.Background(), mux, conn); err != nil {
		log.Fatalf("Failed to register handler: %v", err)
	}

	// Запускаем HTTP сервер
	httpServer := &http.Server{
		Addr:    httpPort,
		Handler: mux,
	}

	log.Printf("Starting HTTP server on %s", httpPort)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to serve HTTP: %v", err)
	}
}
