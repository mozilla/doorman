package doorman

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	jwt "gopkg.in/square/go-jose.v2/jwt"
)

// PrincipalsContextKey is the Gin context key to obtain the current user principals.
const PrincipalsContextKey string = "principals"

// Claims is the set of information we extract from the JWT payload.
type Claims struct {
	Subject  string       `json:"sub,omitempty"`
	Audience jwt.Audience `json:"aud,omitempty"`
	Email    string       `json:"email,omitempty"`
	Groups   []string     `json:"groups,omitempty"`
}

// JWTValidator is the interface in charge of extracting JWT claims from request.
type JWTValidator interface {
	Initialize() error
	ExtractClaims(*http.Request) (*Claims, error)
}

// VerifyJWTMiddleware makes sure a valid JWT is provided.
func VerifyJWTMiddleware(validator JWTValidator) gin.HandlerFunc {
	validator.Initialize()

	return func(c *gin.Context) {
		claims, err := validator.ExtractClaims(c.Request)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": err.Error(),
			})
			return
		}

		// The service requesting must send its location. It will be compared
		// with the audiences defined in policies files.
		// XXX: The Origin request header might not be the best choice.
		origin := c.Request.Header.Get("Origin")
		if origin == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": "Missing `Origin` request header",
			})
			return
		}
		// Check that origin matches audiences from JWT token .
		if !claims.Audience.Contains(origin) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"message": "Invalid audience claim",
			})
			return
		}

		// Extract principals from JWT
		var principals Principals
		userid := fmt.Sprintf("userid:%s", claims.Subject)
		principals = append(principals, userid)

		// Main email (no alias)
		if claims.Email != "" {
			email := fmt.Sprintf("email:%s", claims.Email)
			principals = append(principals, email)
		}

		// Groups
		for _, group := range claims.Groups {
			prefixed := fmt.Sprintf("group:%s", group)
			principals = append(principals, prefixed)
		}

		c.Set(PrincipalsContextKey, principals)

		c.Next()
	}
}
