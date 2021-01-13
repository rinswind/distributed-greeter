package main

import (
	"context"
	"fmt"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/rinswind/auth-go/tokens"
	"github.com/rinswind/distributed-greeter/greeter/internal/server"
	"github.com/rinswind/distributed-greeter/greeter/internal/users"
)

func main() {
	// Read the HTTP port
	port := os.Getenv("HTTP_PORT")
	iface := fmt.Sprintf(":%v", port)

	// Init Redis client
	dsn := os.Getenv("REDIS_DSN")
	redis := redis.NewClient(&redis.Options{
		Addr: dsn, //redis port
	})
	_, err := redis.Ping(context.Background()).Result()
	if err != nil {
		panic(err)
	}

	// Init access token settings
	atSecret := os.Getenv("ACCESS_TOKEN_SECRET")

	// Init refresh token settings
	rtSecret := os.Getenv("REFRESH_TOKEN_SECRET")

	// Create the auth session manager
	ar := &tokens.AuthReader{Redis: redis, ATSecret: atSecret, RTSecret: rtSecret}

	// Create the user db client
	users := users.Make(redis)
	users.Listen()

	ge := server.GreeterEndpoint{Iface: iface, AuthReader: ar, Users: users}
	ge.Run()
}
