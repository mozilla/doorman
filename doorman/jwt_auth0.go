package doorman

import (
	"fmt"
	"net/http"
	"strings"

	auth0 "github.com/auth0-community/go-auth0"
	log "github.com/sirupsen/logrus"
	jose "gopkg.in/square/go-jose.v2"
)

// Auth0Validator is the implementation of JWTValidator for Auth0.
type Auth0Validator struct {
	Issuer    string
	validator *auth0.JWTValidator
}

// Initialize will fetch Auth0 public keys and instantiate a validator.
func (v *Auth0Validator) Initialize() error {
	if !strings.HasPrefix(v.Issuer, "https://") || !strings.HasSuffix(v.Issuer, "auth0.com/") {
		return fmt.Errorf("issuer %q not supported or has bad format", v.Issuer)
	}
	jwksURI := fmt.Sprintf("%s.well-known/jwks.json", v.Issuer)
	log.Infof("JWT keys: %s", jwksURI)

	// Will check audience only when request comes in, leave empty for now.
	audience := []string{}

	client := auth0.NewJWKClient(auth0.JWKClientOptions{URI: jwksURI})
	config := auth0.NewConfiguration(client, audience, v.Issuer, jose.RS256)
	v.validator = auth0.NewValidator(config)
	return nil
}

// ExtractClaims validates the token from request, and returns the JWT claims.
func (v *Auth0Validator) ExtractClaims(request *http.Request) (*Claims, error) {
	token, err := v.validator.ValidateRequest(request)
	claims := Claims{}
	err = v.validator.Claims(request, token, &claims)
	if err != nil {
		return nil, err
	}
	return &claims, nil
}
