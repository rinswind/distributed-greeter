package users

import (
	"fmt"
	"sync"
)

// User models a user
type User struct {
	ID   uint64
	Name string
	Pass string
}

var (
	lock   *sync.Mutex
	nextID uint64
	byName map[string]*User
	byID   map[uint64]*User
)

func init() {
	lock = &sync.Mutex{}
	nextID = 0
	byName = make(map[string]*User)
	byID = make(map[uint64]*User)
}

// CreateUser adds a new user
func CreateUser(name, pass string) (uint64, error) {
	lock.Lock()
	defer lock.Unlock()

	if _, ok := byName[name]; ok {
		return 0, fmt.Errorf("User %v not available", name)
	}

	user := &User{ID: nextID, Name: name, Pass: pass}
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
