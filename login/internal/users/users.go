package users

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/go-redis/redis/v8"
)

// User models a user
type User struct {
	ID   uint64
	Name string
	Pass string
}

const (
	usersChannel = "/users"
)

var (
	lock   *sync.Mutex      = &sync.Mutex{}
	nextID uint64           = 0
	byName map[string]*User = make(map[string]*User)
	byID   map[uint64]*User = make(map[uint64]*User)

	redisCtx    = context.Background()
	redisClient *redis.Client
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
}

// CreateUser adds a new user
func CreateUser(name, pass string) (uint64, error) {
	lock.Lock()
	defer lock.Unlock()

	if _, ok := byName[name]; ok {
		return 0, fmt.Errorf("User %v not available", name)
	}

	user := &User{ID: nextID, Name: name, Pass: pass}

	// Notify everyone first
	// TODO Instead post a task to a work queue to retry the delivery until success
	userEvent := Event{Type: int(Created), ID: user.ID, Name: user.Name}
	err := redisClient.Publish(redisCtx, usersChannel, userEvent.Marshal()).Err()
	if err != nil {
		return 0, fmt.Errorf("Failed to publish user creation event: %v", err)
	}

	// One everyone knows modify the local persistence, which can't fail at present
	byName[name] = user
	byID[user.ID] = user

	nextID++
	return user.ID, nil
}

// GetUserByName finds a user
func GetUserByName(name string) (*User, error) {
	lock.Lock()
	defer lock.Unlock()

	user, ok := byName[name]
	if !ok {
		return nil, fmt.Errorf("User %v not available", name)
	}
	return user, nil
}

// GetUserByID finds a user
func GetUserByID(id uint64) (*User, error) {
	lock.Lock()
	defer lock.Unlock()

	user, ok := byID[id]
	if !ok {
		return nil, fmt.Errorf("User %v not available", id)
	}
	return user, nil
}

// ListUserIDs lists all user IDs
func ListUserIDs() *[]uint64 {
	lock.Lock()
	defer lock.Unlock()

	ids := make([]uint64, len(byID))
	i := 0
	for id := range byID {
		ids[i] = id
		i++
	}

	return &ids
}

// DeleteUserByID deletes a used by ID
func DeleteUserByID(id uint64) error {
	lock.Lock()
	defer lock.Unlock()

	user, ok := byID[id]
	if !ok {
		return fmt.Errorf("User %v not available", id)
	}

	// Notify everyone first
	// TODO Instead post a task to a work queue to retry the delivery until success
	userEvent := Event{Type: int(Deleted), ID: user.ID, Name: user.Name}
	err := redisClient.Publish(redisCtx, usersChannel, userEvent.Marshal()).Err()
	if err != nil {
		return fmt.Errorf("Failed to publish user deletion event: %v", err)
	}

	// One everyone knows modify the local persistence, which can't fail at present
	delete(byID, id)
	delete(byName, user.Name)
	return nil
}
