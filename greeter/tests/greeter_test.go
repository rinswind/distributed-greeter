package tests

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

var (
	ingress        = "http://10.64.140.43"
	loginService   = ingress + "/greeter/auth"
	greeterService = ingress + "/greeter/messages"
)

func TestAuthzSplit(t *testing.T) {
	loginUser := "tobo"
	loginPass := "obot"

	//
	// Login
	//
	loginServiceUsers := loginService + "/users"
	loginServiceLogin := loginService + "/login"

	type LoginRequest struct {
		User     string `json:"user_name"`
		Password string `json:"user_password"`
	}

	user := &LoginRequest{User: loginUser, Password: loginPass}
	userStr, _ := json.Marshal(user)

	_, err := http.Post(loginServiceUsers, "application/json", strings.NewReader(string(userStr)))
	if err != nil {
		t.Fatal(err)
	}

	type LoginInfo struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}

	resp, err := http.Post(loginServiceLogin, "application/json", strings.NewReader(string(userStr)))
	if err != nil {
		t.Fatal(err)
	}

	loginInfo := &LoginInfo{}
	readJSON(t, resp.Body, loginInfo)

	//
	// Greet
	//
	greeterServiceLangs := greeterService + "/greetings"

	req, _ := http.NewRequest("GET", greeterServiceLangs, nil)
	req.Header.Add("Authorization", "Bearer "+base64.StdEncoding.EncodeToString([]byte(loginInfo.AccessToken)))
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	type Languages struct {
		Langs map[string]string `json:"languages"`
	}
	langJSON := Languages{}
	readJSON(t, resp.Body, &langJSON)

	for _, path := range langJSON.Langs {
		langService := greeterService + path
		req, _ = http.NewRequest("GET", langService, nil)
		req.Header.Add("Authorization", "Bearer "+base64.StdEncoding.EncodeToString([]byte(loginInfo.AccessToken)))
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		type Message struct {
			Lang    string `json:"language"`
			Message string `json:"message"`
		}

		msgJSON := Message{}
		readJSON(t, resp.Body, &msgJSON)

		fmt.Println(msgJSON)
	}

	//
	// Logout
	//
	loginServiceLogout := loginService + "/logout"

	type LogoutInfo struct {
		AccessToken string `json:"access_token"`
	}

	logout := &LogoutInfo{AccessToken: loginInfo.AccessToken}
	logoutStr, _ := json.Marshal(logout)

	_, err = http.Post(loginServiceLogout, "application/json", strings.NewReader(string(logoutStr)))
	if err != nil {
		t.Fatal(err)
	}

	req, _ = http.NewRequest("GET", greeterServiceLangs, nil)
	req.Header.Add("Authorization", "Bearer "+base64.StdEncoding.EncodeToString([]byte(loginInfo.AccessToken)))
	resp, _ = http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatal(resp)
	}
}

func readJSON(t *testing.T, in io.Reader, out interface{}) {
	jsonStr, err := ioutil.ReadAll(in)
	if err != nil {
		t.Fatal(err)
	}

	err = json.Unmarshal(jsonStr, out)
	if err != nil {
		t.Fatal(string(jsonStr), err)
	}
}
