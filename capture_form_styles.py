from seleniumwire import webdriver
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.common.by import By
import json, time, random

# Settings
START_URL = "https://www.webstaurantstore.com/choice-2-1-2-mexican-flag-food-pick/500PKFLAGMXCASE.html"
OUT_REQUESTS = "capture_all.json"
OUT_FORMS = "form_fields.json"

def z(a=0.6,b=1.3): time.sleep(random.uniform(a,b))

# Browser setup
options = Options()
options.binary_location = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
options.add_argument("--disable-blink-features=AutomationControlled")
options.add_experimental_option("excludeSwitches", ["enable-automation"])
options.add_experimental_option("useAutomationExtension", False)

driver = webdriver.Chrome(options=options)

# === Step 1: capture requests while you manually checkout ===
print("[INFO] Opening product page...")
driver.get(START_URL)
print("[INFO] Please go through Add to Cart, Checkout, Sign In, Billing manually.")
print("[INFO] Close the browser window when you are done.")

try:
    while True:
        time.sleep(2)
except KeyboardInterrupt:
    pass
except Exception:
    pass
finally:
    # === Step 2: Save network capture ===
    dump = []
    for req in driver.requests:
        if req.response:
            try:
                body = None
                if req.body:
                    try:
                        body = req.body.decode("utf-8", errors="ignore")
                    except:
                        body = str(req.body)
                dump.append({
                    "method": req.method,
                    "url": req.url,
                    "headers": dict(req.headers),
                    "body": body,
                    "status": req.response.status_code,
                    "response_headers": dict(req.response.headers),
                })
            except Exception:
                continue
    with open(OUT_REQUESTS, "w", encoding="utf-8") as f:
        json.dump(dump, f, indent=2)
    print(f"[SAVED] {OUT_REQUESTS} with {len(dump)} requests.")

    # === Step 3: Extract form fields (shipping & billing) ===
    results = {"shipping": [], "billing": []}
    try:
        # Find inputs
        inputs = driver.find_elements(By.CSS_SELECTOR, "input, select, textarea")
        for el in inputs:
            try:
                if not el.is_displayed(): 
                    continue
                desc = {
                    "tag": el.tag_name,
                    "id": el.get_attribute("id"),
                    "name": el.get_attribute("name"),
                    "type": el.get_attribute("type"),
                    "class": el.get_attribute("class"),
                    "placeholder": el.get_attribute("placeholder"),
                    "aria-label": el.get_attribute("aria-label"),
                }
                # Guess if shipping/billing based on name/id/class
                label = "shipping" if any(k in (desc["id"]+desc["name"]+desc["class"]).lower() 
                                          for k in ["ship"]) else "billing" if any(k in (desc["id"]+desc["name"]+desc["class"]).lower() 
                                          for k in ["bill"]) else "unknown"
                if label in results:
                    results[label].append(desc)
            except:
                continue
        with open(OUT_FORMS, "w", encoding="utf-8") as f:
            json.dump(results, f, indent=2)
        print(f"[SAVED] {OUT_FORMS} with {sum(len(v) for v in results.values())} fields.")
    except Exception:
        print("[WARN] Could not extract form fields (maybe window closed too early).")

    driver.quit()