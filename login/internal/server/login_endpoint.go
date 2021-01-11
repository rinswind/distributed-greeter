package server

import (
	"fmt"
	"net/http"
	"strconv"

	"example.org/services/login/internal/auth"
	"example.org/services/login/internal/users"
	"github.com/gin-gonic/gin"
)

var (
	router = gin.Default()
)

func init() {
	authHandler := gin.HandlerFunc(auth.Handler)

	router.POST("/users", handleCreateUser)
	router.GET("/users", authHandler, handleListUsers)

	router.GET("/users/:id", authHandler, handleUserInfo)
	router.DELETE("/users/:id", authHandler, handleUserDelete)

	router.POST("/login", handleLogin)
	router.POST("/logout", authHandler, handleLogout)
}

// Run starts the rest endpoint
func Run(iface string) {
	router.Run(iface)
}

// POST /users
func handleCreateUser(c *gin.Context) {
	type UserCreds struct {
		Name     string `json:"user_name"`
		Password string `json:"user_password"`
	}

	var userCreds UserCreds
	if err := c.ShouldBindJSON(&userCreds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userid, err := users.CreateUser(userCreds.Name, userCreds.Password)
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
func handleDeleteUser(c *gin.Context) {
	authCtx, _ := c.Get(auth.ContextKey)
	authClaims := authCtx.(map[string]interface{})
	userID, _ := authClaims["user_id"].(uint64)

	err := users.DeleteUserByID(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// GET /users
func handleListUsers(c *gin.Context) {
	type UsersInfo struct {
		IDs []uint64 `json:"user_ids"`
	}

	userIds := UsersInfo{IDs: *users.ListUserIDs()}
	c.JSON(http.StatusOK, &userIds)
}

// GET /users/:id
func handleUserInfo(c *gin.Context) {
	idParam := c.Param("id")

	var id uint64
	var err error

	if idParam == "current" {
		authCtx, _ := c.Get(auth.ContextKey)
		authClaims := authCtx.(map[string]interface{})

		id = uint64(authClaims["user_id"].(float64))
	} else {
		id, err = strconv.ParseUint(idParam, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v not a valid user ID", idParam)})
			return
		}
	}

	user, err := users.GetUserByID(id)
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
func handleUserDelete(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v not a valid user ID", idParam)})
		return
	}

	err = users.DeleteUserByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("User %v not found", id)})
		return
	}

	c.Status(http.StatusOK)
}

// POST /login
func handleLogin(c *gin.Context) {
	type UserCreds struct {
		Name     string `json:"user_name"`
		Password string `json:"user_password"`
	}

	var userCreds UserCreds
	if err := c.ShouldBindJSON(&userCreds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := users.GetUserByName(userCreds.Name)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Bad user or password"})
		return
	}

	if user.Pass != userCreds.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Bad user or password"})
		return
	}

	token, err := auth.CreateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = auth.CreateAuth(token)
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
func handleLogout(c *gin.Context) {
	authCtx, _ := c.Get(auth.ContextKey)
	authClaims := authCtx.(map[string]interface{})

	if _, err := auth.DeleteAuth(authClaims); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}
