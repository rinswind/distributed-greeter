package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/rinswind/auth-go/tokens"
	"github.com/rinswind/distributed-greeter/login/internal/server"
	"github.com/rinswind/distributed-greeter/login/internal/users"
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
	atExp, err := strconv.Atoi(os.Getenv("ACCESS_TOKEN_EXPIRY"))
	if err != nil {
		panic(err)
	}
	atExpiry := time.Minute * time.Duration(atExp)

	// Init refresh token settings
	rtSecret := os.Getenv("REFRESH_TOKEN_SECRET")
	rtExp, err := strconv.Atoi(os.Getenv("REFRESH_TOKEN_EXPIRY"))
	if err != nil {
		panic(err)
	}
	rtExpiry := time.Minute * time.Duration(rtExp)

	le := server.LoginEndpoint{
		Iface:      iface,
		AuthReader: &tokens.AuthReader{Redis: redis, ATSecret: atSecret, RTSecret: rtSecret},
		AuthWriter: &tokens.AuthWriter{Redis: redis, ATSecret: atSecret, ATExpiry: atExpiry, RTSecret: rtSecret, RTExpiry: rtExpiry},
		Users:      users.Make(redis),
	}
	le.Run()
}
