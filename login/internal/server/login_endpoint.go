package server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	ginauth "github.com/rinswind/auth-go/gin"
	"github.com/rinswind/auth-go/tokens"
	"github.com/rinswind/distributed-greeter/login/internal/users"
)

// LoginEndpoint is the REST endpoint for the login service
type LoginEndpoint struct {
	Iface      string
	AuthReader *tokens.AuthReader
	AuthWriter *tokens.AuthWriter
	Users      *users.Store
}

// Run starts the rest endpoint
func (le *LoginEndpoint) Run() {
	router := gin.Default()

	authHandler := ginauth.MakeHandler(le.AuthReader)

	// TODO: must secure the API call, must not secure the user ID (https?)
	router.POST("/users", le.handleCreateUser)
	router.GET("/users/:uid", authHandler, le.handleUserInfo)
	router.DELETE("/users/:uid", authHandler, le.handleUserDelete)

	// TODO: must secure the API call, must not secure the user ID (https?)
	router.POST("/logins", le.handleLogin)
	router.DELETE("/logins/:uuid", authHandler, le.handleLogout)

	router.Run(le.Iface)
}

// POST /users
func (le *LoginEndpoint) handleCreateUser(c *gin.Context) {
	type UserCreds struct {
		Name     string `json:"user_name"`
		Password string `json:"user_password"`
	}

	var userCreds UserCreds
	if err := c.ShouldBindJSON(&userCreds); err != nil {
		c.Error(err)
		// Report full details when the REST API contract is violated
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to create user %v: %v", userCreds.Name, err)})
		return
	}

	userid, err := le.Users.CreateUser(userCreds.Name, userCreds.Password)
	if err != nil {
		c.Error(err)
		// Report sparse details when the deeper processing fails
		// TODO But what if the REST API does describe things like uniqueness of the user name?
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to create user: %v", userCreds.Name)})
		return
	}

	type UserInfo struct {
		ID   uint64 `json:"user_id"`
		Name string `json:"user_name"`
	}

	userInfo := UserInfo{ID: userid, Name: userCreds.Name}
	c.JSON(http.StatusOK, &userInfo)
}

// GET /users/:uid
func (le *LoginEndpoint) handleUserInfo(c *gin.Context) {
	idParam := c.Param("uid")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.Error(fmt.Errorf("Failed to parse uid %v: %v", idParam, err))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to parse uid %v: %v", idParam, err)})
		return
	}

	user, err := le.Users.GetUserByID(id)
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Failed to find user %v", id)})
		return
	}

	type UserInfo struct {
		ID   uint64 `json:"user_id"`
		Name string `json:"user_name"`
	}

	userInfo := UserInfo{ID: user.ID, Name: user.Name}
	c.JSON(http.StatusOK, &userInfo)
}

// DELETE /users/:uid
func (le *LoginEndpoint) handleUserDelete(c *gin.Context) {
	idParam := c.Param("uid")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.Error(fmt.Errorf("Failed to parse uid %v: %v", idParam, err))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to parse uid %v: %v", idParam, err)})
		return
	}

	err = le.Users.DeleteUserByID(id)
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Failed to delete user %v", id)})
		return
	}

	c.Status(http.StatusOK)
}

// POST /logins
func (le *LoginEndpoint) handleLogin(c *gin.Context) {
	type UserCreds struct {
		Name     string `json:"user_name"`
		Password string `json:"user_password"`
	}

	var userCreds UserCreds
	if err := c.ShouldBindJSON(&userCreds); err != nil {
		c.Error(err)
		// Report full details when the REST API contract is violated
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to login: %v", err)})
		return
	}

	user, err := le.Users.GetUserByName(userCreds.Name)
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Bad user or password"})
		return
	}

	if user.Password != userCreds.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Bad user or password"})
		return
	}

	token, err := le.AuthWriter.CreateToken(user.ID)
	if err != nil {
		c.Error(err)
		c.JSON(
			http.StatusInternalServerError,
			gin.H{"error": fmt.Sprintf("Failed to create login token for %v", userCreds.Name)})
		return
	}

	err = le.AuthWriter.CreateAuth(token)
	if err != nil {
		c.Error(err)
		c.JSON(
			http.StatusInternalServerError,
			gin.H{"error": fmt.Sprintf("Failed to record authentication %v", token.AccessUUID)})
		return
	}

	type LoginInfo struct {
		UserID       uint64 `json:"user_id"`
		LoginID      string `json:"login_id"`
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}

	loginInfo := LoginInfo{
		LoginID:      token.AccessUUID,
		UserID:       user.ID,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken}
	c.JSON(http.StatusOK, &loginInfo)
}

// DELETE /logins/:uuid
func (le *LoginEndpoint) handleLogout(c *gin.Context) {
	atUUID := c.Param("uuid")

	// TODO Check the kind of error: is the UUID missing or?
	if _, err := le.AuthWriter.DeleteAuth(atUUID); err != nil {
		c.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to delete authentication %v", atUUID)})
		return
	}

	c.Status(http.StatusOK)
}
