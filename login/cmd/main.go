package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
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
		log.Fatal(err)
	}
	defer redis.Close()

	// Init access token settings
	atSecret := os.Getenv("ACCESS_TOKEN_SECRET")
	atExp, err := strconv.Atoi(os.Getenv("ACCESS_TOKEN_EXPIRY"))
	if err != nil {
		log.Fatal(err)
	}
	atExpiry := time.Minute * time.Duration(atExp)

	// Init refresh token settings
	rtSecret := os.Getenv("REFRESH_TOKEN_SECRET")
	rtExp, err := strconv.Atoi(os.Getenv("REFRESH_TOKEN_EXPIRY"))
	if err != nil {
		log.Fatal(err)
	}
	rtExpiry := time.Minute * time.Duration(rtExp)

	// Create the DB client
	dbAddr := os.Getenv("DB_ADDR")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	db, err := sql.Open("mysql", fmt.Sprintf("%v:%v@%v", dbUser, dbPass, dbAddr))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create and init the database
	users := users.Make(db, redis)
	err = users.Init()
	if err != nil {
		log.Fatal(err)
	}

	// Run the REST endpoint
	le := server.LoginEndpoint{
		Iface:      iface,
		AuthReader: &tokens.AuthReader{Redis: redis, ATSecret: atSecret, RTSecret: rtSecret},
		AuthWriter: &tokens.AuthWriter{Redis: redis, ATSecret: atSecret, ATExpiry: atExpiry, RTSecret: rtSecret, RTExpiry: rtExpiry},
		Users:      users,
	}
	le.Run()
}
