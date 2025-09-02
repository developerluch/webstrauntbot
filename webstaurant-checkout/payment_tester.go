package main

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"
)

// TestCard represents a test payment card with expected behavior
type TestCard struct {
	Number           string    `json:"number"`
	ExpiryMonth      string    `json:"expiry_month"`
	ExpiryYear       string    `json:"expiry_year"`
	CVV              string    `json:"cvv"`
	CardholderName   string    `json:"cardholder_name"`
	ExpectedResult   string    `json:"expected_result"`
	DeclineReason    string    `json:"decline_reason"`
	Description      string    `json:"description"`
	SuccessRate      float64   `json:"success_rate"`
	LastUsed         time.Time `json:"last_used"`
	UseCount         int       `json:"use_count"`
	SuccessCount     int       `json:"success_count"`
}

// PaymentTestScenario represents a complete payment testing scenario
type PaymentTestScenario struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	TestCards   []*TestCard `json:"test_cards"`
	Enabled     bool       `json:"enabled"`
	Priority    int        `json:"priority"`
}

// PaymentTester handles payment testing with various scenarios
type PaymentTester struct {
	Scenarios []*PaymentTestScenario `json:"scenarios"`
	Logger    *log.Logger           `json:"logger"`
}

// NewPaymentTester creates a new payment tester with predefined test scenarios
func NewPaymentTester(logger *log.Logger) *PaymentTester {
	tester := &PaymentTester{
		Scenarios: []*PaymentTestScenario{},
		Logger:    logger,
	}

	tester.initializeTestScenarios()
	return tester
}

// initializeTestScenarios sets up predefined payment testing scenarios
func (pt *PaymentTester) initializeTestScenarios() {
	// Scenario 1: Basic decline testing
	basicDecline := &PaymentTestScenario{
		ID:          "basic_decline",
		Name:        "Basic Decline Testing",
		Description: "Test various decline scenarios with different card types",
		Enabled:     true,
		Priority:    1,
		TestCards: []*TestCard{
			{
				Number:         "4000000000000002",
				ExpiryMonth:    "12",
				ExpiryYear:     "2025",
				CVV:            "123",
				CardholderName: "Test User",
				ExpectedResult: "DECLINE",
				DeclineReason:  "CARD_DECLINED",
				Description:    "Generic decline - card declined",
				SuccessRate:    1.0,
			},
			{
				Number:         "4000000000000127",
				ExpiryMonth:    "12",
				ExpiryYear:     "2025",
				CVV:            "123",
				CardholderName: "Test User",
				ExpectedResult: "DECLINE",
				DeclineReason:  "INSUFFICIENT_FUNDS",
				Description:    "Insufficient funds",
				SuccessRate:    1.0,
			},
			{
				Number:         "4000000000000069",
				ExpiryMonth:    "12",
				ExpiryYear:     "2025",
				CVV:            "123",
				CardholderName: "Test User",
				ExpectedResult: "DECLINE",
				DeclineReason:  "EXPIRED_CARD",
				Description:    "Expired card",
				SuccessRate:    1.0,
			},
		},
	}

	// Scenario 2: Fraud prevention testing
	fraudScenario := &PaymentTestScenario{
		ID:          "fraud_prevention",
		Name:        "Fraud Prevention Testing",
		Description: "Test fraud detection and prevention mechanisms",
		Enabled:     true,
		Priority:    2,
		TestCards: []*TestCard{
			{
				Number:         "4000000000000259",
				ExpiryMonth:    "12",
				ExpiryYear:     "2025",
				CVV:            "123",
				CardholderName: "Test User",
				ExpectedResult: "DECLINE",
				DeclineReason:  "FRAUD_DETECTED",
				Description:    "Fraud detected",
				SuccessRate:    0.95,
			},
			{
				Number:         "4000000000009235",
				ExpiryMonth:    "12",
				ExpiryYear:     "2025",
				CVV:            "123",
				CardholderName: "Test User",
				ExpectedResult: "DECLINE",
				DeclineReason:  "BLOCKED_CARD",
				Description:    "Card blocked",
				SuccessRate:    1.0,
			},
		},
	}

	// Scenario 3: Success testing (for comparison)
	successScenario := &PaymentTestScenario{
		ID:          "success_testing",
		Name:        "Success Testing",
		Description: "Test successful payment processing",
		Enabled:     true,
		Priority:    3,
		TestCards: []*TestCard{
			{
				Number:         "4111111111111111",
				ExpiryMonth:    "12",
				ExpiryYear:     "2025",
				CVV:            "123",
				CardholderName: "Test User",
				ExpectedResult: "SUCCESS",
				DeclineReason:  "",
				Description:    "Successful payment",
				SuccessRate:    0.9,
			},
			{
				Number:         "4000000000000077",
				ExpiryMonth:    "12",
				ExpiryYear:     "2025",
				CVV:            "123",
				CardholderName: "Test User",
				ExpectedResult: "SUCCESS",
				DeclineReason:  "",
				Description:    "Successful payment with processing delay",
				SuccessRate:    0.85,
			},
		},
	}

	// Scenario 4: Edge cases and error conditions
	edgeCaseScenario := &PaymentTestScenario{
		ID:          "edge_cases",
		Name:        "Edge Cases Testing",
		Description: "Test edge cases and unusual payment conditions",
		Enabled:     true,
		Priority:    4,
		TestCards: []*TestCard{
			{
				Number:         "4000000000000119",
				ExpiryMonth:    "12",
				ExpiryYear:     "2025",
				CVV:            "123",
				CardholderName: "Test User",
				ExpectedResult: "DECLINE",
				DeclineReason:  "PROCESSING_ERROR",
				Description:    "Processing error",
				SuccessRate:    0.8,
			},
			{
				Number:         "4242424242424242",
				ExpiryMonth:    "12",
				ExpiryYear:     "2025",
				CVV:            "123",
				CardholderName: "Test User",
				ExpectedResult: "SUCCESS",
				DeclineReason:  "",
				Description:    "Visa test card",
				SuccessRate:    0.95,
			},
		},
	}

	pt.Scenarios = append(pt.Scenarios, basicDecline, fraudScenario, successScenario, edgeCaseScenario)
	pt.Logger.Printf("✅ Initialized %d payment testing scenarios", len(pt.Scenarios))
}

