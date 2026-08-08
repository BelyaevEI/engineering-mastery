package main

import (
	"context"
	user "github.com/engineering-mastery/protobuf/examples/golang/pkg"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
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

// LoggingInterceptor — interceptor для логирования вызовов методов
func LoggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	log.Printf("[Interceptor] Calling method: %s", info.FullMethod)

	// Продолжаем обработку запроса
	resp, err := handler(ctx, req)

	if err != nil {
		log.Printf("[Interceptor] Method %s returned error: %v", info.FullMethod, err)
	} else {
		log.Printf("[Interceptor] Method %s completed successfully", info.FullMethod)
	}

	return resp, err
}

// AuthInterceptor — interceptor для проверки авторизации
func AuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Проверяем наличие токена в metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Printf("[Auth] No metadata in context")
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	// Проверяем заголовок "authorization"
	tokens := md["authorization"]
	if len(tokens) == 0 {
		log.Printf("[Auth] No authorization token provided")
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	// Простая проверка токена (в реальном приложении здесь была бы валидация JWT или т.д.)
	actualToken := tokens[0]
	if actualToken != "Bearer secret-token" {
		log.Printf("[Auth] Invalid token: %s", actualToken)
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	log.Printf("[Auth] Token validated successfully")

	// Продолжаем обработку запроса
	return handler(ctx, req)
}

// GetUser - получение пользователя по ID
func (s *server) GetUser(ctx context.Context, req *user.GetUserRequest) (*user.User, error) {
	log.Printf("Received request for user ID: %s", req.GetUserId())

	// Пример: если текущее время на сервере делится на 2, возвращаем пользователя
	// Иначе возвращаем ошибку 403 (PermissionDenied)
	now := time.Now()
	if now.Second()%2 == 0 {
		log.Printf("Time %v: second %d is even - returning user", now, now.Second())
		return &user.User{
			Name:      req.GetUserId(),
			Email:     "john.doe@example.com",
			Age:       30,
			Status:    user.UserStatus_ACTIVE,
			CreatedAt: timestamppb.New(time.Now()),
			UpdatedAt: timestamppb.New(time.Now()),
		}, nil
	}

	// Возвращаем ошибку 403 (PermissionDenied) с детальным описанием
	log.Printf("Time %v: second %d is odd - returning 403 error", now, now.Second())
	return nil, status.Error(codes.PermissionDenied, "access denied: server time validation failed")
}

// CreateUser - создание нового пользователя
func (s *server) CreateUser(ctx context.Context, req *user.CreateUserRequest) (*user.User, error) {
	log.Printf("Received request to create user: name=%s, email=%s, age=%d",
		req.GetName(), req.GetEmail(), req.GetAge())

	// Генерируем user_id, если он не указан
	userID := req.GetName()
	if userID == "" {
		// Генерируем ID в формате user/{uuid}
		generatedID := uuid.New().String()
		log.Printf("Generated user ID: %s", generatedID)
		userID = generatedID
	}

	// Возвращаем пользователя с сгенерированным ID и данными из запроса
	return &user.User{
		Name:      userID,
		Email:     req.GetEmail(),
		Age:       req.GetAge(),
		Status:    user.UserStatus_ACTIVE,
		Metadata:  req.GetMetadata(),
		CreatedAt: timestamppb.New(time.Now()),
		UpdatedAt: timestamppb.New(time.Now()),
	}, nil
}

func main() {
	// Создаем gRPC сервер с interceptor-ами
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(LoggingInterceptor),
		grpc.UnaryInterceptor(AuthInterceptor),
	)

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
