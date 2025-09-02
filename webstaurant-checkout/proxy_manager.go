package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// ProxyConfig holds configuration for proxy management
type ProxyConfig struct {
	Enabled        bool          `json:"enabled"`
	RotationMode   string        `json:"rotation_mode"`   // "round_robin", "random", "weighted"
	HealthCheck    bool          `json:"health_check"`
	CheckInterval  time.Duration `json:"check_interval"`
	MaxFailures    int           `json:"max_failures"`
	Timeout        time.Duration `json:"timeout"`
	RetryDelay     time.Duration `json:"retry_delay"`
}

// ProxyStatus represents the health status of a proxy
type ProxyStatus struct {
	URL         string
	LastUsed    time.Time
	LastChecked time.Time
	Failures    int
	ResponseTime time.Duration
	IsHealthy   bool
	Weight      int // For weighted rotation
}

// ProxyManager handles proxy rotation and health monitoring
type ProxyManager struct {
	Proxies     []*ProxyStatus
	Config      *ProxyConfig
	CurrentIndex int
	Logger      *log.Logger
	mu          sync.RWMutex
}

// NewProxyManager creates a new proxy manager with default configuration
func NewProxyManager(logger *log.Logger) *ProxyManager {
	config := &ProxyConfig{
		Enabled:       false,
		RotationMode:  "round_robin",
		HealthCheck:   true,
		CheckInterval: 30 * time.Second,
		MaxFailures:   3,
		Timeout:       10 * time.Second,
		RetryDelay:    5 * time.Second,
	}

	return &ProxyManager{
		Proxies: []*ProxyStatus{},
		Config:  config,
		Logger:  logger,
	}
}

// AddProxy adds a new proxy to the manager
func (pm *ProxyManager) AddProxy(proxyURL string, weight int) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	proxyStatus := &ProxyStatus{
		URL:       proxyURL,
		LastUsed:  time.Time{},
		Failures:  0,
		IsHealthy: true,
		Weight:    weight,
	}

	pm.Proxies = append(pm.Proxies, proxyStatus)
	pm.Logger.Printf("✅ Added proxy: %s (weight: %d)", proxyURL, weight)
}

// AddProxyList adds multiple proxies from a list
func (pm *ProxyManager) AddProxyList(proxies []string) {
	for _, proxy := range proxies {
		pm.AddProxy(proxy, 1) // Default weight of 1
	}
}

// LoadProxyConfig loads proxy configuration from environment or config file
func (pm *ProxyManager) LoadProxyConfig() {
	// Load from environment variables or config file
	// For now, using default configuration
	pm.Logger.Println("📋 Proxy configuration loaded with defaults")
}

// GetNextProxy returns the next proxy based on rotation mode
func (pm *ProxyManager) GetNextProxy() string {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if len(pm.Proxies) == 0 || !pm.Config.Enabled {
		return ""
	}

	var selectedProxy *ProxyStatus

	switch pm.Config.RotationMode {
	case "random":
		selectedProxy = pm.getRandomProxy()
	case "weighted":
		selectedProxy = pm.getWeightedProxy()
	default: // round_robin
		selectedProxy = pm.getRoundRobinProxy()
	}

	if selectedProxy != nil && selectedProxy.IsHealthy {
		selectedProxy.LastUsed = time.Now()
		pm.Logger.Printf("🔄 Selected proxy: %s", selectedProxy.URL)
		return selectedProxy.URL
	}

	// If no healthy proxy found, try to find any available proxy
	for _, proxy := range pm.Proxies {
		if proxy.IsHealthy {
			proxy.LastUsed = time.Now()
			pm.Logger.Printf("🔄 Fallback proxy: %s", proxy.URL)
			return proxy.URL
		}
	}

	pm.Logger.Println("⚠️ No healthy proxies available")
	return ""
}

// getRoundRobinProxy implements round-robin proxy selection
func (pm *ProxyManager) getRoundRobinProxy() *ProxyStatus {
	healthyProxies := pm.getHealthyProxies()
	if len(healthyProxies) == 0 {
		return nil
	}

	// Find next healthy proxy after current index
	startIndex := pm.CurrentIndex
	for i := 0; i < len(healthyProxies); i++ {
		index := (startIndex + i) % len(healthyProxies)
		if healthyProxies[index].IsHealthy {
			pm.CurrentIndex = (index + 1) % len(pm.Proxies)
			return healthyProxies[index]
		}
	}

	return nil
}

// getRandomProxy implements random proxy selection
func (pm *ProxyManager) getRandomProxy() *ProxyStatus {
	healthyProxies := pm.getHealthyProxies()
	if len(healthyProxies) == 0 {
		return nil
	}

	randomIndex := rand.Intn(len(healthyProxies))
	return healthyProxies[randomIndex]
}

// getWeightedProxy implements weighted random proxy selection
func (pm *ProxyManager) getWeightedProxy() *ProxyStatus {
	healthyProxies := pm.getHealthyProxies()
	if len(healthyProxies) == 0 {
		return nil
	}

	totalWeight := 0
	for _, proxy := range healthyProxies {
		totalWeight += proxy.Weight
	}

	if totalWeight == 0 {
		return pm.getRandomProxy()
	}

	randomValue := rand.Intn(totalWeight)
	currentWeight := 0

	for _, proxy := range healthyProxies {
		currentWeight += proxy.Weight
		if randomValue < currentWeight {
			return proxy
		}
	}

	return healthyProxies[len(healthyProxies)-1]
}