// GetNextTestCard returns the next test card to use based on current scenario
func (pt *PaymentTester) GetNextTestCard(scenarioID string) *TestCard {
	for _, scenario := range pt.Scenarios {
		if scenario.ID == scenarioID && scenario.Enabled {
			if len(scenario.TestCards) == 0 {
				continue
			}

			// Simple round-robin selection for now
			selectedCard := scenario.TestCards[rand.Intn(len(scenario.TestCards))]
			selectedCard.LastUsed = time.Now()
			selectedCard.UseCount++

			return selectedCard
		}
	}

	// Default to basic decline scenario
	return pt.GetNextTestCard("basic_decline")
}

// GetRandomTestCard returns a random test card from enabled scenarios
func (pt *PaymentTester) GetRandomTestCard() *TestCard {
	enabledScenarios := []*PaymentTestScenario{}
	for _, scenario := range pt.Scenarios {
		if scenario.Enabled {
			enabledScenarios = append(enabledScenarios, scenario)
		}
	}

	if len(enabledScenarios) == 0 {
		return nil
	}

	selectedScenario := enabledScenarios[rand.Intn(len(enabledScenarios))]
	if len(selectedScenario.TestCards) == 0 {
		return nil
	}

	selectedCard := selectedScenario.TestCards[rand.Intn(len(selectedScenario.TestCards))]
	selectedCard.LastUsed = time.Now()
	selectedCard.UseCount++

	return selectedCard
}

