package auth

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis/v8"
	uuid "github.com/satori/go.uuid"
)

// TokenDetails represents info
type TokenDetails struct {
	userID uint64

	// Access token
	AccessToken   string
	AccessUUID    string
	AccessExpires int64

	// Refresh token
	RefreshToken   string
	RefreshUUID    string
	RefreshExpires int64
}

var (
	redisCtx    = context.Background()
	redisClient *redis.Client

	atSecret string
	atExpiry time.Duration

	rtSecret string
	rtExpiry time.Duration
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
	atExp, err := strconv.Atoi(os.Getenv("ACCESS_TOKEN_EXPIRY"))
	if err != nil {
		panic(err)
	}
	atExpiry = time.Minute * time.Duration(atExp)

	// Init refresh token settings
	rtSecret = os.Getenv("REFRESH_TOKEN_SECRET")
	rtExp, err := strconv.Atoi(os.Getenv("REFRESH_TOKEN_EXPIRY"))
	if err != nil {
		panic(err)
	}
	rtExpiry = time.Minute * time.Duration(rtExp)
}

// CreateToken makes a new token for a given user
func CreateToken(userID uint64) (*TokenDetails, error) {
	td := &TokenDetails{userID: userID}

	var err error

	// Creating Access Token
	td.AccessExpires = time.Now().Add(atExpiry).Unix()
	td.AccessUUID = uuid.NewV4().String()

	atClaims := jwt.MapClaims{}
	atClaims["access_uuid"] = td.AccessUUID
	atClaims["user_id"] = userID
	atClaims["exp"] = td.AccessExpires
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	td.AccessToken, err = at.SignedString([]byte(atSecret))
	if err != nil {
		return nil, err
	}

	// Creating Refresh Token
	td.RefreshExpires = time.Now().Add(rtExpiry).Unix()
	td.RefreshUUID = uuid.NewV4().String()

	rtClaims := jwt.MapClaims{}
	rtClaims["refresh_uuid"] = td.RefreshUUID
	rtClaims["user_id"] = userID
	rtClaims["exp"] = td.RefreshExpires
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	td.RefreshToken, err = rt.SignedString([]byte(rtSecret))
	if err != nil {
		return nil, err
	}

	return td, nil
}

// CreateAuth records the login details
func CreateAuth(td *TokenDetails) error {
	at := time.Unix(td.AccessExpires, 0) //converting Unix to UTC(to Time object)
	rt := time.Unix(td.RefreshExpires, 0)
	now := time.Now()

	userIDStr := strconv.FormatUint(td.userID, 10)

	var err error

	err = redisClient.Set(redisCtx, td.AccessUUID, userIDStr, at.Sub(now)).Err()
	if err != nil {
		return err
	}
	err = redisClient.Set(redisCtx, td.RefreshUUID, userIDStr, rt.Sub(now)).Err()
	if err != nil {
		return err
	}
	return nil
}

// DeleteAuth destroys a users's login
// TODO Require the "access_uuid" instead? This API models Authentication not as a JWT token, but rather as a map of claims
func DeleteAuth(claims map[string]interface{}) (uint64, error) {
	atUUID, ok := claims["access_uuid"].(string)
	if !ok {
		return 0, fmt.Errorf("No %v claim in token", "access_uuid")
	}

	userID, err := redisClient.Del(redisCtx, atUUID).Uint64()
	if err != nil {
		return 0, err
	}
	return userID, nil
}

// DecodeToken obtains a cached user login
func DecodeToken(tokenStr string) (map[string]interface{}, error) {
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
