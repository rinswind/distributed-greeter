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

	router.POST("/users", le.handleCreateUser)
	router.GET("/users", authHandler, le.handleListUsers)

	router.GET("/users/:id", authHandler, le.handleUserInfo)
	router.DELETE("/users/:id", authHandler, le.handleUserDelete)

	router.POST("/login", le.handleLogin)
	router.POST("/logout", authHandler, le.handleLogout)
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userid, err := le.Users.CreateUser(userCreds.Name, userCreds.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	type UserInfo struct {
		ID   uint64 `json:"user_id"`
		Name string `json:"user_name"`
	}

	userInfo := UserInfo{ID: userid, Name: userCreds.Name}
	c.JSON(http.StatusOK, &userInfo)
}

// DELETE /users
func (le *LoginEndpoint) handleDeleteUser(c *gin.Context) {
	authCtx, _ := c.Get(ginauth.ContextKey)
	authClaims := authCtx.(map[string]interface{})
	userID, _ := authClaims["user_id"].(uint64)

	err := le.Users.DeleteUserByID(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// GET /users
func (le *LoginEndpoint) handleListUsers(c *gin.Context) {
	type UsersInfo struct {
		IDs []uint64 `json:"user_ids"`
	}

	userIds := UsersInfo{IDs: *le.Users.ListUserIDs()}
	c.JSON(http.StatusOK, &userIds)
}

// GET /users/:id
func (le *LoginEndpoint) handleUserInfo(c *gin.Context) {
	idParam := c.Param("id")

	var id uint64
	var err error

	if idParam == "current" {
		authCtx, _ := c.Get(ginauth.ContextKey)
		authClaims := authCtx.(map[string]interface{})

		id = uint64(authClaims["user_id"].(float64))
	} else {
		id, err = strconv.ParseUint(idParam, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v not a valid user ID", idParam)})
			return
		}
	}

	user, err := le.Users.GetUserByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("User %v not found", id)})
		return
	}

	type UserInfo struct {
		ID   uint64 `json:"user_id"`
		Name string `json:"user_name"`
	}

	userInfo := UserInfo{ID: user.ID, Name: user.Name}
	c.JSON(http.StatusOK, &userInfo)
}

// DELETE /users/:id
func (le *LoginEndpoint) handleUserDelete(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v not a valid user ID", idParam)})
		return
	}

	err = le.Users.DeleteUserByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("User %v not found", id)})
		return
	}

	c.Status(http.StatusOK)
}

// POST /login
func (le *LoginEndpoint) handleLogin(c *gin.Context) {
	type UserCreds struct {
		Name     string `json:"user_name"`
		Password string `json:"user_password"`
	}

	var userCreds UserCreds
	if err := c.ShouldBindJSON(&userCreds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := le.Users.GetUserByName(userCreds.Name)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Bad user or password"})
		return
	}

	if user.Pass != userCreds.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Bad user or password"})
		return
	}

	token, err := le.AuthWriter.CreateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = le.AuthWriter.CreateAuth(token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type LoginInfo struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}

	loginInfo := LoginInfo{AccessToken: token.AccessToken, RefreshToken: token.RefreshToken}
	c.JSON(http.StatusOK, &loginInfo)
}

// POST /logout
func (le *LoginEndpoint) handleLogout(c *gin.Context) {
	authCtx, _ := c.Get(ginauth.ContextKey)
	authClaims := authCtx.(map[string]interface{})

	if _, err := le.AuthWriter.DeleteAuth(authClaims); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}