// ValidatePaymentResult validates if the payment result matches expectations
func (pt *PaymentTester) ValidatePaymentResult(testCard *TestCard, actualResult string, actualReason string) *PaymentValidationResult {
	result := &PaymentValidationResult{
		TestCard:      testCard,
		ActualResult:  actualResult,
		ActualReason:  actualReason,
		Expected:      true,
		Timestamp:     time.Now(),
	}

	// Check if result matches expectation
	if strings.ToUpper(actualResult) == strings.ToUpper(testCard.ExpectedResult) {
		result.Expected = true
		result.Success = true
		testCard.SuccessCount++
	} else {
		result.Expected = false
		result.Success = false
	}

	// Log validation result
	if result.Success {
		pt.Logger.Printf("✅ Payment validation PASSED: %s - %s", testCard.Description, actualResult)
	} else {
		pt.Logger.Printf("❌ Payment validation FAILED: %s - Expected: %s, Got: %s",
			testCard.Description, testCard.ExpectedResult, actualResult)
	}

	return result
}

// PaymentValidationResult represents the result of payment validation
type PaymentValidationResult struct {
	TestCard     *TestCard `json:"test_card"`
	ActualResult string    `json:"actual_result"`
	ActualReason string    `json:"actual_reason"`
	Expected     bool      `json:"expected"`
	Success      bool      `json:"success"`
	Timestamp    time.Time `json:"timestamp"`
}

// GetTestCardByNumber finds a test card by its number
func (pt *PaymentTester) GetTestCardByNumber(cardNumber string) *TestCard {
	for _, scenario := range pt.Scenarios {
		for _, card := range scenario.TestCards {
			if card.Number == cardNumber {
				return card
			}
		}
	}
	return nil
}

// EnableScenario enables a specific testing scenario
func (pt *PaymentTester) EnableScenario(scenarioID string) error {
	for _, scenario := range pt.Scenarios {
		if scenario.ID == scenarioID {
			scenario.Enabled = true
			pt.Logger.Printf("✅ Enabled payment testing scenario: %s", scenario.Name)
			return nil
		}
	}
	return fmt.Errorf("scenario not found: %s", scenarioID)
}

// DisableScenario disables a specific testing scenario
func (pt *PaymentTester) DisableScenario(scenarioID string) error {
	for _, scenario := range pt.Scenarios {
		if scenario.ID == scenarioID {
			scenario.Enabled = false
			pt.Logger.Printf("🚫 Disabled payment testing scenario: %s", scenario.Name)
			return nil
		}
	}
	return fmt.Errorf("scenario not found: %s", scenarioID)
}

// GetEnabledScenarios returns all enabled scenarios
func (pt *PaymentTester) GetEnabledScenarios() []*PaymentTestScenario {
	enabled := []*PaymentTestScenario{}
	for _, scenario := range pt.Scenarios {
		if scenario.Enabled {
			enabled = append(enabled, scenario)
		}
	}
	return enabled
}

// GetScenarioStats returns statistics for all scenarios
func (pt *PaymentTester) GetScenarioStats() map[string]interface{} {
	stats := map[string]interface{}{
		"total_scenarios": len(pt.Scenarios),
		"enabled_scenarios": 0,
		"total_test_cards": 0,
		"scenarios": []map[string]interface{}{},
	}

	for _, scenario := range pt.Scenarios {
		if scenario.Enabled {
			stats["enabled_scenarios"] = stats["enabled_scenarios"].(int) + 1
		}

		scenarioStats := map[string]interface{}{
			"id": scenario.ID,
			"name": scenario.Name,
			"enabled": scenario.Enabled,
			"priority": scenario.Priority,
			"test_cards_count": len(scenario.TestCards),
		}

		totalUses := 0
		totalSuccesses := 0
		for _, card := range scenario.TestCards {
			totalUses += card.UseCount
			totalSuccesses += card.SuccessCount
		}

		scenarioStats["total_uses"] = totalUses
		scenarioStats["total_successes"] = totalSuccesses
		if totalUses > 0 {
			scenarioStats["success_rate"] = float64(totalSuccesses) / float64(totalUses)
		} else {
			scenarioStats["success_rate"] = 0.0
		}

		stats["total_test_cards"] = stats["total_test_cards"].(int) + len(scenario.TestCards)
		stats["scenarios"] = append(stats["scenarios"].([]map[string]interface{}), scenarioStats)
	}

	return stats
}

