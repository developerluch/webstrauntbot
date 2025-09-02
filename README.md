# WebstaurantStore Checkout Automation System

A comprehensive, production-ready checkout automation system for WebstaurantStore with advanced features including proxy rotation, intelligent rate limiting, error monitoring, and payment testing.

## 🚀 Features

### Core Functionality
- **Complete Checkout Flow**: Automated product discovery, cart management, and checkout process
- **Session Management**: Advanced CSRF token handling and cookie management
- **Real-time Data Processing**: Live HTML parsing and form data extraction
- **Cloudflare Bypass**: Integrated Cloudflare challenge solving

### Advanced Features
- **🔄 Proxy Rotation**: Intelligent proxy management with health monitoring
- **⏱️ Rate Limiting**: Smart delays to mimic human behavior
- **📊 Error Monitoring**: Comprehensive error tracking and alerting
- **💳 Payment Testing**: Built-in test card framework for secure testing
- **🔍 Data Validation**: Enhanced product data extraction with validation
- **📈 Performance Monitoring**: Detailed statistics and logging

## 🏗️ Architecture

### Components
- **Go Application** (`webstaurant-checkout/`): Main automation engine
- **Python Utilities**: Network request capture and analysis
- **Node.js Tools**: Puppeteer-based web scraping utilities

### Key Modules
- `main.go`: Core checkout automation logic
- `proxy_manager.go`: Proxy rotation and health monitoring
- `error_monitor.go`: Error tracking and alerting system
- `payment_tester.go`: Test card management and validation
- `session_generator.go`: Session data generation utilities

## 📋 Prerequisites

### Go Requirements
- Go 1.19 or later
- Required packages:
  ```bash
  go mod tidy
  ```

### Python Requirements (Optional)
- Python 3.8+
- Selenium WebDriver
- Chrome/Chromium browser

### Node.js Requirements (Optional)
- Node.js 16+
- Puppeteer
- Chrome/Chromium browser

## 🛠️ Installation

1. **Clone the repository**:
   ```bash
   git clone https://github.com/yourusername/webstaurant-checkout.git
   cd webstaurant-checkout
   ```

2. **Install Go dependencies**:
   ```bash
   cd webstaurant-checkout
   go mod download
   ```

3. **Build the application**:
   ```bash
   go build .
   ```

## 🚀 Usage

### Basic Checkout Automation
```bash
./webstaurant-checkout
```

The application will:
1. Initialize session with captured credentials
2. Verify product URL accessibility
3. Execute complete checkout flow
4. Process payment with test cards
5. Generate comprehensive logs and reports

### Advanced Configuration

#### Proxy Setup
```go
proxyList := []string{
    "http://proxy1.example.com:8080",
    "http://proxy2.example.com:8080",
}
client := NewWebstaurantClient(proxyList)
```

#### Error Monitoring
```go
errorMonitor := NewErrorMonitor(logger)
errorMonitor.ConfigureEmail("smtp.example.com", 587, "user", "pass", "from@example.com", []string{"admin@example.com"})
```

#### Payment Testing
```go
paymentTester := NewPaymentTester(logger)
// Uses built-in test cards for secure testing
testCard := paymentTester.GetRandomTestCard()
```

## 🔧 Configuration

### Environment Variables
- `PROXY_LIST`: Comma-separated list of proxy URLs
- `EMAIL_SMTP_HOST`: SMTP server for error alerts
- `WEBHOOK_URL`: Webhook endpoint for error notifications

### Test Card Configuration
The system includes pre-configured test cards for different scenarios:
- Successful payments
- Declined transactions
- Insufficient funds
- Card expired
- Various error conditions

## 📊 Monitoring & Logging

### Error Monitoring
- Real-time error tracking
- Configurable alert thresholds
- Email and webhook notifications
- Persistent error state storage

### Performance Metrics
- Request response times
- Proxy health statistics
- Checkout success rates
- Error frequency analysis

## 🧪 Testing

### Unit Tests
```bash
cd webstaurant-checkout
go test -v
```

### Integration Tests
```bash
go test -v -tags=integration
```

### Payment Testing
```bash
# Run payment module tests
go test -run TestPayment -v
```

## 🔒 Security Features

### Data Protection
- No real credit card data storage
- Encrypted session management
- Secure API key handling
- Sensitive data exclusion from version control

### Compliance
- Industry-standard test card usage
- Secure proxy rotation
- Rate limiting to prevent abuse
- Comprehensive audit logging

## 📈 Performance

### Benchmarks
- **Checkout Completion**: < 45 seconds
- **Product Data Extraction**: < 2 seconds
- **Session Initialization**: < 1 second
- **Error Recovery**: Automatic with exponential backoff

### Optimization Features
- Connection pooling
- Intelligent caching
- Concurrent request handling
- Memory-efficient data structures

## 🔧 API Reference

### Core Functions

#### `NewWebstaurantClient(proxyList []string) *WebstaurantClient`
Creates a new client instance with proxy support.

#### `checkoutFlow(productURL, quantity string) (*CheckoutResponse, error)`
Executes the complete checkout process.

#### `GetProxyStats() map[string]interface{}`
Returns proxy performance statistics.

#### `GetErrorStats() map[string]interface{}`
Returns error monitoring statistics.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## ⚠️ Disclaimer

This software is for educational and testing purposes only. Ensure compliance with WebstaurantStore's Terms of Service and applicable laws. The authors are not responsible for misuse of this software.

## 📞 Support

For support, please open an issue on GitHub or contact the maintainers.

---

## 🏆 Acknowledgments

- Built with Go for high performance
- Inspired by modern e-commerce automation needs
- Designed for reliability and security

**Version**: 1.0.0
**Last Updated**: September 2025
