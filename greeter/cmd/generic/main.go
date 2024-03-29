package main

import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	"github.com/rinswind/auth-go/tokens"
	"github.com/rinswind/distributed-greeter/greeter/internal/config"
	"github.com/rinswind/distributed-greeter/greeter/internal/server"
	"github.com/rinswind/distributed-greeter/greeter/internal/users"
)

func main() {
	// Setup logging
	log.SetFlags(log.Ldate | log.Ltime | log.Lmsgprefix | log.Lshortfile)

	var err error
	cfg := config.ReadConfig()

	// Create the Redis client
	log.Printf("Resolved Redis endpoint: %v", cfg.Redis.Endpoint)
	redisOpts, err := redis.ParseURL(cfg.Redis.Dsn)
	check(err)
	if cfg.Redis.TLS {
		redisOpts.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}

	redis := redis.NewClient(redisOpts)
	_, err = redis.Ping(context.Background()).Result()
	check(err)
	defer redis.Close()

	// Create the DB client
	log.Printf("Resolved MySQL endpoint: %v", cfg.Db.Endpoint)
	db, err := sql.Open(cfg.Db.Driver, cfg.Db.Dsn)
	check(err)
	defer db.Close()

	// Create and init the Users store
	users := users.Make(db, redis)
	err = users.Init()
	check(err)
	users.Listen()

	// Create the auth session manager
	authReader := &tokens.AuthReader{
		Redis:    redis,
		ATSecret: cfg.AccessToken.AccessTokenSecret,
		RTSecret: cfg.AccessToken.RefreshTokenSecret}

	// Create and run the greeter endpoint
	iface := fmt.Sprintf(":%v", cfg.Http.Port)
	log.Printf("Resolved HTTP server endpoint: %v", iface)
	greeterEndpoint := server.GreeterEndpoint{
		Iface:      iface,
		AuthReader: authReader,
		Users:      users}
	greeterEndpoint.Run()
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
