package main

import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/rinswind/auth-go/tokens"
	_ "github.com/rinswind/azure-msi"
	"github.com/rinswind/distributed-greeter/login/internal/config"
	"github.com/rinswind/distributed-greeter/login/internal/server"
	"github.com/rinswind/distributed-greeter/login/internal/users"
)

func main() {
	// Setup logging
	log.SetFlags(log.Ldate | log.Ltime | log.Lmsgprefix | log.Lshortfile)

	cfg := config.ReadConfig()

	var err error

	// Create the Redis client
	log.Printf("Resolved Redis endpoint: %v", cfg.Redis.Endpoint)
	redisOpts := redis.Options{
		Addr:     cfg.Redis.Endpoint,
		Username: cfg.Redis.User,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.Db,
	}
	if cfg.Redis.TLS {
		redisOpts.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}
	redis := redis.NewClient(&redisOpts)
	_, err = redis.Ping(context.Background()).Result()
	check(err)
	defer redis.Close()

	// Create the DB client
	db, err := sql.Open("mysqlMsi", cfg.Db.Endpoint)
	check(err)
	defer db.Close()

	// Create and init the DB
	users := users.Make(db, redis)
	err = users.Init()
	check(err)

	// Create the Auth token handlers
	authReader := tokens.AuthReader{
		Redis:    redis,
		ATSecret: cfg.AccessToken.AccessTokenSecret,
		RTSecret: cfg.AccessToken.RefreshTokenSecret}

	authWriter := tokens.AuthWriter{
		Redis:    redis,
		ATSecret: cfg.AccessToken.AccessTokenSecret,
		ATExpiry: time.Minute * time.Duration(cfg.AccessToken.AccessTokenExpiry),
		RTSecret: cfg.AccessToken.RefreshTokenSecret,
		RTExpiry: time.Minute * time.Duration(cfg.AccessToken.AccessTokenExpiry)}

	// Create and run the REST endpoint
	iface := fmt.Sprintf(":%v", cfg.Http.Port)
	le := server.LoginEndpoint{
		Iface:      iface,
		AuthReader: &authReader,
		AuthWriter: &authWriter,
		Users:      users,
	}
	le.Run()
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
