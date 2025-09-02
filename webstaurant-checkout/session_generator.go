package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

// SessionData holds session information including CSRF tokens and cookies
type SessionData struct {
	SessionID       string            `json:"session_id"`
	CFID            string            `json:"cfid"`
	CFToken         string            `json:"cftoken"`
	CSRFToken       string            `json:"csrf_token"`
	CFGLOBALS       string            `json:"cfglobals"`
	Cookies         map[string]string `json:"cookies"`
	LastUpdated     time.Time         `json:"last_updated"`
	CorrelationID   string            `json:"correlation_id"`
}

// WebstaurantClient represents the main client for WebstaurantStore operations
type WebstaurantClient struct {
	Client       interface{} // Simplified for session generator
	Session      *SessionData
	ProxyList    []string
	CurrentProxy int
	APIKey       string
	CFEndpoint   string
	Logger       *log.Logger
}

// SessionGenerator handles session creation and persistence
type SessionGenerator struct {
	Client       *WebstaurantClient
	Logger       *log.Logger
	SessionFile  string
}

// generateSessionFromCapturedData creates session data from captured request data
func (wc *WebstaurantClient) generateSessionFromCapturedData(capturedData map[string]interface{}) error {
	wc.Session = &SessionData{
		Cookies:     make(map[string]string),
		LastUpdated: time.Now(),
	}

	// Extract cookies from captured data
	if cookies, ok := capturedData["cookies"].(map[string]interface{}); ok {
		for name, value := range cookies {
			if val, ok := value.(string); ok {
				wc.Session.Cookies[name] = val

				// Extract specific session tokens
				switch name {
				case "SESSION_ID":
					wc.Session.SessionID = val
				case "CFID":
					wc.Session.CFID = val
				case "CFTOKEN":
					wc.Session.CFToken = val
				case "CFGLOBALS":
					wc.Session.CFGLOBALS = val
				case "CSRF_TOKEN":
					wc.Session.CSRFToken = val
				}
			}
		}
	}

	// Extract from response headers if available
	if headers, ok := capturedData["headers"].(map[string]interface{}); ok {
		if correlationID, ok := headers["correlation-id"].(string); ok {
			wc.Session.CorrelationID = correlationID
		}
	}

	wc.Logger.Printf("🔄 Generated session from captured data:")
	wc.Logger.Printf("   SessionID: %s", wc.Session.SessionID)
	wc.Logger.Printf("   CSRF Token: %s", wc.Session.CSRFToken)
	wc.Logger.Printf("   Cookies: %d", len(wc.Session.Cookies))

	return nil
}

// NewWebstaurantClient creates a new WebstaurantStore client instance
func NewWebstaurantClient(proxyList []string) *WebstaurantClient {
	return &WebstaurantClient{
		ProxyList:    proxyList,
		CurrentProxy: 0,
		APIKey:       "13e3814907f29f0c8c407b79e9e42ecd31cca082764c5ed92b47479717ccb81b",
		CFEndpoint:   "https://cloudfreed.com/solvereq",
		Logger:       log.New(log.Writer(), "[WebstaurantStore] ", log.LstdFlags),
	}
}

// NewSessionGenerator creates a new session generator
func NewSessionGenerator(proxyList []string) *SessionGenerator {
	return &SessionGenerator{
		Client:      NewWebstaurantClient(proxyList),
		Logger:      log.New(log.Writer(), "[SessionGenerator] ", log.LstdFlags),
		SessionFile: "generated_session.json",
	}
}