// RunPaymentTest runs a complete payment test scenario
func (pt *PaymentTester) RunPaymentTest(scenarioID string, client *WebstaurantClient, cartData *CartData) (*PaymentTestResult, error) {
	testCard := pt.GetNextTestCard(scenarioID)
	if testCard == nil {
		return nil, fmt.Errorf("no test card available for scenario: %s", scenarioID)
	}

	pt.Logger.Printf("🧪 Running payment test: %s with card ending in %s",
		testCard.Description, testCard.Number[len(testCard.Number)-4:])

	paymentInfo := &PaymentInfo{
		CardNumber:     testCard.Number,
		ExpiryMonth:    testCard.ExpiryMonth,
		ExpiryYear:     testCard.ExpiryYear,
		CVV:            testCard.CVV,
		CardholderName: testCard.CardholderName,
	}

	// Execute payment
	checkoutResp, err := client.processPayment(cartData, paymentInfo)

	result := &PaymentTestResult{
		ScenarioID:   scenarioID,
		TestCard:     testCard,
		PaymentInfo:  paymentInfo,
		Error:        err,
		Response:     checkoutResp,
		Timestamp:    time.Now(),
	}

	// Validate result if no error occurred
	if err == nil && checkoutResp != nil {
		validationResult := pt.ValidatePaymentResult(testCard,
			checkoutResp.Error, checkoutResp.Message)
		result.Validation = validationResult
	}

	return result, nil
}

// PaymentTestResult represents the result of a payment test
type PaymentTestResult struct {
	ScenarioID  string                    `json:"scenario_id"`
	TestCard    *TestCard                `json:"test_card"`
	PaymentInfo *PaymentInfo             `json:"payment_info"`
	Error       error                    `json:"error,omitempty"`
	Response    *CheckoutResponse        `json:"response,omitempty"`
	Validation  *PaymentValidationResult `json:"validation,omitempty"`
	Timestamp   time.Time                `json:"timestamp"`
}

// GenerateTestReport generates a comprehensive test report
func (pt *PaymentTester) GenerateTestReport() *PaymentTestReport {
	report := &PaymentTestReport{
		GeneratedAt: time.Now(),
		Scenarios:   []*PaymentTestScenario{},
		Stats:       pt.GetScenarioStats(),
	}

	// Deep copy scenarios with current state
	for _, scenario := range pt.Scenarios {
		scenarioCopy := *scenario
		report.Scenarios = append(report.Scenarios, &scenarioCopy)
	}

	return report
}

// PaymentTestReport represents a comprehensive payment testing report
type PaymentTestReport struct {
	GeneratedAt time.Time                `json:"generated_at"`
	Scenarios   []*PaymentTestScenario  `json:"scenarios"`
	Stats       map[string]interface{}  `json:"stats"`
}

// CreateDeclineTestCard creates a new test card for decline testing
func (pt *PaymentTester) CreateDeclineTestCard(number, expiryMonth, expiryYear, cvv, cardholderName, declineReason, description string) *TestCard {
	return &TestCard{
		Number:         number,
		ExpiryMonth:    expiryMonth,
		ExpiryYear:     expiryYear,
		CVV:            cvv,
		CardholderName: cardholderName,
		ExpectedResult: "DECLINE",
		DeclineReason:  declineReason,
		Description:    description,
		SuccessRate:    1.0,
	}
}

// CreateSuccessTestCard creates a new test card for success testing
func (pt *PaymentTester) CreateSuccessTestCard(number, expiryMonth, expiryYear, cvv, cardholderName, description string) *TestCard {
	return &TestCard{
		Number:         number,
		ExpiryMonth:    expiryMonth,
		ExpiryYear:     expiryYear,
		CVV:            cvv,
		CardholderName: cardholderName,
		ExpectedResult: "SUCCESS",
		DeclineReason:  "",
		Description:    description,
		SuccessRate:    0.9,
	}
}
