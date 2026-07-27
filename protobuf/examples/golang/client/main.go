package main

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"

	user "github.com/engineering-mastery/protobuf/examples/golang/pkg"
)

const (
	address = "localhost:50051"
)

func waitForReady(ctx context.Context, conn *grpc.ClientConn) error {
	for {
		state := conn.GetState()
		if state == connectivity.Ready {
			return nil
		}

		if !conn.WaitForStateChange(ctx, state) {
			return ctx.Err()
		}
	}
}

func main() {
	// Устанавливаем соединение с сервером
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Did not connect: %v", err)
	}
	defer conn.Close()

	// Ждем пока сервер будет готов (wait for ready)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Println("Waiting for server to be ready...")
	if err := waitForReady(ctx, conn); err != nil {
		log.Fatalf("Server not ready: %v", err)
	}
	log.Println("Server is ready")

	// Создаем клиент для UserService
	client := user.NewUserServiceClient(conn)

	// Устанавливаем таймаут для запроса
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Вызываем метод GetUser
	log.Println("Sending request to get user...")
	response, err := client.GetUser(ctx, &user.GetUserRequest{UserId: "user/123"})
	if err != nil {
		log.Fatalf("GetUser failed: %v", err)
	}

	// Выводим результат
	log.Println("User received:")
	log.Printf("  Name: %s", response.GetName())
	log.Printf("  Email: %s", response.GetEmail())
	log.Printf("  Age: %d", response.GetAge())
	log.Printf("  Status: %v", response.GetStatus())

	// Пример вызова CreateUser
	log.Println("\nSending request to create user...")
	createResponse, err := client.CreateUser(ctx, &user.CreateUserRequest{
		Name:  "user/456",
		Email: "new.user@example.com",
		Age:   25,
		Metadata: map[string]string{
			"source": "client-test",
		},
	})
	if err != nil {
		log.Fatalf("CreateUser failed: %v", err)
	}

	// Выводим результат создания
	log.Println("User created:")
	log.Printf("  Name: %s", createResponse.GetName())
	log.Printf("  Email: %s", createResponse.GetEmail())
	log.Printf("  Age: %d", createResponse.GetAge())
	log.Printf("  Status: %v", createResponse.GetStatus())
	log.Printf("  Metadata: %v", createResponse.GetMetadata())
}