// GenerateSessionFromCapturedData creates a session using captured request data
func (sg *SessionGenerator) GenerateSessionFromCapturedData(capturedData map[string]interface{}) error {
	sg.Logger.Println("🔄 Generating session from captured data...")

	var data map[string]interface{}

	if capturedData == nil {
		// Use built-in test data
		data = map[string]interface{}{
			"cookies": map[string]interface{}{
				"SESSION_ID":      "81b95268-15a9-4fb0-b43f-9ff409446dea",
				"CFID":           "81b95268-15a9-4fb0-b43f-9ff409446dea",
				"CFTOKEN":        "81b95268-15a9-4fb0-b43f-9ff409446dea",
				"CSRF_TOKEN":     "4091420B276B2EC04C68208D55FF7242D89819B1",
				"CFGLOBALS":      "urltoken%3DCFID%23%3D81b95268%2D15a9%2D4fb0%2Db43f%2D9ff409446dea%26CFTOKEN%23%3D81b95268%2D15a9%2D4fb0%2Db43f%2D9ff409446dea%23lastvisit%3D%7Bts%20%272025%2D09%2D01%2008%3A41%3A34%27%7D%23hitcount%3D1%23timecreated%3D%7Bts%20%272025%2D09%2D01%2008%3A41%3A34%27%7D%23cftoken%3D81b95268%2D15a9%2D4fb0%2Db43f%2D9ff409446dea%23cfid%3D81b95268%2D15a9%2D4fb0%2Db43f%2D9ff409446dea%23",
				"REFERERSOURCE":  "%7B%22referer%22%3A%22https%3A%2F%2Fwww.google.com%2F%22%2C%22details%22%3A%7B%7D%2C%22entryDate%22%3A%2207%2F03%2F25%22%2C%22entryTime%22%3A%2211%3A44%22%7D",
				"DATACENTER_ID":  "2",
				"cf_clearance":   "979PIgca25beflvtiimajFIvTKRUTusChZQCrtn5Bwg-1756755536-1.2.1.1-bCIYhYEU7pJVlkL6S5j2WOvOUvn85nk8TDuto6brNDqAERejsMkkHteyxdynbCohThjBxtFCH8oBRh54LjOivfu_dz3o1P1TzkB9pr0xk5BWjnNcG3t94UC0QVbEBKKqXFP2Fnua593todf_gCuYSXwstF6_EbMT9hET9iBhDAVSYtJKN1oipVbXSdliF4TZ.3Azu.MOrnhAss7PVYyKKxlzgECmza2trT3xxTxQflM",
				"ContinueShopping": "https%3A%2F%2Fwww.webstaurantstore.com%2Fchoice-2-1-2-mexican-flag-food-pick%2F500PKFLAGMXCASE.html",
				"ajs_anonymous_id": "c03c881b-f7f9-473b-860e-8fba5fd585da",
				"_derived_epik": "dj0yJnU9UHZjTS1IVlFxUTB1MUFUOHRfM3dnWjJpS3hrYndTSlEmbj1BRmtMMjkwdHJCLUE2TGExTzlxLUR3Jm09NCZ0PUFBQUFBR2kyT1NRJnJtPTQmcnQ9QUFBQUFHaTJPU1Emc3A9Mg",
				"__cf_bm": "FNkEvfkeAPZ39js5rbSyCDk.84tZNOD8sh31fLM5aw8-1756773227-1.0.1.1-RWYgmTRX_exGrAPRzFhV36fnJsxzaq4I88VdQv9mMJz9DQqnstd.CHDPiTatURkZqUHdqWxK1I0ttrxXWjs.PoG2f8PUzaLT82_TVYxONU4",
				"_cfuvid": "8qZseveHIE0idmPkwxwKxkkBiF9Hr7lgCK.9defc.1A-1756773657224-0.0.1.1-604800000",
				"tzo": "-4",
				"tzn": "Eastern%20Daylight%20Time",
				"lang": "en-US",
				"HIDEPOSTREGISTRATIONPASSWORDFORM": "0",
				"WSS_TAG": "47908493",
				"GUESTCOUPONSIGNUP": "Vr4CU5j8iJ8WTcXrFZmOgtvsQgmYyReBNyHZXnObg9z3a24%2FKH9YOlRRNqFdcjLXjvI178trZ%2Fr9xO7yJyGJV9%2BptafnGRplVhOGH7haxzBiv1qJVqU8idwrooKzeAQsp4LjLik4AGKVbqLca7zmkEebrjHsEEmu1qygWGLQdIo%3D",
				"shipping_method": "Ground",
				"SHIPTOTYPE": "zDnIM1M8chWsC8z6R0y51w%3D%3D",
				"SHIPTOZIP": "fmf8rD%2BeQ8EGpK%2FoCNCmyQ%3D%3D",
				"SHIPTOCOUNTRY": "xkUE0jJJJA%2BXKlwsiOlKYA%3D%3D",
				"_uetsid": "97dbd9e0876a11f0b84d0b6b990e86f3|p2kkk5|2|fyz|0|2070",
				"_clsk": "g5y1r4%5E1756773668169%5E22%5E1%5Ey.clarity.ms%2Fcollect",
				"_uetvid": "998f0a10582411f0bb14dbacf15ad35a|o1baq3|1756773668574|25|1|bat.bing.com/p/insights/c/e",
				"_ga_ZFM16S3J5F": "GS2.1.s1756769304$o3$g1$t1756773693$j25$l0$h134404432",
			},
			"headers": map[string]interface{}{
				"correlation-id": "8ceecb6a-a260-4559-81e2-65e03b87afb6",
			},
		}
	} else {
		data = capturedData
	}

	// Generate session from captured data
	err := sg.Client.generateSessionFromCapturedData(data)
	if err != nil {
		return fmt.Errorf("failed to generate session: %v", err)
	}

	return nil
}

