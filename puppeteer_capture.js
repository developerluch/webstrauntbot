const puppeteer = require('puppeteer');
const fs = require('fs');
const path = require('path');

class WebstaurantCapture {
    constructor() {
        this.requests = [];
        this.responses = [];
        this.cookies = [];
        this.endpoints = new Map();
        this.tokens = new Map();
    }

    async init() {
        this.browser = await puppeteer.launch({
            headless: true,
            args: [
                '--no-sandbox',
                '--disable-setuid-sandbox',
                '--disable-dev-shm-usage',
                '--disable-accelerated-2d-canvas',
                '--no-first-run',
                '--no-zygote',
                '--disable-gpu',
                '--disable-web-security',
                '--disable-features=VizDisplayCompositor',
                '--disable-background-timer-throttling',
                '--disable-backgrounding-occluded-windows',
                '--disable-renderer-backgrounding'
            ]
        });

        this.page = await this.browser.newPage();

        // Set user agent to avoid detection
        await this.page.setUserAgent('Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36');

        // Enable request interception
        await this.page.setRequestInterception(true);

        // Capture all requests
        this.page.on('request', (request) => {
            const requestData = {
                url: request.url(),
                method: request.method(),
                headers: request.headers(),
                postData: request.postData(),
                timestamp: new Date().toISOString(),
                resourceType: request.resourceType()
            };

            this.requests.push(requestData);

            // Extract potential tokens from headers
            this.extractTokens(request.headers(), request.url());

            // Classify endpoints
            this.classifyEndpoint(request.url(), request.method());

            request.continue();
        });

        // Capture all responses
        this.page.on('response', (response) => {
            const responseData = {
                url: response.url(),
                status: response.status(),
                headers: response.headers(),
                timestamp: new Date().toISOString()
            };

            this.responses.push(responseData);
        });

        return this;
    }

    extractTokens(headers, url) {
        // Extract common tokens
        const tokenPatterns = [
            { name: 'csrf_token', patterns: [/csrf[_-]?token/i, /xsrf[_-]?token/i] },
            { name: 'session_id', patterns: [/session[_-]?id/i, /session/i] },
            { name: 'auth_token', patterns: [/auth[_-]?token/i, /bearer/i] },
            { name: 'api_key', patterns: [/api[_-]?key/i, /apikey/i] },
            { name: 'jwt', patterns: [/authorization/i] },
            { name: 'cart_token', patterns: [/cart[_-]?token/i] },
            { name: 'user_token', patterns: [/user[_-]?token/i] }
        ];

        for (const [headerName, headerValue] of Object.entries(headers)) {
            for (const tokenType of tokenPatterns) {
                for (const pattern of tokenType.patterns) {
                    if (pattern.test(headerName) || pattern.test(headerValue)) {
                        if (!this.tokens.has(tokenType.name)) {
                            this.tokens.set(tokenType.name, []);
                        }
                        this.tokens.get(tokenType.name).push({
                            header: headerName,
                            value: headerValue,
                            url: url,
                            timestamp: new Date().toISOString()
                        });
                    }
                }
            }
        }
    }

    classifyEndpoint(url, method) {
        const endpointPatterns = [
            { category: 'cart', patterns: [/\/cart/i, /\/basket/i, /\/shopping/i] },
            { category: 'checkout', patterns: [/\/checkout/i, /\/payment/i, /\/billing/i, /\/shipping/i] },
            { category: 'product', patterns: [/\/product/i, /\/item/i, /\/catalog/i] },
            { category: 'api', patterns: [/\/api/i, /\/graphql/i, /\/ajax/i] },
            { category: 'auth', patterns: [/\/login/i, /\/signin/i, /\/auth/i, /\/session/i] },
            { category: 'search', patterns: [/\/search/i, /\/find/i] }
        ];

        for (const endpointType of endpointPatterns) {
            for (const pattern of endpointType.patterns) {
                if (pattern.test(url)) {
                    if (!this.endpoints.has(endpointType.category)) {
                        this.endpoints.set(endpointType.category, []);
                    }
                    this.endpoints.get(endpointType.category).push({
                        url: url,
                        method: method,
                        timestamp: new Date().toISOString()
                    });
                    return;
                }
            }
        }

        // Default category
        if (!this.endpoints.has('other')) {
            this.endpoints.set('other', []);
        }
        this.endpoints.get('other').push({
            url: url,
            method: method,
            timestamp: new Date().toISOString()
        });
    }

