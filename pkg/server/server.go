package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ouzi-dev/needs-retitle/pkg/plugin"
	"github.com/sirupsen/logrus"
	"k8s.io/test-infra/prow/github"
)

// Server implements http.Handler. It validates incoming GitHub webhooks and
// then dispatches them to the appropriate plugins.
type Server struct {
	tokenGenerator func() []byte
	ghc            github.Client
	log            *logrus.Entry
	p              *plugin.Plugin
}

func NewServer(tokenGenerator func() []byte, ghc github.Client, log *logrus.Entry, p *plugin.Plugin) *Server {
	return &Server{
		tokenGenerator: tokenGenerator,
		ghc:            ghc,
		log:            log,
		p:              p,
	}
}

// ServeHTTP validates an incoming webhook and puts it into the event channel.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: Move webhook handling logic out of hook binary so that we don't have to import all
	// plugins just to validate the webhook.
	eventType, eventGUID, payload, ok, _ := github.ValidateWebhook(w, r, s.tokenGenerator)
	if !ok {
		return
	}
	fmt.Fprint(w, "Event received. Have a nice day.")

	if err := s.handleEvent(eventType, eventGUID, payload); err != nil {
		logrus.WithError(err).Error("Error parsing event.")
	}
}

func (s *Server) handleEvent(eventType, eventGUID string, payload []byte) error {
	l := s.log.WithFields(
		logrus.Fields{
			"event-type":     eventType,
			github.EventGUID: eventGUID,
		},
	)
	switch eventType {
	case "pull_request":
		var pre github.PullRequestEvent
		if err := json.Unmarshal(payload, &pre); err != nil {
			return err
		}
		go func() {
			if err := s.p.HandlePullRequestEvent(l, s.ghc, &pre); err != nil {
				l.WithField("event-type", eventType).WithError(err).Info("Error handling event.")
			}
		}()
	case "issue_comment":
		var ice github.IssueCommentEvent
		if err := json.Unmarshal(payload, &ice); err != nil {
			return err
		}
		go func() {
			if err := s.p.HandleIssueCommentEvent(l, s.ghc, &ice); err != nil {
				l.WithField("event-type", eventType).WithError(err).Info("Error handling event.")
			}
		}()
	default:
		s.log.Debugf("received an event of type %q but didn't ask for it", eventType)
	}
	return nil
}
