package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/rinswind/auth-go/tokens"
	_ "github.com/rinswind/azure-msi"
	"github.com/rinswind/distributed-greeter/greeter/internal/server"
	"github.com/rinswind/distributed-greeter/greeter/internal/users"
)

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func readJsonFile(file string, out interface{}) error {
	raw, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	err = json.Unmarshal(raw, out)
	return err
}

func main() {
	// Setup logging
	log.SetFlags(log.Ldate | log.Ltime | log.Lmsgprefix | log.Lshortfile)

	var err error

	// Read the HTTP port
	port := os.Getenv("HTTP_PORT")
	iface := fmt.Sprintf(":%v", port)

	// Init the Redis client
	redisDsn := os.Getenv("REDIS_ADDR")

	redisCredsFile := os.Getenv("REDIS_CREDS")
	var redisCreds struct {
		AccessKey string `json:"accessKey"`
	}
	err = readJsonFile(redisCredsFile, &redisCreds)
	check(err)

	redis := redis.NewClient(&redis.Options{
		Addr:     redisDsn,
		Password: redisCreds.AccessKey,
	})
	_, err = redis.Ping(context.Background()).Result()
	check(err)
	defer redis.Close()

	// Create the DB client
	dbDsn := os.Getenv("DB_ADDR")

	db, err := sql.Open("mysqlMsi", dbDsn)
	check(err)
	defer db.Close()

	// Create and init the Users store
	users := users.Make(db, redis)
	err = users.Init()
	check(err)
	users.Listen()

	// Create the auth session manager
	var accessCreds struct {
		AccessTokenSecret  string `json:"accessTokenSecret"`
		RefreshTokenSecret string `json:"refreshTokenSecret"`
	}
	authCredsFile := os.Getenv("AUTH_TOKEN_CREDS")
	err = readJsonFile(authCredsFile, &accessCreds)
	check(err)

	authReader := &tokens.AuthReader{
		Redis:    redis,
		ATSecret: accessCreds.AccessTokenSecret,
		RTSecret: accessCreds.RefreshTokenSecret}

	// Create and run the greeter endpoint
	greeterEndpoint := server.GreeterEndpoint{
		Iface:      iface,
		AuthReader: authReader,
		Users:      users}
	greeterEndpoint.Run()
}
