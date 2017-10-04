package doorman

import (
	"net/http"
	"fmt"

	auth0 "github.com/auth0-community/go-auth0"
	"github.com/gin-gonic/gin"
	jose "gopkg.in/square/go-jose.v2"
	jwt "gopkg.in/square/go-jose.v2/jwt"
	log "github.com/sirupsen/logrus"
)

func newJWTValidator(jwtIssuer string) *auth0.JWTValidator {
	jwksURI := fmt.Sprintf("%s.well-known/jwks.json", jwtIssuer)
	log.Infof("JWT keys: %s", jwksURI)

	// XXX: do not expected any specific audience?
	audience := []string{}

	client := auth0.NewJWKClient(auth0.JWKClientOptions{URI: jwksURI})
	config := auth0.NewConfiguration(client, audience, jwtIssuer, jose.RS256)
	validator := auth0.NewValidator(config)

	return validator
}

func verifyJWT(validator *auth0.JWTValidator, request *http.Request) (*jwt.Claims, error) {
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
func VerifyJWTMiddleware(jwtIssuer string) gin.HandlerFunc {
	validator := newJWTValidator(jwtIssuer)

	return func(c *gin.Context) {
		claims, err := verifyJWT(validator, c.Request)

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
