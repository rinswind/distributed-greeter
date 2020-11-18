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

func TestAuthzSplit(t *testing.T) {
	//
	// Login
	//
	loginService := "http://localhost:8080"
	loginServiceUsers := loginService + "/users"
	loginServiceLogin := loginService + "/login"

	type User struct {
		User     string `json:"user"`
		Password string `json:"password"`
	}

	user := User{User: "tobo", Password: "obot"}

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

	// key := func(t *jwt.Token) (interface{}, error) { return []byte("secret"), nil }

	// token, err := jwt.Parse(tokenJSON.Token, key)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// fmt.Println("jwt claims: ", token.Claims)

	//
	// Greet
	//
	greeterService := "http://localhost:8090"
	greeterServiceLangs := greeterService + "/greeters"

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
