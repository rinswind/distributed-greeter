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

	router.GET("/greetings", handleAllGreeters)
	router.POST("/greetings/:lang", ge.handleGreeter)

	router.Run(ge.Iface)
}

// GET /greetings
func handleAllGreeters(c *gin.Context) {
	type Languages struct {
		Langs map[string]string `json:"languages"`
	}

	langs := Languages{Langs: make(map[string]string)}
	for ln := range messages.Greeters {
		langs.Langs[ln] = c.Request.RequestURI + "/" + ln
	}

	c.JSON(http.StatusOK, &langs)
}

// POST /greetings/:lang
func (ge *GreeterEndpoint) handleGreeter(c *gin.Context) {
	lang := c.Param("lang")

	authCtx, _ := c.Get(ginauth.ContextKey)
	authClaims := authCtx.(map[string]interface{})
	userID, _ := authClaims["user_id"].(uint64)

	user, err := ge.Users.GetUser(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	msg := messages.Greeters[lang](user.Name)

	type Message struct {
		Language string `json:"language"`
		Message  string `json:"message"`
	}

	message := Message{Language: lang, Message: msg}
	c.JSON(http.StatusOK, &message)
}

// GET /users/:uid
func (ge *GreeterEndpoint) handleUserInfo(c *gin.Context) {
	uidParam := c.Param("uid")
	uid, err := strconv.ParseUint(uidParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v not a valid user ID", uidParam)})
		return
	}

	user, err := ge.Users.GetUser(uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v not a valid user ID", uidParam)})
		return
	}

	type UserInfo struct {
		Language string `json:"user_language"`
	}

	var userInfo UserInfo
	if err := c.ShouldBindJSON(&userInfo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := ge.Users.GetUser(uid)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	user = &users.User{ID: uid, Name: user.Name, Language: userInfo.Language}
	err = ge.Users.UpdateUser(user)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}
