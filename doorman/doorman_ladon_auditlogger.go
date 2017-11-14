package doorman

import (
	"os"

    "github.com/ory/ladon"
    "github.com/sirupsen/logrus"
    "go.mozilla.org/mozlogrus"
)

type AuditLogger struct {
    logger *logrus.Logger
}

func NewAuditLogger() *AuditLogger {
    authzLog := &logrus.Logger{
        Out:       os.Stdout,
        Formatter: &mozlogrus.MozLogFormatter{LoggerName: "iam", Type: "request.authorization"},
        Hooks:     make(logrus.LevelHooks),
        Level:     logrus.InfoLevel,
    }
    authzLog.Info("ta mere en slip")
    return &AuditLogger{logger: authzLog}
}

func (a *AuditLogger) logRequest(allowed bool, r *ladon.Request, policies ladon.Policies) {
    policiesNames := []string{}
    for _, p := range policies {
        policiesNames = append(policiesNames, p.GetID())
    }

    a.logger.WithFields(
        logrus.Fields{
            "allowed":  allowed,
            "policies": policiesNames,
            "subject":  r.Subject,
            "action":   r.Action,
            "resource": r.Resource,
            "context":  r.Context,
        },
    ).Info("")
}

// LogRejectedAccessRequest is called by Ladon when a request is denied.
func (a *AuditLogger) LogRejectedAccessRequest(request *ladon.Request, pool ladon.Policies, deciders ladon.Policies) {
    if len(deciders) > 0 {
        // Explicitly denied by the last one.
        a.logRequest(false, request, deciders[len(deciders) - 1:len(deciders) - 1])
    } else {
        // No matching policy.
        a.logRequest(false, request, deciders)
    }
}

// LogGrantedAccessRequest is called by Ladon when a request is granted.
func (a *AuditLogger) LogGrantedAccessRequest(request *ladon.Request, pool ladon.Policies, deciders ladon.Policies) {
    a.logRequest(true, request, deciders)
}
