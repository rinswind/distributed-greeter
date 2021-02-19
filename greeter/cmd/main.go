package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
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
	defer redis.Close()

	// Init access token settings
	atSecret := os.Getenv("ACCESS_TOKEN_SECRET")

	// Init refresh token settings
	rtSecret := os.Getenv("REFRESH_TOKEN_SECRET")

	// Create the auth session manager
	ar := &tokens.AuthReader{Redis: redis, ATSecret: atSecret, RTSecret: rtSecret}

	// Create the DB client
	dbAddr := os.Getenv("DB_ADDR")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	db, err := sql.Open("mysql", fmt.Sprintf("%v:%v@%v", dbUser, dbPass, dbAddr))
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Create the user store
	users := users.Make(db, redis)
	users.Listen()

	ge := server.GreeterEndpoint{Iface: iface, AuthReader: ar, Users: users}
	ge.Run()
}
