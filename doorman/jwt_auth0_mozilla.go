package doorman

import (
	"net/http"

	auth0 "github.com/auth0-community/go-auth0"
	jwt "gopkg.in/square/go-jose.v2/jwt"
)

// MozillaClaims uses specific attributes for emails and groups
type MozillaClaims struct {
	Subject  string       `json:"sub,omitempty"`
	Audience jwt.Audience `json:"aud,omitempty"`
	Emails   []string     `json:"https://sso.mozilla.com/claim/emails,omitempty"`
	Groups   []string     `json:"https://sso.mozilla.com/claim/groups,omitempty"`
}

// MozillaAuth0Validator is the implementation of JWTValidator for Auth0.
type MozillaAuth0Validator struct {
	Issuer    string
	validator *auth0.JWTValidator
}

// Initialize will fetch Auth0 public keys and instantiate a validator.
func (v *MozillaAuth0Validator) Initialize() error {
	validator, err := auth0Validator(v.Issuer)
	if err != nil {
		return err
	}
	v.validator = validator
	return nil
}

// ExtractClaims validates the token from request, and returns the JWT claims.
func (v *MozillaAuth0Validator) ExtractClaims(request *http.Request) (*Claims, error) {
	token, err := v.validator.ValidateRequest(request)
	mozclaims := MozillaClaims{}
	err = v.validator.Claims(request, token, &mozclaims)
	if err != nil {
		return nil, err
	}
	claims := Claims{
		Subject:  mozclaims.Subject,
		Audience: mozclaims.Audience,
		Email:    mozclaims.Emails[0],
		Groups:   mozclaims.Groups,
	}
	return &claims, nil
}
