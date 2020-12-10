package users

import (
	"encoding/json"
)

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

// Unmarshal converts a string to an Event
func (e *Event) Unmarshal(str string) error {
	return json.Unmarshal([]byte(str), e)
}

// StringToEventType parses a string to an EventType
// func StringToEventType(str string) (EventType, error) {
// 	switch str {
// 	case "Created":
// 		return Created, nil
// 	case "Deleted":
// 		return Deleted, nil
// 	}
// 	return -1, fmt.Errorf("Unknown event type %v", str)
// }
