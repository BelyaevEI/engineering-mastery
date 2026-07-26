package main

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"

	user "github.com/engineering-mastery/protobuf/examples/go/pkg"
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
	response, err := client.GetUser(ctx, &user.GetUserRequest{Id: "12345"})
	if err != nil {
		log.Fatalf("GetUser failed: %v", err)
	}

	// Выводим результат
	log.Println("User received:")
	log.Printf("  ID: %s", response.GetId())
	log.Printf("  Name: %s", response.GetName())
	log.Printf("  Email: %s", response.GetEmail())
	log.Printf("  Age: %d", response.GetAge())
}
