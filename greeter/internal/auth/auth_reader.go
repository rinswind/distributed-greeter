package auth

import (
	"context"
	"fmt"
	"os"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis/v8"
)

var (
	redisCtx    = context.Background()
	redisClient *redis.Client

	atSecret string
	rtSecret string
)

func init() {
	// Init Redis client
	dsn := os.Getenv("REDIS_DSN")
	redisClient = redis.NewClient(&redis.Options{
		Addr: dsn, //redis port
	})
	_, err := redisClient.Ping(redisCtx).Result()
	if err != nil {
		panic(err)
	}

	// Init access token settings
	atSecret = os.Getenv("ACCESS_TOKEN_SECRET")

	// Init refresh token settings
	rtSecret = os.Getenv("REFRESH_TOKEN_SECRET")
}

// GetAuth obtains a cached user login
func GetAuth(tokenStr string) (map[string]interface{}, error) {
	claims, err := decodeToken(tokenStr, atSecret)
	if err != nil {
		return nil, err
	}

	atUUID, ok := claims["access_uuid"].(string)
	if !ok {
		return nil, fmt.Errorf("No %v claim in token", "access_uuid")
	}

	err = redisClient.Get(redisCtx, atUUID).Err()
	if err != nil {
		return nil, err
	}

	return claims, nil
}

func decodeToken(encoded string, key string) (map[string]interface{}, error) {
	token, err := jwt.Parse(encoded, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(key), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("Token invalid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("Bad claims type %T", claims)
	}
	return claims, nil
}
