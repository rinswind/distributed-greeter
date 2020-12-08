package tests

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"
)

var (
	ingress      = "http://10.64.140.43"
	loginService = ingress + "/greeter/auth"
)

func TestDeleteUser(t *testing.T) {
	loginServiceUsers := loginService + "/users"

	type UserCreds struct {
		Name     string `json:"user_name"`
		Password string `json:"user_password"`
	}

	type UserInfo struct {
		ID   uint64 `json:"user_id"`
		Name string `json:"user_name"`
	}

	//
	// Create user
	//
	credsStr := writeJSON(t, &UserCreds{Name: fmt.Sprintf("user-%v", time.Now().Format("15:04:05.000")), Password: "pass"})
	resp, err := http.Post(loginServiceUsers, "application/json", strings.NewReader(string(credsStr)))
	if err != nil {
		t.Fatal(err)
	}

	user := &UserInfo{}
	readJSON(t, resp.Body, user)

	userURL := fmt.Sprint(loginServiceUsers, "/", user.ID)
	t.Log("Created user")

	//
	// Get user
	//
	resp, err = http.Get(userURL)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Invalid status '%v %v' on GET %v", resp.StatusCode, resp.Status, userURL)
	}

	t.Log("Obtained user")

	//
	// Delete user
	//
	req, err := http.NewRequest(http.MethodDelete, userURL, nil)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Invalid status '%v %v' on DELETE %v", resp.StatusCode, resp.Status, userURL)
	}
	t.Log("Deleted user")

	//
	// Check user is missing
	//
	resp, err = http.Get(fmt.Sprint(loginServiceUsers, "/", user.ID))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("Invalid status %v on GET %v", http.StatusNotFound, userURL)
	}
	t.Log("Filed to get user")
}

func TestAuthzSplit(t *testing.T) {
	loginServiceUsers := loginService + "/users"

	//
	// Create users
	//
	for i := 0; i < 10; i++ {
		type LoginRequest struct {
			User     string `json:"user_name"`
			Password string `json:"user_password"`
		}

		user := &LoginRequest{User: fmt.Sprintf("user-%v", i), Password: fmt.Sprintf("pass-%v", i)}
		userStr := writeJSON(t, user)

		_, err := http.Post(loginServiceUsers, "application/json", strings.NewReader(string(userStr)))
		if err != nil {
			t.Fatal(err)
		}
	}

	//
	// List user ids
	//
	type Users struct {
		UserIDs []uint64 `json:"user_ids"`
	}

	resp, err := http.Get(loginServiceUsers)
	if err != nil {
		t.Fatal(err)
	}

	userIds := &Users{}
	readJSON(t, resp.Body, userIds)

	//
	// Get back the users
	//
	for id := range userIds.UserIDs {
		type UserInfo struct {
			UserID   uint64 `json:"user_id"`
			UserName string `json:"user_name"`
		}

		resp, err := http.Get(fmt.Sprint(loginServiceUsers, "/", id))
		if err != nil {
			t.Fatal(err)
		}

		user := &UserInfo{}
		readJSON(t, resp.Body, user)
	}
}

func readJSON(t *testing.T, in io.Reader, out interface{}) {
	jsonStr, err := ioutil.ReadAll(in)
	if err != nil {
		t.Logf("read: %v -> %v\n", string(jsonStr), err)
		t.Fatal(err)
	}

	err = json.Unmarshal(jsonStr, out)
	if err != nil {
		t.Logf("read: %v -> %v\n", string(jsonStr), err)
		t.Fatal(string(jsonStr), err)
	}

	t.Logf("read: %v -> %v\n", string(jsonStr), out)
}

func writeJSON(t *testing.T, out interface{}) string {
	jsonStr, err := json.Marshal(out)
	if err != nil {
		t.Logf("write: %v -> %v\n", out, err)
		t.Fatal(err)
	}

	t.Logf("write: %v -> %v\n", out, string(jsonStr))
	return string(jsonStr)
}
