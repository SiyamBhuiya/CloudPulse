package alerting

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/smtp"
	"strings"
	"sync"
	"time"

	"cloudpulse/internal/config"
	"cloudpulse/internal/repository"
)

type Alert struct {
	Channel     string
	Email       string
	RuleName    string
	ServerName  string
	Metric      string
	Threshold   float64
	Value       float64
	DurationMin int
	Resolved    bool
}

func newAlert(c repository.Candidate, value float64, resolved bool) Alert {
	return Alert{
		Channel: c.Channel, Email: c.Email, RuleName: c.RuleName, ServerName: c.ServerName,
		Metric: c.Metric, Threshold: c.Threshold, Value: value,
		DurationMin: c.DurationMin, Resolved: resolved,
	}
}

var metricLabel = map[string]string{"cpu": "CPU", "memory": "Memory", "disk": "Disk"}

func (a Alert) state() string {
	if a.Resolved {
		return "RESOLVED"
	}
	return "FIRING"
}

func (a Alert) subject() string {
	return clean(fmt.Sprintf("[CloudPulse] %s: %s on %s", a.state(), a.RuleName, a.ServerName))
}

func (a Alert) text() string {
	if a.Resolved {
		return fmt.Sprintf("%s on %s is back to normal: %s is %.0f%% (limit %.0f%%).",
			a.RuleName, a.ServerName, metricLabel[a.Metric], a.Value, a.Threshold)
	}
	return fmt.Sprintf("%s on %s: %s has been above %.0f%% for %d min (now %.0f%%).",
		a.RuleName, a.ServerName, metricLabel[a.Metric], a.Threshold, a.DurationMin, a.Value)
}

// clean removes line breaks so user-written names cannot inject email headers.
func clean(s string) string {
	return strings.NewReplacer("\r", " ", "\n", " ").Replace(s)
}

type Notifier struct {
	cfg    *config.Config
	client *http.Client
	wg     sync.WaitGroup
}

func NewNotifier(cfg *config.Config) *Notifier {
	return &Notifier{cfg: cfg, client: &http.Client{Timeout: 10 * time.Second}}
}

// Notify sends in the background so a slow Slack or mail server never holds up the evaluator.
func (n *Notifier) Notify(a Alert) {
	n.wg.Add(1)
	go func() {
		defer n.wg.Done()
		var err error
		switch a.Channel {
		case "slack":
			err = n.slack(a)
		case "webhook":
			err = n.webhook(a)
		case "email":
			err = n.email(a)
		}
		if err != nil {
			slog.Warn("notification not sent", "channel", a.Channel, "rule", a.RuleName, "err", err)
		}
	}()
}

func (n *Notifier) Wait() { n.wg.Wait() }

func (n *Notifier) postJSON(url string, body any) error {
	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}
	resp, err := n.client.Post(url, "application/json", bytes.NewReader(raw))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("returned status %d", resp.StatusCode)
	}
	return nil
}

func (n *Notifier) slack(a Alert) error {
	if n.cfg.SlackWebhookURL == "" {
		return fmt.Errorf("SLACK_WEBHOOK_URL is not set")
	}
	return n.postJSON(n.cfg.SlackWebhookURL, map[string]string{"text": a.subject() + "\n" + a.text()})
}

func (n *Notifier) webhook(a Alert) error {
	if n.cfg.AlertWebhookURL == "" {
		return fmt.Errorf("ALERT_WEBHOOK_URL is not set")
	}
	return n.postJSON(n.cfg.AlertWebhookURL, map[string]any{
		"event":       strings.ToLower(a.state()),
		"rule":        a.RuleName,
		"server":      a.ServerName,
		"metric":      a.Metric,
		"threshold":   a.Threshold,
		"value":       a.Value,
		"durationMin": a.DurationMin,
		"message":     a.text(),
		"time":        time.Now().UTC().Format(time.RFC3339),
	})
}

func (n *Notifier) email(a Alert) error {
	c := n.cfg
	if c.SMTPHost == "" || c.SMTPFrom == "" {
		return fmt.Errorf("SMTP_HOST and SMTP_FROM are not set")
	}
	var auth smtp.Auth
	if c.SMTPUser != "" {
		auth = smtp.PlainAuth("", c.SMTPUser, c.SMTPPass, c.SMTPHost)
	}
	msg := "From: " + clean(c.SMTPFrom) + "\r\n" +
		"To: " + clean(a.Email) + "\r\n" +
		"Subject: " + a.subject() + "\r\n" +
		"MIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n" +
		a.text() + "\r\n"
	return smtp.SendMail(c.SMTPHost+":"+c.SMTPPort, auth, c.SMTPFrom, []string{a.Email}, []byte(msg))
}
