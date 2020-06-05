package api

import (
	"crypto/ed25519"
	"encoding/base64"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/haveachin/infrared/api/service"
	"github.com/pascaldekloe/jwt"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	claimsKey    = "claims"
	rolesKey     = "roles"
	isAdminKey   = "isAdmin"
	canCreateKey = "canCreate"
)

type credentials struct {
	Username string `json:"username,omitempty" binding:"required,base64"`
	Password string `json:"password,omitempty" binding:"required,base64"`
}

func (cred *credentials) decode() error {
	usernameBytes, err := base64.StdEncoding.DecodeString(cred.Username)
	if err != nil {
		return err
	}

	passwordBytes, err := base64.StdEncoding.DecodeString(cred.Password)
	if err != nil {
		return err
	}

	cred.Username = string(usernameBytes)
	cred.Password = string(passwordBytes)
	return nil
}

func jwtAuthenticator(publicKey ed25519.PublicKey) gin.HandlerFunc {
	keyRegister := jwt.KeyRegister{EdDSAs: []ed25519.PublicKey{publicKey}}
	return func(c *gin.Context) {
		claims, err := keyRegister.CheckHeader(c.Request)
		if err != nil {
			errStr := strings.TrimPrefix(err.Error(), "jwt: ")
			if err == jwt.ErrNoHeader {
				c.Header("WWW-Authenticate", "Bearer")
			} else {
				c.Header("WWW-Authenticate", `Bearer error="invalid_token", error_description=`+strconv.QuoteToASCII(errStr))
			}
			c.JSON(http.StatusUnauthorized, errorResp{Message: errStr})
			c.Abort()
			return
		}

		if !claims.Valid(time.Now()) {
			c.Header("WWW-Authenticate", `Bearer error="invalid_token", error_description="time constraints exceeded"`)
			c.JSON(http.StatusUnauthorized, errTokenOutdated)
			c.Abort()
			return
		}

		c.Set(claimsKey, claims)
		c.Set(isAdminKey, claims.Set[isAdminKey])
		c.Set(rolesKey, claims.Set[rolesKey])
		c.Set(canCreateKey, claims.Set[canCreateKey])
		c.Next()
	}
}

func jwtAuthenticationEndpoint(userService service.UserService, privateKey ed25519.PrivateKey) gin.HandlerFunc {
	return func(c *gin.Context) {
		var cred credentials
		if err := c.ShouldBind(&cred); err != nil {
			c.JSON(http.StatusBadRequest, formatValidationErrors(err))
			return
		}

		if err := cred.decode(); err != nil {
			c.JSON(http.StatusBadRequest, errorResp{Message: err.Error()})
			return
		}

		if err := binding.Validator.ValidateStruct(cred); err != nil {
			c.JSON(http.StatusBadRequest, formatValidationErrors(err))
			return
		}

		user := userService.User(cred.Username)
		if err := bcrypt.CompareHashAndPassword(user.Password, []byte(cred.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, errInvalidCredentials)
			return
		}

		roleIDs := make([]uint, len(user.Roles))

		for n, role := range user.Roles {
			roleIDs[n] = role.ID
		}

		now := time.Now()

		var claims = jwt.Claims{
			Registered: jwt.Registered{
				Issuer:  user.Username,
				Issued:  jwt.NewNumericTime(now),
				Expires: jwt.NewNumericTime(now.Add(12 * time.Hour)),
			},
			Set: map[string]interface{}{
				rolesKey:     roleIDs,
				isAdminKey:   user.IsAdmin,
				canCreateKey: user.CanCreate,
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
