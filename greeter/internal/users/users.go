package users

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"example.org/services/greeter/internal/messages"
	"github.com/go-redis/redis/v8"
)

// User models a user
type User struct {
	ID       uint64
	Name     string
	Language string
}

const (
	usersChannel = "/users"
)

var (
	lock *sync.Mutex      = &sync.Mutex{}
	byID map[uint64]*User = make(map[uint64]*User)

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

	// Start listening for user events
	userEvents := redisClient.Subscribe(redisCtx, usersChannel)
	_, err = userEvents.Receive(redisCtx)
	if err != nil {
		panic(err)
	}

	go func() {
		for msg := range userEvents.Channel() {
			event := &Event{}
			err := event.Unmarshal(msg.Payload)
			if err != nil {
				log.Print(err)
				return
			}

			log.Printf("User event: %v", event)

			switch event.Type {
			case int(Created):
				err = CreateUser(event.ID, event.Name, messages.DefaultLanguage)
			case int(Deleted):
				err = DeleteUser(event.ID)
			}

			if err != nil {
				log.Printf("Failed to process user event %v: %v", event, err)
			}
		}
	}()
}

// CreateUser adds a new user
func CreateUser(id uint64, name string, lang string) error {
	lock.Lock()
	defer lock.Unlock()

	if _, ok := byID[id]; ok {
		return fmt.Errorf("User %v not available", id)
	}

	user := &User{ID: id, Name: name, Language: lang}
	byID[user.ID] = user
	return nil
}

// GetUser finds a user
func GetUser(id uint64) (*User, error) {
	lock.Lock()
	defer lock.Unlock()

	user, ok := byID[id]
	if !ok {
		return nil, fmt.Errorf("User %v not available", id)
	}
	return user, nil
}

// UpdateUser updates a user record
func UpdateUser(newUser *User) error {
	lock.Lock()
	defer lock.Unlock()

	if _, ok := byID[newUser.ID]; !ok {
		return fmt.Errorf("User %v not available", newUser.ID)
	}

	byID[newUser.ID] = newUser
	return nil
}

// DeleteUser deletes a used by ID
func DeleteUser(id uint64) error {
	lock.Lock()
	defer lock.Unlock()

	if _, ok := byID[id]; !ok {
		return fmt.Errorf("User %v not available", id)
	}

	delete(byID, id)
	return nil
}
