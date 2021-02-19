package users

import (
	"context"
	"database/sql"
	"log"

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
	db    *sql.DB
	redis *redis.Client
}

// Make create a new Store
func Make(db *sql.DB, redis *redis.Client) *Store {
	return &Store{db: db, redis: redis}
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
				user := &User{ID: event.ID, Name: event.Name, Language: messages.DefaultLanguage}
				err = s.CreateUser(user)
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
func (s *Store) CreateUser(newUser *User) error {
	_, err := s.db.Exec("INSERT INTO users (id, name, language) VALUES (?, ?, ?)", newUser.ID, newUser.Name, newUser.Language)
	return err
}

// GetUser finds a user
func (s *Store) GetUser(id uint64) (*User, error) {
	var user User
	err := s.db.QueryRow("SELECT id, name, language FROM users WHERE id=?", id).Scan(&user.ID, &user.Name, &user.Language)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser updates a user record
func (s *Store) UpdateUser(newUser *User) error {
	_, err := s.db.Exec("UPDATE users SET name=?, language=? WHERE id=?", newUser.Name, newUser.Language, newUser.ID)
	return err
}

// DeleteUser deletes a used by ID
func (s *Store) DeleteUser(id uint64) error {
	_, err := s.db.Exec("DELETE FROM users WHERE id=?", id)
	return err
}
