package doorman

import (
	"os"

	"github.com/ory/ladon"
	"github.com/sirupsen/logrus"
	"go.mozilla.org/mozlogrus"
)

type auditLogger struct {
	logger *logrus.Logger
}

func newAuditLogger() *auditLogger {
	authzLog := &logrus.Logger{
		Out:       os.Stdout,
		Formatter: &mozlogrus.MozLogFormatter{LoggerName: "doorman", Type: "request.authorization"},
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.InfoLevel,
	}
	return &auditLogger{logger: authzLog}
}

func (a *auditLogger) logRequest(allowed bool, r *ladon.Request, policies ladon.Policies) {
	policiesNames := []string{}
	for _, p := range policies {
		policiesNames = append(policiesNames, p.GetID())
	}

	// Remove custom values out of context for nicer logging (were set in handler)
	var principals Principals
	var audience string
	var remoteIP string
	context := map[string]interface{}{}
	for k, v := range r.Context {
		if k == "principals" {
			principals = v.(Principals)
		} else if k == "audience" {
			audience = v.(string)
		} else if k == "remoteIP" {
			remoteIP = v.(string)
		} else {
			context[k] = v
		}
	}

	a.logger.WithFields(
		logrus.Fields{
			"allowed":    allowed,
			"principals": principals,
			"audience":   audience,
			"remoteIP":   remoteIP,
			"policies":   policiesNames,
			"action":     r.Action,
			"resource":   r.Resource,
			"context":    context,
		},
	).Info("")
}

// LogRejectedAccessRequest is called by Ladon when a request is denied.
func (a *auditLogger) LogRejectedAccessRequest(request *ladon.Request, pool ladon.Policies, deciders ladon.Policies) {
	// Since we iterate on principals to test individual subjects, when a request is denied
	// we want to log the last one only, ie. when r.subject == last(principals)
	principals := request.Context["principals"].(Principals)
	if request.Subject != principals[len(principals)-1] {
		return
	}

	if len(deciders) > 0 {
		// Explicitly denied by the last one.
		a.logRequest(false, request, deciders[len(deciders)-1:])
	} else {
		// No matching policy.
		a.logRequest(false, request, deciders)
	}
}

// LogGrantedAccessRequest is called by Ladon when a request is granted.
func (a *auditLogger) LogGrantedAccessRequest(request *ladon.Request, pool ladon.Policies, deciders ladon.Policies) {
	a.logRequest(true, request, deciders)
}