// getHealthyProxies returns a slice of healthy proxies
func (pm *ProxyManager) getHealthyProxies() []*ProxyStatus {
	var healthy []*ProxyStatus
	for _, proxy := range pm.Proxies {
		if proxy.IsHealthy {
			healthy = append(healthy, proxy)
		}
	}
	return healthy
}

// MarkProxyFailure marks a proxy as failed and potentially disables it
func (pm *ProxyManager) MarkProxyFailure(proxyURL string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for _, proxy := range pm.Proxies {
		if proxy.URL == proxyURL {
			proxy.Failures++
			proxy.LastChecked = time.Now()

			if proxy.Failures >= pm.Config.MaxFailures {
				proxy.IsHealthy = false
				pm.Logger.Printf("❌ Proxy disabled due to %d failures: %s", proxy.Failures, proxyURL)
			} else {
				pm.Logger.Printf("⚠️ Proxy failure %d/%d: %s", proxy.Failures, pm.Config.MaxFailures, proxyURL)
			}
			break
		}
	}
}

// MarkProxySuccess marks a proxy as healthy and resets failure count
func (pm *ProxyManager) MarkProxySuccess(proxyURL string, responseTime time.Duration) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for _, proxy := range pm.Proxies {
		if proxy.URL == proxyURL {
			proxy.IsHealthy = true
			proxy.Failures = 0
			proxy.LastChecked = time.Now()
			proxy.ResponseTime = responseTime
			pm.Logger.Printf("✅ Proxy healthy: %s (response: %v)", proxyURL, responseTime)
			break
		}
	}
}

// StartHealthCheck starts the background health check routine
func (pm *ProxyManager) StartHealthCheck() {
	if !pm.Config.HealthCheck {
		return
	}

	go func() {
		ticker := time.NewTicker(pm.Config.CheckInterval)
		defer ticker.Stop()

		for range ticker.C {
			pm.performHealthChecks()
		}
	}()

	pm.Logger.Println("🏥 Proxy health check routine started")
}

// performHealthChecks checks the health of all proxies
func (pm *ProxyManager) performHealthChecks() {
	pm.mu.RLock()
	proxies := make([]*ProxyStatus, len(pm.Proxies))
	copy(proxies, pm.Proxies)
	pm.mu.RUnlock()

	for _, proxy := range proxies {
		go pm.checkProxyHealth(proxy)
	}
}

// checkProxyHealth checks the health of a single proxy
func (pm *ProxyManager) checkProxyHealth(proxy *ProxyStatus) {
	start := time.Now()

	client := &http.Client{
		Timeout: pm.Config.Timeout,
	}

	// Parse proxy URL
	proxyURL, err := url.Parse(proxy.URL)
	if err != nil {
		pm.Logger.Printf("❌ Invalid proxy URL: %s", proxy.URL)
		return
	}

	client.Transport = &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}

	// Test with a simple request
	resp, err := client.Get("https://httpbin.org/ip")
	responseTime := time.Since(start)

	if err != nil {
		pm.MarkProxyFailure(proxy.URL)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		pm.MarkProxySuccess(proxy.URL, responseTime)
	} else {
		pm.MarkProxyFailure(proxy.URL)
	}
}

// GetStats returns proxy statistics
func (pm *ProxyManager) GetStats() map[string]interface{} {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	totalProxies := len(pm.Proxies)
	healthyProxies := 0
	totalFailures := 0

	for _, proxy := range pm.Proxies {
		if proxy.IsHealthy {
			healthyProxies++
		}
		totalFailures += proxy.Failures
	}

	return map[string]interface{}{
		"total_proxies":     totalProxies,
		"healthy_proxies":   healthyProxies,
		"unhealthy_proxies": totalProxies - healthyProxies,
		"total_failures":    totalFailures,
		"rotation_mode":     pm.Config.RotationMode,
		"enabled":          pm.Config.Enabled,
	}
}

// Enable enables proxy usage
func (pm *ProxyManager) Enable() {
	pm.mu.Lock()
	pm.Config.Enabled = true
	pm.mu.Unlock()
	pm.Logger.Println("🔓 Proxy manager enabled")
}

// Disable disables proxy usage
func (pm *ProxyManager) Disable() {
	pm.mu.Lock()
	pm.Config.Enabled = false
	pm.mu.Unlock()
	pm.Logger.Println("🔒 Proxy manager disabled")
}

// SetRotationMode sets the proxy rotation mode
func (pm *ProxyManager) SetRotationMode(mode string) error {
	validModes := []string{"round_robin", "random", "weighted"}

	for _, validMode := range validModes {
		if mode == validMode {
			pm.mu.Lock()
			pm.Config.RotationMode = mode
			pm.mu.Unlock()
			pm.Logger.Printf("🔄 Proxy rotation mode set to: %s", mode)
			return nil
		}
	}

	return fmt.Errorf("invalid rotation mode: %s. Valid modes: %v", mode, validModes)
}
