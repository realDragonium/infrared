package api

import (
	"encoding/base64"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/haveachin/infrared/api/service"
	"net/http"
)

func userEndpoint(userService service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		username, err := usernameFromURL(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, errorResp{Message: err.Error()})
			return
		}

		c.JSON(http.StatusCreated, userService.User(username))
	}
}

func usersEndpoint(userService service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusCreated, userService.Users())
	}
}

func usernameFromURL(c *gin.Context) (string, error) {
	encodedUsername, ok := c.GetQuery("username")
	if !ok {
		return "", errors.New("no username in query params")
	}

	usernameBytes, err := base64.URLEncoding.DecodeString(encodedUsername)
	if err != nil {
		return "", errors.New("invalid encoding")
	}

	return string(usernameBytes), nil
}
