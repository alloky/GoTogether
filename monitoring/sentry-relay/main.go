// sentry-relay: receives GlitchTip (or Sentry) outbound webhook events
// and forwards them to a Telegram group as formatted messages.
//
// Supports both GlitchTip and Sentry webhook payload formats.
//
// Environment variables:
//   ALERT_BOT_TOKEN  — Telegram bot token (the alert bot, NOT the GoTogether bot)
//   ALERT_CHAT_ID    — Telegram group chat ID (negative number for groups, e.g. -1001234567890)
//   SENTRY_SECRET    — optional shared secret to validate webhook calls
//   GLITCHTIP_URL    — base URL of GlitchTip instance (default: http://localhost:8100)
//   PORT             — HTTP listen port (default: 9456)

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// SentryEvent covers both Sentry and GlitchTip webhook payloads.
// GlitchTip sends a subset of Sentry's fields.
type SentryEvent struct {
	// Sentry format: "action" at top level
	Action string `json:"action"`
	Data   struct {
		Issue IssuePayload `json:"issue"`
	} `json:"data"`
	Actor struct {
		Name string `json:"name"`
	} `json:"actor"`

	// GlitchTip format: issue fields may be at top level
	// (GlitchTip sends {"issue_id":..., "project":..., ...})
	IssueID    int    `json:"issue_id,omitempty"`
	Title      string `json:"title,omitempty"`
	Culprit    string `json:"culprit,omitempty"`
	Level      string `json:"level,omitempty"`
	ProjectStr string `json:"project,omitempty"`
}

type IssuePayload struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Level    string `json:"level"`
	Status   string `json:"status"`
	Culprit  string `json:"culprit"`
	ShortID  string `json:"shortId"`
	Metadata struct {
		Value string `json:"value"`
		Type  string `json:"type"`
	} `json:"metadata"`
	Project struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	} `json:"project"`
	Permalink string `json:"permalink"`
	FirstSeen string `json:"firstSeen"`
	LastSeen  string `json:"lastSeen"`
	TimesSeen int    `json:"timesSeen"`
}

// normalizeEvent merges GlitchTip's flat format into the Sentry-style
// nested format so formatting logic can stay in one place.
func normalizeEvent(e *SentryEvent, glitchtipURL string) {
	// GlitchTip sends issue_id at root, no nested data.issue
	if e.IssueID > 0 && e.Data.Issue.ID == "" {
		e.Data.Issue.ID = fmt.Sprintf("%d", e.IssueID)
		e.Data.Issue.Title = e.Title
		e.Data.Issue.Level = e.Level
		e.Data.Issue.Culprit = e.Culprit
		e.Data.Issue.Project.Name = e.ProjectStr
		e.Data.Issue.Permalink = fmt.Sprintf("%s/issues/%d", glitchtipURL, e.IssueID)
		if e.Action == "" {
			e.Action = "created"
		}
	}
}

func levelEmoji(level string) string {
	switch strings.ToLower(level) {
	case "fatal":
		return "💀"
	case "error":
		return "🔴"
	case "warning", "warn":
		return "🟡"
	case "info":
		return "🔵"
	default:
		return "⚪"
	}
}

func actionEmoji(action string) string {
	switch action {
	case "created":
		return "🆕"
	case "resolved":
		return "✅"
	case "assigned":
		return "👤"
	case "ignored":
		return "🔕"
	default:
		return "📌"
	}
}

func formatTelegramMessage(e SentryEvent) string {
	issue := e.Data.Issue
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%s %s <b>[%s]</b> %s\n",
		actionEmoji(e.Action),
		levelEmoji(issue.Level),
		strings.ToUpper(e.Action),
		htmlEscape(issue.Title),
	))

	sb.WriteString(fmt.Sprintf("📦 Project: <code>%s</code>\n", htmlEscape(issue.Project.Name)))

	if issue.Culprit != "" {
		sb.WriteString(fmt.Sprintf("📍 Culprit: <code>%s</code>\n", htmlEscape(issue.Culprit)))
	}

	if issue.Metadata.Value != "" {
		sb.WriteString(fmt.Sprintf("💬 Error: <code>%s</code>\n", htmlEscape(issue.Metadata.Value)))
	}

	sb.WriteString(fmt.Sprintf("👁 Seen: %d time(s)\n", issue.TimesSeen))

	if issue.Permalink != "" {
		sb.WriteString(fmt.Sprintf("🔗 <a href=\"%s\">View in Sentry</a>", issue.Permalink))
	}

	return sb.String()
}

func htmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

func sendTelegram(token, chatID, message string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	payload := map[string]string{
		"chat_id":    chatID,
		"text":       message,
		"parse_mode": "HTML",
		"disable_web_page_preview": "true",
	}
	body, _ := json.Marshal(payload)

	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("telegram request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram returned %d: %s", resp.StatusCode, string(b))
	}
	return nil
}

func webhookHandler(token, chatID, secret, glitchtipURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Optional Sentry shared secret validation
		if secret != "" {
			if r.Header.Get("Sentry-Hook-Signature") != secret {
				log.Printf("WARN: invalid Sentry-Hook-Signature from %s", r.RemoteAddr)
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
		}

		body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20)) // 1 MB limit
		if err != nil {
			log.Printf("ERROR: reading body: %v", err)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		var event SentryEvent
		if err := json.Unmarshal(body, &event); err != nil {
			log.Printf("ERROR: parsing webhook event: %v", err)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		// Normalize GlitchTip's flat format into Sentry-style nested format
		normalizeEvent(&event, glitchtipURL)

		// Only forward actionable events
		if event.Action != "created" && event.Action != "resolved" {
			w.WriteHeader(http.StatusOK)
			return
		}

		msg := formatTelegramMessage(event)
		if err := sendTelegram(token, chatID, msg); err != nil {
			log.Printf("ERROR: sending Telegram message: %v", err)
			http.Error(w, "upstream error", http.StatusBadGateway)
			return
		}

		log.Printf("INFO: forwarded Sentry event action=%s project=%s title=%s",
			event.Action, event.Data.Issue.Project.Name, event.Data.Issue.Title)
		w.WriteHeader(http.StatusOK)
	}
}

func main() {
	token := os.Getenv("ALERT_BOT_TOKEN")
	chatID := os.Getenv("ALERT_CHAT_ID")
	secret := os.Getenv("SENTRY_SECRET")
	glitchtipURL := os.Getenv("GLITCHTIP_URL")
	port := os.Getenv("PORT")

	if token == "" || chatID == "" {
		log.Fatal("ALERT_BOT_TOKEN and ALERT_CHAT_ID must be set")
	}
	if port == "" {
		port = "9456"
	}
	if glitchtipURL == "" {
		glitchtipURL = "http://localhost:8100"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", webhookHandler(token, chatID, secret, glitchtipURL))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	})

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("INFO: sentry-relay listening on :%s", port)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("FATAL: %v", err)
	}
}
