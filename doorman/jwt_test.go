package doorman

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestMain defined in doorman_test.go
// func TestMain(m *testing.M) {}

type TestValidator struct {
	mock.Mock
}

func (v *TestValidator) Initialize() error {
	args := v.Called()
	return args.Error(0)
}
func (v *TestValidator) ExtractClaims(request *http.Request) (*Claims, error) {
	args := v.Called(request)
	return args.Get(0).(*Claims), args.Error(1)
}

func TestJWTMiddleware(t *testing.T) {
	v := &TestValidator{}
	v.On("Initialize").Return(nil)
	handler := VerifyJWTMiddleware(v)

	// Initialize() is called on server startup.
	v.AssertCalled(t, "Initialize")

	// Extract claims is ran on every request.
	claims := &Claims{
		Subject:  "ldap|user",
		Audience: []string{"https://some.domain.com"},
		Email:    "user@corp.com",
		Groups:   []string{"Employee", "Admins"},
	}
	v.On("ExtractClaims", mock.Anything).Return(claims, nil)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("GET", "/get", nil)
	c.Request.Header.Set("Origin", "https://some.domain.com")

	handler(c)

	v.AssertCalled(t, "ExtractClaims", c.Request)

	// Principals are set in context.
	principals, ok := c.Get(PrincipalsContextKey)
	require.True(t, ok)
	assert.Equal(t, principals, Principals{
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

	// Missing attributes in Payload
	claims = &Claims{
		Subject:  "ldap|user",
		Audience: []string{"https://some.domain.com"},
	}
	v = &TestValidator{}
	v.On("Initialize").Return(nil)
	v.On("ExtractClaims", mock.Anything).Return(claims, nil)
	c, _ = gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("GET", "/get", nil)
	c.Request.Header.Set("Origin", "https://some.domain.com")
	handler = VerifyJWTMiddleware(v)
	handler(c)
	principals, _ = c.Get(PrincipalsContextKey)
	assert.Equal(t, Principals{"userid:ldap|user"}, principals)
}
