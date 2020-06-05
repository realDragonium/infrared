package api

import (
	"crypto/ed25519"
	"github.com/haveachin/infrared/api/service"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func New(cfg Config) (*http.Server, error) {
	gin.SetMode(gin.ReleaseMode)
	r, err := ginRouter(cfg)
	if err != nil {
		return nil, err
	}

	return &http.Server{
		Handler:      r,
		Addr:         cfg.Address,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}, nil
}

func ginRouter(cfg Config) (http.Handler, error) {
	if err := setupValidator(); err != nil {
		return nil, err
	}

	dataService, err := service.NewMySQL(cfg.MySQL)
	if err != nil {
		return nil, err
	}

	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, err
	}

	router := gin.New()
	router.Use(gin.Recovery())

	api := router.Group("/api")
	{
		api.POST("/authenticate", jwtAuthenticationEndpoint(dataService, privateKey))

		v1 := api.Group("/v1", jwtAuthenticator(publicKey))
		{
			// Register
			v1.GET("/register", registerTokenEndpoint(privateKey))
			v1.POST("/register", registerEndpoint(dataService))

			// User
			v1.GET("/user", userEndpoint(dataService))
			v1.GET("/users", usersEndpoint(dataService))
		}
	}

	return router, nil
}
