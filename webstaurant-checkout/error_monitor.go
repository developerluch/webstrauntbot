package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"sync"
	"time"
)

// ErrorLevel represents the severity of an error
type ErrorLevel int

const (
	ERROR_INFO ErrorLevel = iota
	ERROR_WARNING
	ERROR_ERROR
	ERROR_CRITICAL
)

// ErrorType represents different types of errors
type ErrorType int

const (
	ERROR_NETWORK ErrorType = iota
	ERROR_PROXY
	ERROR_AUTH
	ERROR_CHECKOUT
	ERROR_PAYMENT
	ERROR_PARSING
	ERROR_VALIDATION
	ERROR_TIMEOUT
)

// ErrorEvent represents a single error event
type ErrorEvent struct {
	ID          string                 `json:"id"`
	Timestamp   time.Time              `json:"timestamp"`
	Level       ErrorLevel             `json:"level"`
	Type        ErrorType              `json:"type"`
	Message     string                 `json:"message"`
	Details     map[string]interface{} `json:"details"`
	StackTrace  string                 `json:"stack_trace,omitempty"`
	Resolved    bool                   `json:"resolved"`
	ResolvedAt  *time.Time             `json:"resolved_at,omitempty"`
	RetryCount  int                    `json:"retry_count"`
}

// AlertConfig holds configuration for alerting
type AlertConfig struct {
	Enabled         bool          `json:"enabled"`
	EmailEnabled    bool          `json:"email_enabled"`
	WebhookEnabled  bool          `json:"webhook_enabled"`
	SMTPHost        string        `json:"smtp_host"`
	SMTPPort        int           `json:"smtp_port"`
	SMTPUsername    string        `json:"smtp_username"`
	SMTPPassword    string        `json:"smtp_password"`
	FromEmail       string        `json:"from_email"`
	ToEmails        []string      `json:"to_emails"`
	WebhookURL      string        `json:"webhook_url"`
	WebhookHeaders  map[string]string `json:"webhook_headers"`
	ErrorThresholds map[ErrorType]int  `json:"error_thresholds"`
	CooldownPeriod  time.Duration `json:"cooldown_period"`
}

// ErrorMonitor handles error tracking and alerting
type ErrorMonitor struct {
	Events        []*ErrorEvent      `json:"events"`
	Config        *AlertConfig       `json:"config"`
	Logger        *log.Logger        `json:"logger"`
	EventCounts   map[ErrorType]int  `json:"event_counts"`
	LastAlertTime map[ErrorType]time.Time `json:"last_alert_time"`
	ErrorFile     string             `json:"error_file"`
	mu            sync.RWMutex       `json:"-"`
}

// NewErrorMonitor creates a new error monitor
func NewErrorMonitor(logger *log.Logger) *ErrorMonitor {
	config := &AlertConfig{
		Enabled:        true,
		EmailEnabled:   false, // Disabled by default
		WebhookEnabled: false, // Disabled by default
		ErrorThresholds: map[ErrorType]int{
			ERROR_NETWORK:    5,
			ERROR_PROXY:      3,
			ERROR_AUTH:       2,
			ERROR_CHECKOUT:   2,
			ERROR_PAYMENT:    1,
			ERROR_PARSING:    10,
			ERROR_VALIDATION: 5,
			ERROR_TIMEOUT:    3,
		},
		CooldownPeriod: 5 * time.Minute,
	}

	return &ErrorMonitor{
		Events:        []*ErrorEvent{},
		Config:        config,
		Logger:        logger,
		EventCounts:   make(map[ErrorType]int),
		LastAlertTime: make(map[ErrorType]time.Time),
		ErrorFile:     "error_monitor.json",
	}
}

// ReportError reports a new error event
func (em *ErrorMonitor) ReportError(errorType ErrorType, level ErrorLevel, message string, details map[string]interface{}) {
	em.mu.Lock()
	defer em.mu.Unlock()

	event := &ErrorEvent{
		ID:         fmt.Sprintf("err_%d_%d", time.Now().Unix(), len(em.Events)),
		Timestamp:  time.Now(),
		Level:      level,
		Type:       errorType,
		Message:    message,
		Details:    details,
		Resolved:   false,
		RetryCount: 0,
	}

	em.Events = append(em.Events, event)
	em.EventCounts[errorType]++

	em.Logger.Printf("🚨 Error reported [%s]: %s", em.getErrorTypeString(errorType), message)

	// Check if alert should be sent
	if em.shouldSendAlert(errorType) {
		em.sendAlert(event)
		em.LastAlertTime[errorType] = time.Now()
		em.resetEventCount(errorType)
	}

	// Auto-resolve INFO level errors after some time
	if level == ERROR_INFO {
		go em.autoResolveError(event.ID, 30*time.Second)
	}
}

