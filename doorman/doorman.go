package doorman

import (
	"fmt"
)

// Context is used as request's context.
type Context map[string]interface{}

// Principals represent a user (userid, email, tags, ...)
type Principals []string

// Request is the authorization request.
type Request struct {
	// Principals are strings that identify the user.
	Principals Principals
	// Resource is the resource that access is requested to.
	Resource string
	// Action is the action that is requested on the resource.
	Action string
	// Context is the request's environmental context.
	Context Context
}

// Roles reads the roles from request context and returns the principals.
func (r *Request) Roles() Principals {
	p := Principals{}
	if roles, ok := r.Context["roles"]; ok {
		if rolesI, ok := roles.([]interface{}); ok {
			for _, roleI := range rolesI {
				if role, ok := roleI.(string); ok {
					prefixed := fmt.Sprintf("role:%s", role)
					p = append(p, prefixed)
				}
			}
		}
	}
	return p
}

// Doorman is the backend in charge of checking requests against policies.
type Doorman interface {
	// JWTIssuer returns the URL of the JWT issuer (if configured)
	JWTIssuer() string
	// ExpandPrincipals looks up and add extra principals to the ones specified.
	ExpandPrincipals(audience string, principals Principals) Principals
	// IsAllowed is responsible for deciding if subject can perform action on a resource with a context.
	IsAllowed(audience string, request *Request) bool
}
