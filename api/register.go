package api

import (
	"crypto/ed25519"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/haveachin/infrared/api/service"
	"github.com/pascaldekloe/jwt"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"
)

func registerTokenEndpoint(privateKey ed25519.PrivateKey) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !c.MustGet(isAdminKey).(bool) {
			c.JSON(http.StatusUnauthorized, errNotAnAdmin)
			return
		}

		usernameOfNewUser, err := usernameFromURL(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, errorResp{Message: err.Error()})
			return
		}

		if usernameOfNewUser == "" {
			c.JSON(http.StatusBadRequest, errUsernameEmpty)
			return
		}

		now := time.Now()

		var claims = jwt.Claims{
			Registered: jwt.Registered{
				Issuer:  c.MustGet(claimsKey).(jwt.Claims).Issuer,
				Subject: usernameOfNewUser,
				Issued:  jwt.NewNumericTime(now),
				Expires: jwt.NewNumericTime(now.Add(24 * time.Hour)),
			},
		}

		token, err := claims.EdDSASign(privateKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, errTokenSigning)
			return
		}

		c.JSON(http.StatusOK, gin.H{"token": string(token)})
	}
}

func registerEndpoint(userService service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.MustGet(claimsKey).(jwt.Claims).Subject
		if username == "" {
			c.JSON(http.StatusUnauthorized, errTokenInvalid)
			return
		}

		var cred credentials
		if err := c.ShouldBind(&cred); err != nil {
			c.JSON(http.StatusBadRequest, formatValidationErrors(err))
			return
		}

		if err := cred.decode(); err != nil {
			c.JSON(http.StatusBadRequest, errorResp{Message: err.Error()})
			return
		}

		user := service.User{Username: username, Password: []byte(cred.Password)}
		if err := binding.Validator.ValidateStruct(user); err != nil {
			c.JSON(http.StatusBadRequest, formatValidationErrors(err))
			return
		}

		hash, err := bcrypt.GenerateFromPassword(user.Password, 10)
		if err != nil {
			c.JSON(http.StatusInternalServerError, errPasswordHashing)
			return
		}

		user.Password = hash
		if success := userService.CreateUser(&user); !success {
			c.JSON(http.StatusInternalServerError, errUsernameTaken)
			return
		}

		c.JSON(http.StatusCreated, user)
	}
}
