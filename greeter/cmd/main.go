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
	_ "github.com/go-sql-driver/mysql"
	"github.com/rinswind/auth-go/tokens"
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

	// Init access token settings
	var accessCreds struct {
		AccessTokenSecret  string `json:"accessTokenSecret"`
		RefreshTokenSecret string `json:"refreshTokenSecret"`
	}
	authCredsFile := os.Getenv("AUTH_TOKEN_CREDS")
	err = readJsonFile(authCredsFile, &accessCreds)
	check(err)

	// Create the auth session manager
	ar := &tokens.AuthReader{Redis: redis, ATSecret: accessCreds.AccessTokenSecret, RTSecret: accessCreds.RefreshTokenSecret}

	// Read the DB client config
	dbAddr := os.Getenv("DB_ADDR")

	var dbCreds struct {
		User     string `json:"user"`
		Password string `json:"password"`
	}
	dbCredsFile := os.Getenv("DB_CREDS")
	err = readJsonFile(dbCredsFile, &dbCreds)
	check(err)

	// Create the DB client
	db, err := sql.Open("mysql", fmt.Sprintf("%v:%v@%v", dbCreds.User, dbCreds.Password, dbAddr))
	check(err)
	defer db.Close()

	// Create and init the DB
	users := users.Make(db, redis)
	err = users.Init()
	check(err)
	users.Listen()

	ge := server.GreeterEndpoint{Iface: iface, AuthReader: ar, Users: users}
	ge.Run()
}
