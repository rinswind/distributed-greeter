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
	ingress = "http://10.64.140.43"

	loginService = ingress + "/greeter/auth"
	loginUser    = "tobo"
	loginPass    = "obot"

	greeterService = ingress + "/greeter/messages"
)

func TestAuthzSplit(t *testing.T) {
	//
	// Login
	//
	loginServiceUsers := loginService + "/users"
	loginServiceLogin := loginService + "/login"

	type User struct {
		User     string `json:"user"`
		Password string `json:"password"`
	}

	user := User{User: loginUser, Password: loginPass}

	userStr, _ := json.Marshal(user)

	_, err := http.Post(loginServiceUsers, "application/json", strings.NewReader(string(userStr)))
	if err != nil {
		t.Fatal(err)
	}

	type Token struct {
		Token string `json:"token"`
	}

	resp, err := http.Post(loginServiceLogin, "application/json", strings.NewReader(string(userStr)))
	if err != nil {
		t.Fatal(err)
	}

	tokenJSON := Token{}
	readJSON(t, resp.Body, &tokenJSON)

	//
	// Greet
	//
	greeterServiceLangs := greeterService + "/greetings"

	client := &http.Client{}

	req, err := http.NewRequest("GET", greeterServiceLangs, nil)
	req.Header.Add("Authorization", "Bearer "+base64.StdEncoding.EncodeToString([]byte(tokenJSON.Token)))
	resp, err = client.Do(req)
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
		req, err = http.NewRequest("GET", langService, nil)
		req.Header.Add("Authorization", "Bearer "+base64.StdEncoding.EncodeToString([]byte(tokenJSON.Token)))
		resp, err = client.Do(req)
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
