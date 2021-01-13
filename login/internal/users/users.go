package users

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-redis/redis/v8"
)

const (
	usersChannel = "/users"
)

// User models a user
type User struct {
	ID   uint64
	Name string
	Pass string
}

// Store is a User store
type Store struct {
	redis  *redis.Client
	lock   *sync.Mutex
	nextID uint64
	byName map[string]*User
	byID   map[uint64]*User
}

// Make creates a Store client
func Make(redis *redis.Client) *Store {
	return &Store{
		redis:  redis,
		lock:   &sync.Mutex{},
		nextID: 0,
		byName: make(map[string]*User),
		byID:   make(map[uint64]*User),
	}
}

// CreateUser adds a new user
func (s *Store) CreateUser(name, pass string) (uint64, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if _, ok := s.byName[name]; ok {
		return 0, fmt.Errorf("User %v not available", name)
	}

	user := &User{ID: s.nextID, Name: name, Pass: pass}

	// Notify everyone first
	// TODO Instead post a task to a work queue to retry the delivery until success
	userEvent := Event{Type: int(Created), ID: user.ID, Name: user.Name}
	err := s.redis.Publish(context.Background(), usersChannel, userEvent.Marshal()).Err()
	if err != nil {
		return 0, fmt.Errorf("Failed to publish user creation event: %v", err)
	}

	// One everyone knows modify the local persistence, which can't fail at present
	s.byName[name] = user
	s.byID[user.ID] = user

	s.nextID++
	return user.ID, nil
}

// GetUserByName finds a user
func (s *Store) GetUserByName(name string) (*User, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	user, ok := s.byName[name]
	if !ok {
		return nil, fmt.Errorf("User %v not available", name)
	}
	return user, nil
}

// GetUserByID finds a user
func (s *Store) GetUserByID(id uint64) (*User, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	user, ok := s.byID[id]
	if !ok {
		return nil, fmt.Errorf("User %v not available", id)
	}
	return user, nil
}

// ListUserIDs lists all user IDs
func (s *Store) ListUserIDs() *[]uint64 {
	s.lock.Lock()
	defer s.lock.Unlock()

	ids := make([]uint64, len(s.byID))
	i := 0
	for id := range s.byID {
		ids[i] = id
		i++
	}

	return &ids
}

// DeleteUserByID deletes a used by ID
func (s *Store) DeleteUserByID(id uint64) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	user, ok := s.byID[id]
	if !ok {
		return fmt.Errorf("User %v not available", id)
	}

	// Notify everyone first
	// TODO Instead post a task to a work queue to retry the delivery until success
	userEvent := Event{Type: int(Deleted), ID: user.ID, Name: user.Name}
	err := s.redis.Publish(context.Background(), usersChannel, userEvent.Marshal()).Err()
	if err != nil {
		return fmt.Errorf("Failed to publish user deletion event: %v", err)
	}

	// One everyone knows modify the local persistence, which can't fail at present
	delete(s.byID, id)
	delete(s.byName, user.Name)
	return nil
}
