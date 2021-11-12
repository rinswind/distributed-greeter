package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/rinswind/auth-go/tokens"
	_ "github.com/rinswind/azure-msi"
	"github.com/rinswind/distributed-greeter/greeter/internal/config"
	"github.com/rinswind/distributed-greeter/greeter/internal/server"
	"github.com/rinswind/distributed-greeter/greeter/internal/users"
)

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	// Setup logging
	log.SetFlags(log.Ldate | log.Ltime | log.Lmsgprefix | log.Lshortfile)

	var err error
	config := config.ReadConfig()

	redis := redis.NewClient(&redis.Options{
		Addr:     config.Redis.Endpoint,
		Password: config.Redis.AccessKey,
	})
	_, err = redis.Ping(context.Background()).Result()
	check(err)
	defer redis.Close()

	// Create the DB client
	db, err := sql.Open("mysqlMsi", config.Db.Endpoint)
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
		ATSecret: config.AccessToken.AccessTokenSecret,
		RTSecret: config.AccessToken.RefreshTokenSecret}

	// Create and run the greeter endpoint
	iface := fmt.Sprintf(":%v", config.Http.Port)
	greeterEndpoint := server.GreeterEndpoint{
		Iface:      iface,
		AuthReader: authReader,
		Users:      users}
	greeterEndpoint.Run()
}
