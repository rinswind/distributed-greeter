package users

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/go-redis/redis/v8"
)

const (
	usersChannel = "/users"
)

// User models a user
type User struct {
	ID       uint64
	Name     string
	Password string
}

// Store is a User store
type Store struct {
	db    *sql.DB
	redis *redis.Client
}

// Make creates a Store client
func Make(db *sql.DB, redis *redis.Client) *Store {
	return &Store{db: db, redis: redis}
}

// CreateUser adds a new user
func (s *Store) CreateUser(name, pass string) (uint64, error) {
	// if _, ok := s.byName[name]; ok {
	// 	return 0, fmt.Errorf("User %v not available", name)
	// }

	var err error

	tx, err := s.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("Failed to create user %v: %v", name, err)
	}

	// TODO Find a way to get the ID atomically with the INSERT
	_, err = tx.Exec("INSERT INTO users (name, password) VALUES (?, ?)", name, pass)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return 0, fmt.Errorf("Failed to create user %v: %v, rollback also failed: %v", name, err, rollbackErr)
		}
		return 0, fmt.Errorf("Failed to create user %v: %v", name, err)
	}

	var id uint64
	err = tx.QueryRow("SELECT LAST_INSERT_ID() FROM users LIMIT 1;").Scan(&id)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return 0, fmt.Errorf("Failed to create user %v: %v, rollback also failed: %v", name, err, rollbackErr)
		}
		return 0, fmt.Errorf("Failed to create user %v: %v", name, err)
	}

	if commitErr := tx.Commit(); commitErr != nil {
		return 0, fmt.Errorf("Failed to create user %v: %v", name, commitErr)
	}

	// TODO Do this before tx.Commit() and if this fails do tx.Rollback()?
	userEvent := &Event{Type: int(Created), ID: id, Name: name}
	err = s.redis.Publish(context.Background(), usersChannel, userEvent.Marshal()).Err()
	if err != nil {
		return 0, fmt.Errorf("Failed to publish user creation event %v: %v", userEvent, err)
	}

	return id, nil
}

// GetUserByName finds a user
func (s *Store) GetUserByName(name string) (*User, error) {
	var user User
	err := s.db.QueryRow("SELECT id, name, password FROM users WHERE name=?", name).Scan(&user.ID, &user.Name, &user.Password)
	if err != nil {
		return nil, fmt.Errorf("Failed to get user %v: %v", name, err)
	}
	return &user, nil
}

// GetUserByID finds a user
func (s *Store) GetUserByID(id uint64) (*User, error) {
	var user User
	err := s.db.QueryRow("SELECT id, name, password FROM users WHERE id=?", id).Scan(&user.ID, &user.Name, &user.Password)
	if err != nil {
		return nil, fmt.Errorf("Failed to get user %v: %v", id, err)
	}
	return &user, nil
}

// DeleteUserByID deletes a used by ID
func (s *Store) DeleteUserByID(id uint64) error {
	// user, ok := s.byID[id]
	// if !ok {
	// 	return fmt.Errorf("User %v not available", id)
	// }

	var err error

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("Failed to delete user %v: %v", id, err)
	}

	// TODO "SELECT FOR UPDATE"
	var user User
	err = tx.QueryRow("SELECT id, name, password FROM users WHERE id=?", id).Scan(&user.ID, &user.Name, &user.Password)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("Failed to find user %v: %v, unable to rollback: %v", id, err, rollbackErr)
		}
		return fmt.Errorf("Failed to find user %v: %v", id, err)
	}

	_, err = tx.Exec("DELETE FROM users WHERE id=?", user.ID)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("Failed to delete user: %v, unable to rollback: %v", err, rollbackErr)
		}
		return err
	}

	if commitErr := tx.Commit(); commitErr != nil {
		return fmt.Errorf("Failed to delete user %v: %v", id, commitErr)
	}

	// TODO Do this before tx.Commit() and if this fails do tx.Rollback()?
	userEvent := Event{Type: int(Deleted), ID: user.ID, Name: user.Name}
	err = s.redis.Publish(context.Background(), usersChannel, userEvent.Marshal()).Err()
	if err != nil {
		return fmt.Errorf("Failed to publish user deletion event: %v", err)
	}

	return nil
}
