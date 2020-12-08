package server

import (
	"net/http"

	"example.org/services/greeter/internal/auth"
	"example.org/services/greeter/internal/messages"
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

	authCtx, ok := c.Get(auth.ContextKey)
	if !ok {
		c.Status(http.StatusUnauthorized)
		return
	}

	authClaims := authCtx.(map[string]interface{})
	user, _ := authClaims["user_name"].(string)

	msg := messages.Greeters[lang](user)

	type Message struct {
		Language string `json:"language"`
		Message  string `json:"message"`
	}

	message := Message{Language: lang, Message: msg}

	c.JSON(http.StatusOK, &message)
}