// ReportCheckoutFailure reports a checkout failure with specific details
func (em *ErrorMonitor) ReportCheckoutFailure(productURL, errorMessage string, errorCode string, retryCount int) {
	details := map[string]interface{}{
		"product_url": productURL,
		"error_code":  errorCode,
		"retry_count": retryCount,
		"stage":       "checkout",
	}

	em.ReportError(ERROR_CHECKOUT, ERROR_ERROR, fmt.Sprintf("Checkout failed: %s", errorMessage), details)
}

// ReportPaymentFailure reports a payment failure
func (em *ErrorMonitor) ReportPaymentFailure(cardLast4, errorMessage string, declineReason string) {
	details := map[string]interface{}{
		"card_last4":     cardLast4,
		"decline_reason": declineReason,
		"stage":          "payment",
	}

	em.ReportError(ERROR_PAYMENT, ERROR_CRITICAL, fmt.Sprintf("Payment failed: %s", errorMessage), details)
}

// ReportProxyFailure reports a proxy failure
func (em *ErrorMonitor) ReportProxyFailure(proxyURL string, errorMessage string) {
	details := map[string]interface{}{
		"proxy_url": proxyURL,
		"stage":     "proxy",
	}

	em.ReportError(ERROR_PROXY, ERROR_WARNING, fmt.Sprintf("Proxy failed: %s", errorMessage), details)
}

// ReportNetworkFailure reports a network failure
func (em *ErrorMonitor) ReportNetworkFailure(url, errorMessage string, statusCode int) {
	details := map[string]interface{}{
		"url":         url,
		"status_code": statusCode,
		"stage":       "network",
	}

	em.ReportError(ERROR_NETWORK, ERROR_WARNING, fmt.Sprintf("Network error: %s", errorMessage), details)
}

// shouldSendAlert determines if an alert should be sent for this error type
func (em *ErrorMonitor) shouldSendAlert(errorType ErrorType) bool {
	if !em.Config.Enabled {
		return false
	}

	count := em.EventCounts[errorType]
	threshold := em.Config.ErrorThresholds[errorType]

	if count < threshold {
		return false
	}

	// Check cooldown period
	if lastAlert, exists := em.LastAlertTime[errorType]; exists {
		if time.Since(lastAlert) < em.Config.CooldownPeriod {
			return false
		}
	}

	return true
}

// sendAlert sends an alert for the error event
func (em *ErrorMonitor) sendAlert(event *ErrorEvent) {
	em.Logger.Printf("📢 Sending alert for %s error: %s", em.getErrorTypeString(event.Type), event.Message)

	// Send email alert if enabled
	if em.Config.EmailEnabled {
		em.sendEmailAlert(event)
	}

	// Send webhook alert if enabled
	if em.Config.WebhookEnabled {
		em.sendWebhookAlert(event)
	}
}

// sendEmailAlert sends an email alert
func (em *ErrorMonitor) sendEmailAlert(event *ErrorEvent) {
	if !em.Config.EmailEnabled || len(em.Config.ToEmails) == 0 {
		return
	}

	subject := fmt.Sprintf("[%s] WebstaurantStore Error Alert", strings.ToUpper(em.getErrorTypeString(event.Type)))
	body := em.formatEmailBody(event)
	_ = subject // Subject is used in the email body

	auth := smtp.PlainAuth("", em.Config.SMTPUsername, em.Config.SMTPPassword, em.Config.SMTPHost)

	addr := fmt.Sprintf("%s:%d", em.Config.SMTPHost, em.Config.SMTPPort)
	err := smtp.SendMail(addr, auth, em.Config.FromEmail, em.Config.ToEmails, []byte(body))

	if err != nil {
		em.Logger.Printf("❌ Failed to send email alert: %v", err)
	} else {
		em.Logger.Printf("✅ Email alert sent to %v", em.Config.ToEmails)
	}
}

