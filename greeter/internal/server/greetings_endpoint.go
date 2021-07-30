package server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	ginauth "github.com/rinswind/auth-go/gin"
	"github.com/rinswind/auth-go/tokens"
	"github.com/rinswind/distributed-greeter/greeter/internal/messages"
	"github.com/rinswind/distributed-greeter/greeter/internal/users"
)

// GreeterEndpoint is the greeter REST endpoint
type GreeterEndpoint struct {
	Iface      string
	AuthReader *tokens.AuthReader
	Users      *users.Store
}

// Run starts the rest endpoint
func (ge *GreeterEndpoint) Run() {
	router := gin.Default()
	authHandler := ginauth.MakeHandler(ge.AuthReader)
	router.Use(gin.HandlerFunc(authHandler))

	router.GET("/users/:uid", ge.handleUserInfo)
	router.PUT("/users/:uid", ge.handleUserUpdate)

	router.GET("/greetings", handleGreetingLangs)
	router.POST("/greetings", ge.handleGreeting)

	router.Run(ge.Iface)
}

// GET /greetings
func handleGreetingLangs(c *gin.Context) {
	type Languages struct {
		Langs map[string]string `json:"languages"`
	}

	langs := Languages{Langs: make(map[string]string)}
	for ln := range messages.Greeters {
		langs.Langs[ln] = c.Request.RequestURI + "/" + ln
	}

	c.JSON(http.StatusOK, &langs)
}

// POST /greetings
func (ge *GreeterEndpoint) handleGreeting(c *gin.Context) {
	type MessageRequest struct {
		ID       uint64 `json:"user_id"`
		Language string `json:"language"`
	}

	var msgReq MessageRequest
	if err := c.ShouldBindJSON(&msgReq); err != nil {
		c.Error(err)
		// Report full details when the REST API contract is violated
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	user, err := ge.Users.GetUser(msgReq.ID)
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Errorf("Failed to find user %v", msgReq.ID)})
		return
	}

	msg := messages.Greeters[msgReq.Language](user.Name)

	type Message struct {
		ID       uint64 `json:"user_id"`
		Language string `json:"language"`
		Message  string `json:"message"`
	}

	message := Message{ID: msgReq.ID, Language: msgReq.Language, Message: msg}
	c.JSON(http.StatusOK, &message)
}

// GET /users/:uid
func (ge *GreeterEndpoint) handleUserInfo(c *gin.Context) {
	uidParam := c.Param("uid")
	uid, err := strconv.ParseUint(uidParam, 10, 64)
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Errorf("%v not a valid user ID", uidParam)})
		return
	}

	user, err := ge.Users.GetUser(uid)
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Errorf("Failed to find user %v", uid)})
		return
	}

	type UserInfo struct {
		ID       uint64 `json:"user_id"`
		Name     string `json:"user_name"`
		Language string `json:"user_language"`
	}

	userInfo := UserInfo{ID: user.ID, Name: user.Name, Language: user.Language}
	c.JSON(http.StatusOK, &userInfo)
}

// PUT /users/:uid
func (ge *GreeterEndpoint) handleUserUpdate(c *gin.Context) {
	uidParam := c.Param("uid")
	uid, err := strconv.ParseUint(uidParam, 10, 64)
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Errorf("%v not a valid user ID", uidParam)})
		return
	}

	type UserInfo struct {
		Language string `json:"user_language"`
	}

	var userInfo UserInfo
	if err := c.ShouldBindJSON(&userInfo); err != nil {
		c.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := ge.Users.GetUser(uid)
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Errorf("Failed to find user %v", uid)})
		return
	}

	user = &users.User{ID: uid, Name: user.Name, Language: userInfo.Language}
	err = ge.Users.UpdateUser(user)
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Errorf("Failed to update user %v", user)})
		return
	}

	c.Status(http.StatusOK)
}
