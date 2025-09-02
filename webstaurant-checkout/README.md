# WebstaurantStore Checkout Automation Module

A comprehensive Go module for automating WebstaurantStore checkout operations with advanced anti-detection features and robust error handling.

## Features

- 🔐 **Advanced Session Management**: CSRF token handling, cookie persistence, and correlation ID tracking
- 💳 **Enhanced Payment Processing**: Uses the new `/checkout/processorder/?ignore_timeout=true` endpoint with error code 19 handling
- 📊 **Session Persistence**: Save and load session data for reuse across multiple runs
- 🛡️ **Comprehensive Security**: Proper CSRF token validation and anti-forgery protection
- ✅ **HTTP Client with Proxy Support**: Configurable proxy rotation for IP diversity
- ✅ **Cloudflare Challenge Solving**: Integrated API for bypassing protection
- ✅ **Rotating Proxy Support**: Automatic cycling through proxy lists
- ✅ **Comprehensive Error Handling**: Exponential backoff retries (up to 3 attempts)
- ✅ **Random Delay Adjustments**: 2-5 second delays to avoid detection
- ✅ **Product Data Scraping**: Regex-based extraction of product information
- ✅ **Complete Checkout Flow**: Automated sequence from add-to-cart to checkout
- ✅ **Detailed Logging**: Structured output for monitoring and debugging

## Installation

1. Clone or download the project
2. Navigate to the project directory
3. Install dependencies:
```bash
go mod tidy
```

## Session Management

### Session Data Structure
```go
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
```

### Session Generation Script
Generate and manage session data:
```bash
go run session_generator.go
```

### Session Persistence
```go
// Save session
err := client.saveSessionToFile("session.json")

// Load session
err := client.loadSessionFromFile("session.json")
```

## Usage

### Basic Example

```go
package main

import (
    "log"
    "webstaurant-checkout"
)

func main() {
    // Initialize proxy list (leave empty for no proxy)
    proxyList := []string{
        "http://proxy1:8080",
        "http://proxy2:8080",
        "http://proxy3:8080",
    }

    // Create client instance
    client := NewWebstaurantClient(proxyList)

    // Execute checkout flow
    productURL := "https://www.webstaurantstore.com/choice-2-1-2-mexican-flag-food-pick/500PKFLAGMXCASE.html"
    quantity := "1"
    isLoggedIn := false
    sessionID := ""

    err := client.checkoutFlow(productURL, quantity, isLoggedIn, sessionID)
    if err != nil {
        log.Fatalf("Checkout failed: %v", err)
    }

    log.Println("Checkout completed successfully!")
}
```

### Advanced Usage

```go
// With logged-in user
client := NewWebstaurantClient(proxyList)
err := client.checkoutFlow(
    "https://www.webstaurantstore.com/product-url",
    "5",           // quantity
    true,          // is logged in
    "session123",  // session ID
)
```

## API Reference

### WebstaurantClient

#### Methods

- `NewWebstaurantClient(proxyList []string) *WebstaurantClient`
  - Creates a new client instance with proxy rotation

- `createSession() error`
  - Initializes TLS session with custom fingerprinting

- `checkoutFlow(productURL, quantity string, isLoggedIn bool, sessionID string) error`
  - Executes complete checkout automation sequence with session management

- `processPayment(cartData *CartData, paymentInfo *PaymentInfo) (*CheckoutResponse, error)`
  - Processes payment using the new `/checkout/processorder/?ignore_timeout=true` endpoint
  - Handles error code 19 and other decline scenarios
  - Returns detailed error information for declined payments

- `addToCart(productURL, quantity string, productData *ProductData) error`
  - Adds item to cart using multipart form data

- `viewCart() error`
  - Retrieves current cart contents

- `mergeCart(sessionID, redirectURL string) error`
  - Merges guest cart with logged-in user cart

- `productDataScraper(htmlContent string) (*ProductData, error)`
  - Extracts product information from HTML using regex patterns

## TLS Configuration

The module uses highly specific TLS parameters to mimic legitimate browser behavior:

### Key Features:
- **Client Identifier**: Chrome 120
- **Random TLS Extension Order**: Enabled
- **Custom JA3 Fingerprint**: Chrome-like signature
- **HTTP/2 Settings**: Optimized for performance
- **Certificate Compression**: Brotli algorithm
- **Connection Flow**: Randomized between specific ranges

### Supported TLS Versions:
- GREASE
- TLS 1.3
- TLS 1.2

### Supported Curves:
- GREASE
- X25519
- secp256r1
- secp384r1
- Unknown curve 0x11EC

## Cloudflare Integration

Integrated with Cloudflare solver API:
- **Endpoint**: https://cloudfreed.com/solvereq
- **API Key**: Pre-configured
- **Auto-solving**: Triggers on 403/503 responses

## Proxy Rotation

- **Automatic Cycling**: Rotates through proxy list on each request
- **Format Support**: HTTP/HTTPS proxies
- **Fallback**: Works without proxies if list is empty

## Error Handling

### Retry Logic:
- **Max Retries**: 3 attempts
- **Exponential Backoff**: 2^attempt * base_delay
- **Base Delay**: 2 seconds

### Error Types Handled:
- Network timeouts
- HTTP errors (4xx, 5xx)
- Cloudflare challenges
- TLS handshake failures
- Proxy connection issues

## Logging

Comprehensive logging with structured output:
```
[WebstaurantStore] Starting checkout flow for: https://...
[WebstaurantStore] Session created successfully
[WebstaurantStore] Product data extracted: Product Name (ITEM123)
[WebstaurantStore] Waiting 3 seconds before next request...
[WebstaurantStore] Successfully added to cart: Product Name
[WebstaurantStore] Cart viewed successfully
[WebstaurantStore] Checkout flow completed successfully
```

## Payment Processing

### New Payment Endpoint
The module now uses the enhanced payment processing endpoint:
```
POST https://www.webstaurantstore.com/checkout/processorder/?ignore_timeout=true
```

### Error Code 19 Handling
Specifically handles WebstaurantStore's error code 19:
- **Error 19**: Payment declined - redirects to `/viewinfo.cfm?err=19`
- **Detection**: Monitors 302 redirects and location headers
- **Logging**: Detailed decline reason extraction from response body

### Decline Detection
Comprehensive decline indicator detection:
- HTTP status codes (402, 403, 422)
- Response body patterns ("declined", "insufficient funds", etc.)
- Redirect locations indicating errors
- Error code 19 specific handling

## Product Data Extraction

Uses regex patterns to extract:
- **Feed Identifier**: Product catalog ID
- **Item Number**: SKU/Product code
- **Price**: Current selling price
- **Product Name**: Display name

## Security Considerations

- Uses legitimate browser fingerprints
- Implements random delays to avoid rate limiting
- Supports proxy rotation for IP diversity
- Handles Cloudflare challenges automatically
- Includes comprehensive error handling

## Dependencies

- `github.com/PuerkitoBio/goquery v1.8.1` - HTML parsing and scraping
- Standard Go libraries: `net/http`, `encoding/json`, `regexp`, etc.

## License

This project is for educational and research purposes only. Ensure compliance with WebstaurantStore's terms of service and applicable laws.
