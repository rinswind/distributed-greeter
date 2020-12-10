package users

import "encoding/json"

// EventType is the type of user events
type EventType int

const (
	// Created user created event
	Created EventType = iota
	// Deleted user delted event
	Deleted
)

// Event models changes to the users db
type Event struct {
	Type int    `json:"type"`
	ID   uint64 `json:"user_id"`
	Name string `json:"user_name"`
}

// func (e EventType) String() string {
// 	return [...]string{"Created", "Deleted"}[e]
// }

// Marshal converts the Event to string
func (e *Event) Marshal() string {
	res, err := json.Marshal(e)
	if err != nil {
		// No reason for JSON marshalling to fail
		panic(err)
	}
	return string(res)
}
