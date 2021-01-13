package users

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/rinswind/distributed-greeter/greeter/internal/messages"
)

const (
	usersChannel = "/users"
)

// User models a user
type User struct {
	ID       uint64
	Name     string
	Language string
}

// Store is a user preferences store
type Store struct {
	redis *redis.Client
	lock  *sync.Mutex
	byID  map[uint64]*User
}

// Make create a new Store
func Make(redis *redis.Client) *Store {
	return &Store{
		redis: redis,
		lock:  &sync.Mutex{},
		byID:  make(map[uint64]*User),
	}
}

// Listen starts listening for User events
func (s *Store) Listen() error {
	userEvents := s.redis.Subscribe(context.Background(), usersChannel)
	_, err := userEvents.Receive(context.Background())
	if err != nil {
		return err
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
				err = s.CreateUser(event.ID, event.Name, messages.DefaultLanguage)
			case int(Deleted):
				err = s.DeleteUser(event.ID)
			}

			if err != nil {
				log.Printf("Failed to process user event %v: %v", event, err)
			}
		}
	}()

	return nil
}

// CreateUser adds a new user
func (s *Store) CreateUser(id uint64, name string, lang string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if _, ok := s.byID[id]; ok {
		return fmt.Errorf("User %v not available", id)
	}

	user := &User{ID: id, Name: name, Language: lang}
	s.byID[user.ID] = user
	return nil
}

// GetUser finds a user
func (s *Store) GetUser(id uint64) (*User, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	user, ok := s.byID[id]
	if !ok {
		return nil, fmt.Errorf("User %v not available", id)
	}
	return user, nil
}

// UpdateUser updates a user record
func (s *Store) UpdateUser(newUser *User) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if _, ok := s.byID[newUser.ID]; !ok {
		return fmt.Errorf("User %v not available", newUser.ID)
	}

	s.byID[newUser.ID] = newUser
	return nil
}

// DeleteUser deletes a used by ID
func (s *Store) DeleteUser(id uint64) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if _, ok := s.byID[id]; !ok {
		return fmt.Errorf("User %v not available", id)
	}

	delete(s.byID, id)
	return nil
}