// SaveSession saves the current session to file
func (sg *SessionGenerator) SaveSession() error {
	if sg.Client.Session == nil {
		return fmt.Errorf("no session data to save")
	}

	data, err := json.MarshalIndent(sg.Client.Session, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %v", err)
	}

	err = os.WriteFile(sg.SessionFile, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write session file: %v", err)
	}

	sg.Logger.Printf("💾 Session saved to: %s", sg.SessionFile)
	return nil
}

// LoadSession loads session data from file
func (sg *SessionGenerator) LoadSession() error {
	data, err := os.ReadFile(sg.SessionFile)
	if err != nil {
		return fmt.Errorf("failed to read session file: %v", err)
	}

	sg.Client.Session = &SessionData{}
	if err := json.Unmarshal(data, sg.Client.Session); err != nil {
		return fmt.Errorf("failed to unmarshal session data: %v", err)
	}

	sg.Logger.Printf("📂 Session loaded from: %s", sg.SessionFile)
	return nil
}

// PrintSessionInfo prints detailed session information
func (sg *SessionGenerator) PrintSessionInfo() {
	if sg.Client.Session == nil {
		sg.Logger.Println("❌ No session data available")
		return
	}

	fmt.Println("\n🔐 SESSION INFORMATION:")
	fmt.Println("==================================================")
	fmt.Printf("📅 Last Updated: %s\n", sg.Client.Session.LastUpdated.Format(time.RFC3339))
	fmt.Printf("🆔 Session ID: %s\n", sg.Client.Session.SessionID)
	fmt.Printf("🔑 CSRF Token: %s\n", sg.Client.Session.CSRFToken)
	fmt.Printf("🆔 CFID: %s\n", sg.Client.Session.CFID)
	fmt.Printf("🔑 CFToken: %s\n", sg.Client.Session.CFToken)
	fmt.Printf("🔗 Correlation ID: %s\n", sg.Client.Session.CorrelationID)
	fmt.Printf("🍪 Cookies Count: %d\n", len(sg.Client.Session.Cookies))

	fmt.Println("\n🍪 COOKIE DETAILS:")
	for name, value := range sg.Client.Session.Cookies {
		if len(value) > 30 {
			fmt.Printf("  %s: %s...\n", name, value[:30])
		} else {
			fmt.Printf("  %s: %s\n", name, value)
		}
	}
}

// ValidateSession validates that all required session components are present
func (sg *SessionGenerator) ValidateSession() error {
	if sg.Client.Session == nil {
		return fmt.Errorf("no session data available")
	}

	requiredFields := []string{"SessionID", "CSRFToken", "CFID", "CFToken"}
	for _, field := range requiredFields {
		switch field {
		case "SessionID":
			if sg.Client.Session.SessionID == "" {
				return fmt.Errorf("missing required field: %s", field)
			}
		case "CSRFToken":
			if sg.Client.Session.CSRFToken == "" {
				return fmt.Errorf("missing required field: %s", field)
			}
		case "CFID":
			if sg.Client.Session.CFID == "" {
				return fmt.Errorf("missing required field: %s", field)
			}
		case "CFToken":
			if sg.Client.Session.CFToken == "" {
				return fmt.Errorf("missing required field: %s", field)
			}
		}
	}

	sg.Logger.Println("✅ Session validation passed")
	return nil
}

func main() {
	fmt.Println("🔧 WebstaurantStore Session Generator")
	fmt.Println("====================================")

	// Initialize session generator
	sg := NewSessionGenerator([]string{})

	// Generate session from captured data
	if err := sg.GenerateSessionFromCapturedData(nil); err != nil {
		log.Fatalf("Failed to generate session: %v", err)
	}

	// Validate session
	if err := sg.ValidateSession(); err != nil {
		log.Fatalf("Session validation failed: %v", err)
	}

	// Print session information
	sg.PrintSessionInfo()

	// Save session
	if err := sg.SaveSession(); err != nil {
		log.Fatalf("Failed to save session: %v", err)
	}

	fmt.Println("\n✅ Session generation completed successfully!")
	fmt.Printf("📄 Session data saved to: %s\n", sg.SessionFile)
	fmt.Println("\n💡 Use this session file with your WebstaurantStore automation scripts")
}
