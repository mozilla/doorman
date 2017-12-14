package doorman

import (
	"net/http"

	auth0 "github.com/auth0-community/go-auth0"
	jwt "gopkg.in/square/go-jose.v2/jwt"
)

// MozillaClaims uses specific attributes for emails and groups
type MozillaClaims struct {
	Subject  string       `json:"sub"`
	Audience jwt.Audience `json:"aud"`
	Email    string       `json:"email"`
	Emails   []string     `json:"https://sso.mozilla.com/claim/emails"`
	Groups   []string     `json:"https://sso.mozilla.com/claim/groups"`
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

	// In case the JWT was not requested with the profile or email scope.
	email := mozclaims.Email
	if email == "" && len(mozclaims.Emails) > 0 {
		email = mozclaims.Emails[0]
	}

	claims := Claims{
		Subject:  mozclaims.Subject,
		Audience: mozclaims.Audience,
		Email:    email,
		Groups:   mozclaims.Groups,
	}
	return &claims, nil
}
