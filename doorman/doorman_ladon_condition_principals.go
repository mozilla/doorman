package doorman

import (
	"github.com/ory/ladon"
)

// MatchPrincipalsCondition is a condition which is fulfilled if the given value string is among principals.
type MatchPrincipalsCondition struct{}

// Fulfills returns true if the request's subject is equal to the given value string.
// This makes sense only because we iterate on principals and set the Request subject.
func (c *MatchPrincipalsCondition) Fulfills(value interface{}, r *ladon.Request) bool {
	s, ok := value.(string)
	if ok {
		return s == r.Subject
	}
	l, ok := value.([]string)
	if ok {
		for _, s := range l {
			if s == r.Subject {
				return true
			}
		}
	}
	return false
}

// GetName returns the condition's name.
func (c *MatchPrincipalsCondition) GetName() string {
	return "MatchPrincipalsCondition"
}

func init() {
	ladon.ConditionFactories[new(MatchPrincipalsCondition).GetName()] = func() ladon.Condition {
		return new(MatchPrincipalsCondition)
	}
}
