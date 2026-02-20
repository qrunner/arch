// Package notifier implements alert evaluation and notification delivery
// through various channels (webhook, email, telegram).
package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/qrunner/arch/internal/model"
	"go.uber.org/zap"
)

// AlertRule defines a condition that triggers a notification.
type AlertRule struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Actions  []string `json:"actions"`  // e.g. ["asset.created", "asset.removed"]
	Sources  []string `json:"sources"`  // filter by source, empty = all
	Channels []string `json:"channels"` // "webhook", "email", "telegram"
}

// Notifier evaluates alert rules against change events and sends notifications.
type Notifier struct {
	rules      []AlertRule
	webhookURL string
	logger     *zap.Logger
}

// New creates a new Notifier.
func New(logger *zap.Logger, webhookURL string) *Notifier {
	return &Notifier{
		rules:      []AlertRule{},
		webhookURL: webhookURL,
		logger:     logger,
	}
}

// AddRule registers a new alert rule.
func (n *Notifier) AddRule(rule AlertRule) {
	n.rules = append(n.rules, rule)
}

// Evaluate checks a change event against all rules and sends notifications.
func (n *Notifier) Evaluate(ctx context.Context, event *model.ChangeEvent) {
	for _, rule := range n.rules {
		if n.matches(rule, event) {
			n.logger.Info("alert rule matched",
				zap.String("rule", rule.Name),
				zap.String("action", string(event.Action)),
			)
			for _, ch := range rule.Channels {
				switch ch {
				case "webhook":
					n.sendWebhook(ctx, event)
				case "email":
					// TODO: implement SMTP email sending
					n.logger.Info("email notification not yet implemented")
				case "telegram":
					// TODO: implement Telegram bot API
					n.logger.Info("telegram notification not yet implemented")
				}
			}
		}
	}
}

func (n *Notifier) matches(rule AlertRule, event *model.ChangeEvent) bool {
	actionMatch := false
	for _, a := range rule.Actions {
		if a == string(event.Action) {
			actionMatch = true
			break
		}
	}
	if !actionMatch {
		return false
	}

	if len(rule.Sources) > 0 {
		sourceMatch := false
		for _, s := range rule.Sources {
			if s == event.Source {
				sourceMatch = true
				break
			}
		}
		if !sourceMatch {
			return false
		}
	}

	return true
}

func (n *Notifier) sendWebhook(ctx context.Context, event *model.ChangeEvent) {
	if n.webhookURL == "" {
		return
	}

	payload, _ := json.Marshal(event)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.webhookURL, bytes.NewReader(payload))
	if err != nil {
		n.logger.Error("creating webhook request", zap.Error(err))
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		n.logger.Error("sending webhook", zap.Error(err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		n.logger.Warn("webhook returned non-success",
			zap.Int("status", resp.StatusCode),
			zap.String("url", n.webhookURL),
		)
	}
}

// HandleNATSMessage processes a NATS message containing a change event.
func (n *Notifier) HandleNATSMessage(ctx context.Context, data []byte) {
	var event model.ChangeEvent
	if err := json.Unmarshal(data, &event); err != nil {
		n.logger.Error("unmarshaling event", zap.Error(err))
		return
	}
	n.Evaluate(ctx, &event)
}

var _ fmt.Stringer = (*AlertRule)(nil)

// String returns a human-readable representation of the alert rule.
func (r *AlertRule) String() string {
	return fmt.Sprintf("AlertRule{name=%s, actions=%v, channels=%v}", r.Name, r.Actions, r.Channels)
}
