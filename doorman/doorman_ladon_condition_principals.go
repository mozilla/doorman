package doorman

import (
    "github.com/ory/ladon"
)

// InPrincipalsCondition is a condition which is fulfilled if the given value string is among principals.
type InPrincipalsCondition struct{}

// Fulfills returns true if the request's subject is equal to the given value string.
// This makes sense only because we iterate on principals and set the Request subject.
func (c *InPrincipalsCondition) Fulfills(value interface{}, r *ladon.Request) bool {
    s, ok := value.(string)

    return ok && s == r.Subject
}

// GetName returns the condition's name.
func (c *InPrincipalsCondition) GetName() string {
    return "InPrincipalsCondition"
}

func init() {
    ladon.ConditionFactories[new(InPrincipalsCondition).GetName()] = func() ladon.Condition {
        return new(InPrincipalsCondition)
    }
}
