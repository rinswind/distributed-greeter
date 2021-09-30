package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/rinswind/auth-go/tokens"
	_ "github.com/rinswind/azure-msi"
	"github.com/rinswind/distributed-greeter/login/internal/server"
	"github.com/rinswind/distributed-greeter/login/internal/users"
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

	// Read the HTTP port
	port := os.Getenv("HTTP_PORT")
	iface := fmt.Sprintf(":%v", port)

	// Init Redis client
	redisDsn := os.Getenv("REDIS_ADDR")

	redisCredsFile := os.Getenv("REDIS_CREDS")
	var redisCreds struct {
		AccessKey string `json:"accessKey"`
	}
	readJsonFile(redisCredsFile, &redisCreds)

	redis := redis.NewClient(&redis.Options{
		Addr:     redisDsn,
		Password: redisCreds.AccessKey,
	})
	_, err := redis.Ping(context.Background()).Result()
	check(err)
	defer redis.Close()

	// Read the auth-token credentials
	var accessCreds struct {
		AccessTokenSecret string `json:"accessTokenSecret"`
		AccessTokenExpiry int    `json:"accessTokenExpiry"`

		RefreshTokenSecret string `json:"refreshTokenSecret"`
		RefreshTokenExpiry int    `json:"refreshTokenExpiry"`
	}
	authCredsFile := os.Getenv("AUTH_TOKEN_CREDS")
	err = readJsonFile(authCredsFile, &accessCreds)
	check(err)

	// Create the DB client
	dbAddr := os.Getenv("DB_ADDR")

	db, err := sql.Open("mysqlMsi", dbAddr)
	check(err)
	defer db.Close()

	// Create and init the DB
	users := users.Make(db, redis)
	err = users.Init()
	check(err)

	// Run the REST endpoint
	le := server.LoginEndpoint{
		Iface: iface,
		AuthReader: &tokens.AuthReader{
			Redis:    redis,
			ATSecret: accessCreds.AccessTokenSecret,
			RTSecret: accessCreds.RefreshTokenSecret},
		AuthWriter: &tokens.AuthWriter{
			Redis:    redis,
			ATSecret: accessCreds.AccessTokenSecret,
			ATExpiry: time.Minute * time.Duration(accessCreds.AccessTokenExpiry),
			RTSecret: accessCreds.RefreshTokenSecret,
			RTExpiry: time.Minute * time.Duration(accessCreds.RefreshTokenExpiry)},
		Users: users,
	}
	le.Run()
}