    async captureCheckoutFlow() {
        try {
            console.log('🚀 Starting WebstaurantStore checkout capture...');

            // Navigate to product page
            const productUrl = 'https://www.webstaurantstore.com/choice-2-1-2-mexican-flag-food-pick/500PKFLAGMXCASE.html';
            console.log(`📦 Loading product page: ${productUrl}`);
            await this.page.goto(productUrl, { waitUntil: 'networkidle2' });

            // Wait for page to load and extract any initial data
            await this.page.waitForTimeout(2000);

            // Try to add item to cart
            console.log('🛒 Attempting to add item to cart...');
            try {
                // Look for add to cart button
                const addToCartSelectors = [
                    'button#buyButton',
                    'button[name="buyButton"]',
                    'input[type="submit"][value*="Add to Cart"]',
                    'button[type="submit"]',
                    '.add-to-cart',
                    '[data-testid*="add-to-cart"]'
                ];

                for (const selector of addToCartSelectors) {
                    try {
                        await this.page.waitForSelector(selector, { timeout: 2000 });
                        await this.page.click(selector);
                        console.log(`✅ Clicked add to cart button: ${selector}`);
                        break;
                    } catch (e) {
                        // Continue to next selector
                    }
                }
            } catch (e) {
                console.log('⚠️ Could not find add to cart button, continuing...');
            }

            await this.page.waitForTimeout(3000);

            // Navigate to cart
            console.log('🛒 Navigating to cart page...');
            await this.page.goto('https://www.webstaurantstore.com/cart/', { waitUntil: 'networkidle2' });
            await this.page.waitForTimeout(2000);

            // Try to proceed to checkout
            console.log('💳 Attempting to proceed to checkout...');
            try {
                const checkoutSelectors = [
                    'a[href*="checkout"]',
                    'button[name="checkout"]',
                    'input[type="submit"][value*="Checkout"]',
                    '.checkout-button',
                    '[data-testid*="checkout"]'
                ];

                for (const selector of checkoutSelectors) {
                    try {
                        await this.page.waitForSelector(selector, { timeout: 2000 });
                        await this.page.click(selector);
                        console.log(`✅ Clicked checkout button: ${selector}`);
                        break;
                    } catch (e) {
                        // Continue to next selector
                    }
                }
            } catch (e) {
                console.log('⚠️ Could not find checkout button, continuing...');
            }

            await this.page.waitForTimeout(3000);

            // Navigate to checkout page if not already there
            if (!this.page.url().includes('/shipping-billinginfo.cfm')) {
                console.log('📋 Navigating to checkout form...');
                await this.page.goto('https://www.webstaurantstore.com/shipping-billinginfo.cfm', { waitUntil: 'networkidle2' });
                await this.page.waitForTimeout(2000);
            }

            // Extract cookies
            this.cookies = await this.page.cookies();

            console.log('✅ Checkout flow capture completed');

        } catch (error) {
            console.error('❌ Error during checkout flow:', error.message);
        }
    }

    async saveResults() {
        const results = {
            timestamp: new Date().toISOString(),
            summary: {
                totalRequests: this.requests.length,
                totalResponses: this.responses.length,
                totalCookies: this.cookies.length,
                endpointCategories: Array.from(this.endpoints.keys()),
                tokenTypes: Array.from(this.tokens.keys())
            },
            endpoints: Object.fromEntries(this.endpoints),
            tokens: Object.fromEntries(this.tokens),
            cookies: this.cookies,
            recentRequests: this.requests.slice(-10), // Last 10 requests
            recentResponses: this.responses.slice(-10) // Last 10 responses
        };

        const outputFile = path.join(__dirname, 'puppeteer_capture_results.json');
        fs.writeFileSync(outputFile, JSON.stringify(results, null, 2));

        console.log(`📊 Results saved to: ${outputFile}`);
        console.log(`📈 Captured ${this.requests.length} requests, ${this.responses.length} responses`);
        console.log(`🍪 Found ${this.cookies.length} cookies`);
        console.log(`🔗 Classified ${this.endpoints.size} endpoint categories`);
        console.log(`🔑 Found ${this.tokens.size} token types`);

        return results;
    }

    async cleanup() {
        if (this.browser) {
            await this.browser.close();
        }
    }

    // Print summary of captured data
    printSummary() {
        console.log('\n📊 CAPTURE SUMMARY:');
        console.log('='.repeat(50));

        console.log(`\n🔗 ENDPOINTS BY CATEGORY:`);
        for (const [category, endpoints] of this.endpoints) {
            console.log(`  ${category.toUpperCase()}: ${endpoints.length} endpoints`);
        }

        console.log(`\n🔑 TOKENS FOUND:`);
        for (const [tokenType, tokens] of this.tokens) {
            console.log(`  ${tokenType.toUpperCase()}: ${tokens.length} instances`);
        }

        console.log(`\n🍪 COOKIES CAPTURED: ${this.cookies.length}`);
        this.cookies.forEach(cookie => {
            console.log(`  ${cookie.name}: ${cookie.value.substring(0, 20)}...`);
        });

        console.log('\n📝 RECENT REQUESTS:');
        this.requests.slice(-3).forEach((req, i) => {
            console.log(`  ${i + 1}. ${req.method} ${req.url.substring(0, 60)}...`);
        });
    }
}

// Main execution
async function main() {
    const capture = new WebstaurantCapture();

    try {
        await capture.init();
        await capture.captureCheckoutFlow();
        const results = await capture.saveResults();
        capture.printSummary();

        console.log('\n✅ Puppeteer capture completed successfully!');
        console.log('Results saved to puppeteer_capture_results.json');

    } catch (error) {
        console.error('❌ Error during capture:', error);
    } finally {
        await capture.cleanup();
    }
}

// Export for use as module
module.exports = WebstaurantCapture;

// Run if called directly
if (require.main === module) {
    main();
}
