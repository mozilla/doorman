package warden

import (
	"errors"
	"fmt"
	"net/http"

	auth0 "github.com/auth0-community/go-auth0"
	"github.com/gin-gonic/gin"
	jose "gopkg.in/square/go-jose.v2"
	jwt "gopkg.in/square/go-jose.v2/jwt"
)

func verifyJWT(request *http.Request) (*jwt.Claims, error) {
	domain := request.Header.Get("Auth0-Domain")
	if domain == "" {
		return nil, errors.New("Auth0-Domain header missing")
	}

	apiIdentifer := request.Header.Get("Auth0-Audience")
	if apiIdentifer == "" {
		return nil, errors.New("Auth0-API-Identifier header missing")
	}

	jwksURI := fmt.Sprintf("https://%s.auth0.com/.well-known/jwks.json", domain)
	apiIssuer := fmt.Sprintf("https://%s.auth0.com/", domain)
	audience := []string{apiIdentifer}

	client := auth0.NewJWKClient(auth0.JWKClientOptions{URI: jwksURI})
	config := auth0.NewConfiguration(client, audience, apiIssuer, jose.RS256)
	validator := auth0.NewValidator(config)

	token, err := validator.ValidateRequest(request)
	if err != nil {
		return nil, err
	}

	claims := jwt.Claims{}
	err = validator.Claims(request, token, &claims)
	if err != nil {
		return nil, err
	}

	return &claims, nil
}

// VerifyJWTMiddleware makes sure a valid JWT is provided.
func VerifyJWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := verifyJWT(c.Request)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": err.Error(),
			})
			return
		}

		c.Set("JWT", claims)

		c.Next()
	}
}
