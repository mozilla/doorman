package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mozilla/doorman/authn"
	"github.com/mozilla/doorman/doorman"
)

type TestAuthenticator struct {
	mock.Mock
}

func (v *TestAuthenticator) ValidateRequest(request *http.Request) (*authn.UserInfo, error) {
	args := v.Called(request)
	return args.Get(0).(*authn.UserInfo), args.Error(1)
}

func TestAuthnMiddleware(t *testing.T) {
	d := doorman.NewDefaultLadon()
	handler := AuthnMiddleware(d)

	audience := "https://some.api.com"

	// Associate a fake JWT validator to this issuer.
	v := &TestAuthenticator{}
	d.SetAuthenticator(audience, v)

	// Extract claims is ran on every request.
	claims := &authn.UserInfo{
		ID:     "ldap|user",
		Email:  "user@corp.com",
		Groups: []string{"Employee", "Admins"},
	}
	v.On("ValidateRequest", mock.Anything).Return(claims, nil)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("GET", "/get", nil)
	c.Request.Header.Set("Origin", audience)

	handler(c)

	v.AssertCalled(t, "ValidateRequest", c.Request)

	// Principals are set in context.
	principals, ok := c.Get(PrincipalsContextKey)
	require.True(t, ok)
	assert.Equal(t, principals, doorman.Principals{
		"userid:ldap|user",
		"email:user@corp.com",
		"group:Employee",
		"group:Admins",
	})

	c, _ = gin.CreateTestContext(httptest.NewRecorder())

	// Missing origin.
	c.Request, _ = http.NewRequest("GET", "/get", nil)
	handler(c)
	_, ok = c.Get(PrincipalsContextKey)
	assert.False(t, ok)

	// Wrong origin.
	c.Request, _ = http.NewRequest("GET", "/get", nil)
	c.Request.Header.Set("Origin", "https://wrong.com")
	handler(c)
	_, ok = c.Get(PrincipalsContextKey)
	assert.False(t, ok)

	// Authentication not configured for this origin.
	d.SetAuthenticator("https://open", nil)

	c.Request, _ = http.NewRequest("GET", "/get", nil)
	c.Request.Header.Set("Origin", "https://open")
	handler(c)
	_, ok = c.Get(PrincipalsContextKey)
	assert.False(t, ok)

	// Userinfo are set as principals in request context.
	claims = &authn.UserInfo{
		ID: "ldap|user",
	}
	v = &TestAuthenticator{}
	v.On("ValidateRequest", mock.Anything).Return(claims, nil)
	d.SetAuthenticator(audience, v)
	c, _ = gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("GET", "/get", nil)
	c.Request.Header.Set("Origin", audience)
	handler(c)
	principals, _ = c.Get(PrincipalsContextKey)
	assert.Equal(t, doorman.Principals{"userid:ldap|user"}, principals)
}