// sendWebhookAlert sends a webhook alert
func (em *ErrorMonitor) sendWebhookAlert(event *ErrorEvent) {
	if !em.Config.WebhookEnabled || em.Config.WebhookURL == "" {
		return
	}

	payload := map[string]interface{}{
		"alert_type": "error",
		"timestamp":  event.Timestamp.Format(time.RFC3339),
		"error_type": em.getErrorTypeString(event.Type),
		"level":      em.getErrorLevelString(event.Level),
		"message":    event.Message,
		"details":    event.Details,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		em.Logger.Printf("❌ Failed to marshal webhook payload: %v", err)
		return
	}

	req, err := http.NewRequest("POST", em.Config.WebhookURL, strings.NewReader(string(jsonData)))
	if err != nil {
		em.Logger.Printf("❌ Failed to create webhook request: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	// Add custom headers
	for key, value := range em.Config.WebhookHeaders {
		req.Header.Set(key, value)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		em.Logger.Printf("❌ Failed to send webhook alert: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		em.Logger.Printf("✅ Webhook alert sent successfully")
	} else {
		em.Logger.Printf("⚠️ Webhook alert failed with status: %d", resp.StatusCode)
	}
}

// formatEmailBody formats the email body for alerts
func (em *ErrorMonitor) formatEmailBody(event *ErrorEvent) string {
	body := fmt.Sprintf("Subject: WebstaurantStore Error Alert\r\n")
	body += fmt.Sprintf("From: %s\r\n", em.Config.FromEmail)
	body += fmt.Sprintf("To: %s\r\n", strings.Join(em.Config.ToEmails, ","))
	body += "\r\n"

	body += fmt.Sprintf("Error Alert\r\n")
	body += fmt.Sprintf("==============\r\n\r\n")

	body += fmt.Sprintf("Time: %s\r\n", event.Timestamp.Format(time.RFC3339))
	body += fmt.Sprintf("Type: %s\r\n", em.getErrorTypeString(event.Type))
	body += fmt.Sprintf("Level: %s\r\n", em.getErrorLevelString(event.Level))
	body += fmt.Sprintf("Message: %s\r\n", event.Message)

	if len(event.Details) > 0 {
		body += "\r\nDetails:\r\n"
		for key, value := range event.Details {
			body += fmt.Sprintf("  %s: %v\r\n", key, value)
		}
	}

	body += fmt.Sprintf("\r\nEvent ID: %s\r\n", event.ID)
	body += fmt.Sprintf("Total events of this type: %d\r\n", em.EventCounts[event.Type])

	return body
}

// getErrorTypeString returns string representation of error type
func (em *ErrorMonitor) getErrorTypeString(errorType ErrorType) string {
	switch errorType {
	case ERROR_NETWORK:
		return "NETWORK"
	case ERROR_PROXY:
		return "PROXY"
	case ERROR_AUTH:
		return "AUTH"
	case ERROR_CHECKOUT:
		return "CHECKOUT"
	case ERROR_PAYMENT:
		return "PAYMENT"
	case ERROR_PARSING:
		return "PARSING"
	case ERROR_VALIDATION:
		return "VALIDATION"
	case ERROR_TIMEOUT:
		return "TIMEOUT"
	default:
		return "UNKNOWN"
	}
}

// getErrorLevelString returns string representation of error level
func (em *ErrorMonitor) getErrorLevelString(level ErrorLevel) string {
	switch level {
	case ERROR_INFO:
		return "INFO"
	case ERROR_WARNING:
		return "WARNING"
	case ERROR_ERROR:
		return "ERROR"
	case ERROR_CRITICAL:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// resetEventCount resets the event count for a specific error type
func (em *ErrorMonitor) resetEventCount(errorType ErrorType) {
	em.EventCounts[errorType] = 0
}

// autoResolveError automatically resolves an error after a delay
func (em *ErrorMonitor) autoResolveError(eventID string, delay time.Duration) {
	time.Sleep(delay)

	em.mu.Lock()
	defer em.mu.Unlock()

	for _, event := range em.Events {
		if event.ID == eventID && !event.Resolved {
			now := time.Now()
			event.Resolved = true
			event.ResolvedAt = &now
			em.Logger.Printf("✅ Auto-resolved error: %s", eventID)
			break
		}
	}
}

// ResolveError manually resolves an error
func (em *ErrorMonitor) ResolveError(eventID string) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	for _, event := range em.Events {
		if event.ID == eventID {
			if event.Resolved {
				return fmt.Errorf("error already resolved")
			}
			now := time.Now()
			event.Resolved = true
			event.ResolvedAt = &now
			em.Logger.Printf("✅ Manually resolved error: %s", eventID)
			return nil
		}
	}

	return fmt.Errorf("error event not found: %s", eventID)
}

// GetStats returns error monitoring statistics
func (em *ErrorMonitor) GetStats() map[string]interface{} {
	em.mu.RLock()
	defer em.mu.RUnlock()

	stats := map[string]interface{}{
		"total_events":        len(em.Events),
		"unresolved_events":   0,
		"resolved_events":     0,
		"events_by_type":      make(map[string]int),
		"events_by_level":     make(map[string]int),
		"alerts_sent":         make(map[string]int),
	}

	for _, event := range em.Events {
		if event.Resolved {
			stats["resolved_events"] = stats["resolved_events"].(int) + 1
		} else {
			stats["unresolved_events"] = stats["unresolved_events"].(int) + 1
		}

		typeStr := em.getErrorTypeString(event.Type)
		levelStr := em.getErrorLevelString(event.Level)

		if eventsByType, ok := stats["events_by_type"].(map[string]int); ok {
			eventsByType[typeStr]++
		}

		if eventsByLevel, ok := stats["events_by_level"].(map[string]int); ok {
			eventsByLevel[levelStr]++
		}
	}

	return stats
}

// SaveToFile saves error monitor state to file
func (em *ErrorMonitor) SaveToFile() error {
	em.mu.RLock()
	defer em.mu.RUnlock()

	data, err := json.MarshalIndent(em, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal error monitor data: %v", err)
	}

	err = os.WriteFile(em.ErrorFile, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write error monitor file: %v", err)
	}

	em.Logger.Printf("💾 Error monitor state saved to: %s", em.ErrorFile)
	return nil
}

// LoadFromFile loads error monitor state from file
func (em *ErrorMonitor) LoadFromFile() error {
	data, err := os.ReadFile(em.ErrorFile)
	if err != nil {
		if os.IsNotExist(err) {
			em.Logger.Printf("📄 Error monitor file does not exist, starting fresh")
			return nil
		}
		return fmt.Errorf("failed to read error monitor file: %v", err)
	}

	em.mu.Lock()
	defer em.mu.Unlock()

	err = json.Unmarshal(data, em)
	if err != nil {
		return fmt.Errorf("failed to unmarshal error monitor data: %v", err)
	}

	em.Logger.Printf("📂 Error monitor state loaded from: %s", em.ErrorFile)
	return nil
}

// ConfigureEmail configures email alerting
func (em *ErrorMonitor) ConfigureEmail(smtpHost string, smtpPort int, username, password, fromEmail string, toEmails []string) {
	em.mu.Lock()
	defer em.mu.Unlock()

	em.Config.EmailEnabled = true
	em.Config.SMTPHost = smtpHost
	em.Config.SMTPPort = smtpPort
	em.Config.SMTPUsername = username
	em.Config.SMTPPassword = password
	em.Config.FromEmail = fromEmail
	em.Config.ToEmails = toEmails

	em.Logger.Printf("📧 Email alerting configured for: %v", toEmails)
}

// ConfigureWebhook configures webhook alerting
func (em *ErrorMonitor) ConfigureWebhook(webhookURL string, headers map[string]string) {
	em.mu.Lock()
	defer em.mu.Unlock()

	em.Config.WebhookEnabled = true
	em.Config.WebhookURL = webhookURL
	em.Config.WebhookHeaders = headers

	em.Logger.Printf("🔗 Webhook alerting configured: %s", webhookURL)
}

// SetErrorThreshold sets the error threshold for a specific error type
func (em *ErrorMonitor) SetErrorThreshold(errorType ErrorType, threshold int) {
	em.mu.Lock()
	defer em.mu.Unlock()

	em.Config.ErrorThresholds[errorType] = threshold
	em.Logger.Printf("📊 Error threshold for %s set to: %d", em.getErrorTypeString(errorType), threshold)
}
