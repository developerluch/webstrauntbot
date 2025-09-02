package main

import (
	"strings"
	"testing"
)

// TestNewWebstaurantClient tests client creation
func TestNewWebstaurantClient(t *testing.T) {
	proxies := []string{"http://proxy1:8080", "http://proxy2:8080"}
	client := NewWebstaurantClient(proxies)

	if client == nil {
		t.Fatal("NewWebstaurantClient returned nil")
	}

	if client.ProxyManager == nil {
		t.Fatal("ProxyManager should not be nil")
	}

	stats := client.GetProxyStats()
	if totalProxies := stats["total_proxies"].(int); totalProxies != 2 {
		t.Fatalf("Expected 2 proxies, got %d", totalProxies)
	}
}

// TestProductDataScraper tests product data extraction from HTML
func TestProductDataScraper(t *testing.T) {
	client := NewWebstaurantClient([]string{})

	// Test with empty HTML - should still work with fallbacks
	emptyHTML := ""
	productData, err := client.productDataScraper(emptyHTML)
	if err != nil {
		t.Fatalf("Failed to scrape product data from empty HTML: %v", err)
	}

	// Should use fallback values
	if productData.FeedIdentifier == "" {
		t.Fatal("FeedIdentifier should have fallback value")
	}

	if productData.ItemNumber == "" {
		t.Fatal("ItemNumber should have fallback value")
	}

	// Test with HTML containing product data
	htmlWithData := `
	<html>
		<head>
			<title>Test Product Page</title>
		</head>
		<body>
			<h1>Custom Test Product</h1>
			<input type="hidden" name="feed_identifier" value="CUSTOM123" />
			<input type="hidden" name="item_number" value="CUSTOM456" />
			<input type="hidden" name="price" value="39.99" />
		</body>
	</html>
	`

	productData, err = client.productDataScraper(htmlWithData)
	if err != nil {
		t.Fatalf("Failed to scrape product data: %v", err)
	}

	if productData.Name == "" {
		t.Fatal("Product name should not be empty")
	}

	if !strings.Contains(productData.Name, "Custom Test Product") {
		t.Fatalf("Expected product name to contain 'Custom Test Product', got: %s", productData.Name)
	}
}

// TestGenerateSessionFromCapturedData tests session generation from captured data
func TestGenerateSessionFromCapturedData(t *testing.T) {
	client := NewWebstaurantClient([]string{})

	capturedData := map[string]interface{}{
		"cookies": map[string]string{
			"SESSION_ID": "test-session-123",
			"CSRF_TOKEN": "test-csrf-456",
			"CFID":       "test-cfid-789",
			"CFTOKEN":    "test-cftoken-101",
		},
		"headers": map[string]interface{}{
			"correlation-id": "test-correlation-111",
		},
	}

	err := client.generateSessionFromCapturedData(capturedData)
	if err != nil {
		t.Fatalf("Failed to generate session from captured data: %v", err)
	}

	if client.Session == nil {
		t.Fatal("Session should not be nil after generation")
	}

	if client.Session.SessionID != "test-session-123" {
		t.Fatalf("Expected SessionID 'test-session-123', got: %s", client.Session.SessionID)
	}

	if client.Session.CSRFToken != "test-csrf-456" {
		t.Fatalf("Expected CSRFToken 'test-csrf-456', got: %s", client.Session.CSRFToken)
	}

	if client.Session.CorrelationID != "test-correlation-111" {
		t.Fatalf("Expected CorrelationID 'test-correlation-111', got: %s", client.Session.CorrelationID)
	}
}

