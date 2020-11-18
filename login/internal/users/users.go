package users

import "errors"

// User models a user
type User struct {
	Name string `json:"user"`
	Pass string `json:"password"`
}

// UserDB is the users database
type UserDB struct {
	store map[string]User
}

// MakeUserDB creates a user database
func MakeUserDB() *UserDB {
	return &UserDB{store: make(map[string]User)}
}

// CreateUser adds a new user
func (udb *UserDB) CreateUser(name, pass string) error {
	if _, ok := udb.store[name]; ok {
		return errors.New("User " + name + " already exists")
	}
	udb.store[name] = User{Name: name, Pass: pass}
	return nil
}

// GetUser finds a user
func (udb *UserDB) GetUser(name string) (*User, error) {
	user, ok := udb.store[name]
	if !ok {
		return nil, errors.New("User " + name + " not available")
	}
	return &user, nil
}
