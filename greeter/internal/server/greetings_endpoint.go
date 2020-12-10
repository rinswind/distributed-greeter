package server

import (
	"net/http"

	"example.org/services/greeter/internal/auth"
	"example.org/services/greeter/internal/messages"
	"example.org/services/greeter/internal/users"
	"github.com/gin-gonic/gin"
)

var (
	router = gin.Default()
)

// MakeGreeterEndpoint creates a router for the greeter REST endpoint
func init() {
	router.Use(gin.HandlerFunc(auth.Handler))
	router.GET("/greetings", handleAllGreeters)
	router.GET("/greetings/:lang", handleGreeter)
	router.GET("/user", handleUserInfo)
	router.POST("/user", handleUserUpdate)
}

// Run starts the rest endpoint
func Run(iface string) {
	router.Run(iface)
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

// GET /greetings/:lang
func handleGreeter(c *gin.Context) {
	lang := c.Param("lang")

	authCtx, _ := c.Get(auth.ContextKey)
	authClaims := authCtx.(map[string]interface{})
	userID, _ := authClaims["user_id"].(uint64)

	user, err := users.GetUser(userID)
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

// GET /user
func handleUserInfo(c *gin.Context) {
	authCtx, _ := c.Get(auth.ContextKey)
	authClaims := authCtx.(map[string]interface{})
	userID, _ := authClaims["user_id"].(uint64)

	user, err := users.GetUser(userID)
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

// POST /user
func handleUserUpdate(c *gin.Context) {
	type UserInfo struct {
		Language string `json:"user_language"`
	}

	var userInfo UserInfo
	if err := c.ShouldBindJSON(&userInfo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	authCtx, _ := c.Get(auth.ContextKey)
	authClaims := authCtx.(map[string]interface{})
	userID, _ := authClaims["user_id"].(uint64)

	user, err := users.GetUser(userID)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	user = &users.User{ID: userID, Name: user.Name, Language: userInfo.Language}
	err = users.UpdateUser(user)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}