// TestProxyRotation tests proxy list rotation
func TestProxyRotation(t *testing.T) {
	proxies := []string{"http://proxy1:8080", "http://proxy2:8080", "http://proxy3:8080"}
	client := NewWebstaurantClient(proxies)

	// Test initial proxy
	proxy1 := client.ProxyManager.GetNextProxy()
	if proxy1 != "http://proxy1:8080" {
		t.Fatalf("Expected first proxy 'http://proxy1:8080', got: %s", proxy1)
	}

	// Test rotation
	proxy2 := client.ProxyManager.GetNextProxy()
	if proxy2 != "http://proxy2:8080" {
		t.Fatalf("Expected second proxy 'http://proxy2:8080', got: %s", proxy2)
	}

	proxy3 := client.ProxyManager.GetNextProxy()
	if proxy3 != "http://proxy3:8080" {
		t.Fatalf("Expected third proxy 'http://proxy3:8080', got: %s", proxy3)
	}

	// Test wraparound
	proxy1Again := client.ProxyManager.GetNextProxy()
	if proxy1Again != "http://proxy1:8080" {
		t.Fatalf("Expected first proxy again 'http://proxy1:8080', got: %s", proxy1Again)
	}
}

// TestEmptyProxyList tests behavior with empty proxy list
func TestEmptyProxyList(t *testing.T) {
	client := NewWebstaurantClient([]string{})

	proxy := client.ProxyManager.GetNextProxy()
	if proxy != "" {
		t.Fatalf("Expected empty proxy for empty list, got: %s", proxy)
	}
}

// TestRandomDelay tests random delay generation (smoke test)
func TestRandomDelay(t *testing.T) {
	client := NewWebstaurantClient([]string{})

	// This is a basic smoke test - just ensure it doesn't panic
	client.randomDelay()
}

// TestSessionClearing tests session clearing functionality
func TestSessionClearing(t *testing.T) {
	client := NewWebstaurantClient([]string{})

	// Set up a session with data
	client.Session = &SessionData{
		SessionID:     "test-session",
		CSRFToken:     "test-csrf",
		CFID:          "test-cfid",
		CFToken:       "test-cftoken",
		Cookies:       map[string]string{"test": "value"},
		CorrelationID: "test-correlation",
		CFGLOBALS:     "test-globals",
	}

	// Clear session
	client.clearSession()

	if client.Session.SessionID != "" {
		t.Fatal("SessionID should be empty after clearing")
	}

	if client.Session.CSRFToken != "" {
		t.Fatal("CSRFToken should be empty after clearing")
	}

	if len(client.Session.Cookies) != 0 {
		t.Fatal("Cookies should be empty after clearing")
	}
}

// TestShippingInfo tests shipping info creation
func TestShippingInfo(t *testing.T) {
	client := NewWebstaurantClient([]string{})

	shippingInfo := client.getShippingInfo()
	if shippingInfo == nil {
		t.Fatal("getShippingInfo returned nil")
	}

	if shippingInfo.FirstName == "" {
		t.Fatal("FirstName should not be empty")
	}

	if shippingInfo.Email == "" {
		t.Fatal("Email should not be empty")
	}

	if shippingInfo.ZipCode == "" {
		t.Fatal("ZipCode should not be empty")
	}
}

// TestBillingInfo tests billing info creation
func TestBillingInfo(t *testing.T) {
	client := NewWebstaurantClient([]string{})

	billingInfo := client.getBillingInfo()
	if billingInfo == nil {
		t.Fatal("getBillingInfo returned nil")
	}

	if billingInfo.FirstName == "" {
		t.Fatal("FirstName should not be empty")
	}

	if billingInfo.Email == "" {
		t.Fatal("Email should not be empty")
	}

	if billingInfo.ZipCode == "" {
		t.Fatal("ZipCode should not be empty")
	}
}

// TestPaymentInfo tests payment info creation
func TestPaymentInfo(t *testing.T) {
	client := NewWebstaurantClient([]string{})

	paymentInfo := client.getPaymentInfo()
	if paymentInfo == nil {
		t.Fatal("getPaymentInfo returned nil")
	}

	if paymentInfo.CardNumber == "" {
		t.Fatal("CardNumber should not be empty")
	}

	if paymentInfo.ExpiryMonth == "" {
		t.Fatal("ExpiryMonth should not be empty")
	}

	if paymentInfo.CVV == "" {
		t.Fatal("CVV should not be empty")
	}

	if paymentInfo.CardholderName == "" {
		t.Fatal("CardholderName should not be empty")
	}
}
