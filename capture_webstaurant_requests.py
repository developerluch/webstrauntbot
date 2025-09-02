# capture_commerce_requests.py
import json, random, time, re, traceback
from datetime import datetime
from selenium import webdriver
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC

# ========= SETTINGS =========
START_URL = "https://www.webstaurantstore.com/choice-2-1-2-mexican-flag-food-pick/500PKFLAGMXCASE.html"
HEADLESS = False
DO_SIGNIN_PROBE = True        # tries to open the sign-in form and click continue with a dummy email
DUMP_FILE = "capture_dump.json"
MAX_BODY_CHARS = 2000
# ============================

# Action patterns (liberal, site may change paths)
ACTIONS = {
    "fetch_product": [r"/graphql", r"/api/.*product", r"/ajax/.*product", r"/product.*price", r"/inventory", r"/availability"],
    "add_to_cart":   [r"/cart.*add", r"/cart/line", r"/ajax/.*cart", r"/api/.*cart", r"/cart\?"],
    "view_cart":     [r"^https://[^/]+/cart/?$", r"/cart(\?|$)", r"/api/.*cart"],
    "go_to_checkout":[r"/checkout", r"/checkout/start", r"/api/.*checkout", r"/ajax/.*checkout"],
    "sign_in":       [r"/login", r"/sign-?in", r"/session", r"/oauth", r"/identity", r"/api/.*auth", r"/auth"],
}

HIDE_HDRS = {"cookie","authorization","x-csrf-token","x-xsrf-token"}

def z(a=0.6,b=1.4): time.sleep(random.uniform(a,b))
def redact(h): return {k:("<redacted>" if k.lower() in HIDE_HDRS else v) for k,v in (h or {}).items()}
def matches(url, pats): return any(re.search(p, url, re.I) for p in pats)

def human_scroll(d, n=6):
    h = d.execute_script("return document.body.scrollHeight") or 2000
    for i in range(n):
        d.execute_script(f"window.scrollTo(0,{int((i+1)*h/(n+1))});"); z()

def enable_net(d):
    d.execute_cdp_cmd("Network.enable", {})
    d.execute_cdp_cmd("Page.addScriptToEvaluateOnNewDocument", {
        "source": "Object.defineProperty(navigator,'webdriver',{get:()=>undefined})"
    })

def perf(d):
    for e in d.get_log("performance"):
        try: yield json.loads(e["message"])["message"]
        except: pass

def collect(d):
    seq=0; reqs={}; resps={}
    for m in perf(d):
        name=m.get("method"); p=m.get("params",{})
        if name=="Network.requestWillBeSent":
            rid=p.get("requestId"); r=p.get("request",{})
            if rid:
                seq+=1
                reqs[rid]={
                    "seq":seq,
                    "url":r.get("url",""),
                    "method":r.get("method","GET"),
                    "headers":r.get("headers",{}),
                    "postData":r.get("postData")
                }
        elif name=="Network.responseReceived":
            rid=p.get("requestId"); r=p.get("response",{})
            if rid:
                resps[rid]={
                    "status":r.get("status"),
                    "headers":r.get("headers",{}),
                    "mimeType":r.get("mimeType")
                }
    return reqs, resps

def get_body(d, rid):
    try:
        rb = d.execute_cdp_cmd("Network.getResponseBody", {"requestId": rid})
        return rb.get("body",""), rb.get("base64Encoded", False)
    except: return "", False

def pick_best(cands):
    if not cands: return None
    # Prefer POST, then most recent (largest seq)
    cands = sorted(cands, key=lambda x: (0 if x["method"].upper()=="POST" else 1, -x["seq"]))
    return cands[0]

def print_request(title, req, resp, body, b64):
    print(f"\n## {title}")
    print(f"METHOD: {req['method']}")
    print(f"URL: {req['url']}")
    print("REQUEST_HEADERS:"); print(json.dumps(redact(req.get("headers")), indent=2))
    pd = req.get("postData")
    if pd:
        try:
            print("REQUEST_BODY_JSON:"); print(json.dumps(json.loads(pd), indent=2)[:MAX_BODY_CHARS])
        except:
            print("REQUEST_BODY_RAW:"); print((pd or "")[:MAX_BODY_CHARS])
    print(f"STATUS: {resp.get('status')}")
    print("RESPONSE_HEADERS:"); print(json.dumps(resp.get("headers",{}), indent=2))
    if body:
        if b64: print("RESPONSE_BODY_SNIPPET:\n[base64 not expanded]")
        else:
            snippet = body[:MAX_BODY_CHARS] + ("... <truncated>" if len(body)>MAX_BODY_CHARS else "")
            print("RESPONSE_BODY_SNIPPET:"); print(snippet)

def main():
    options = webdriver.ChromeOptions()
    if HEADLESS: options.add_argument("--headless=new")
    options.binary_location = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
    options.add_experimental_option("excludeSwitches", ["enable-automation"])
    options.add_experimental_option("useAutomationExtension", False)
    options.add_argument("--disable-blink-features=AutomationControlled")
    options.set_capability("goog:loggingPrefs", {"performance":"ALL"})

    d = webdriver.Chrome(options=options)
    w = WebDriverWait(d, 25)

    all_snapshots = []  # for JSON dump

    try:
        enable_net(d)

        # 1) PDP load → fetch_product
        print(f"[STEP] Open PDP: {START_URL}")
        d.get(START_URL); z(1.0,1.8); human_scroll(d, 5)
        reqs, resps = collect(d); all_snapshots.append({"phase":"pdp", "reqs":reqs, "resps":resps})

        # 2) Add to Cart
        print("[STEP] Add to Cart")
        for by,sel in [(By.CSS_SELECTOR,"button#buyButton"),
                       (By.XPATH,"//button[contains(., 'Add to Cart')]"),
                       (By.CSS_SELECTOR,"form[action*='cart'] button[type='submit']")]:
            try:
                btn = w.until(EC.element_to_be_clickable((by,sel)))
                d.execute_script("arguments[0].scrollIntoView({block:'center'})", btn); z(0.3,0.8)
                btn.click(); print("[OK] Add to Cart clicked"); break
            except: pass
        z(1.0,1.6)
        reqs2, resps2 = collect(d); all_snapshots.append({"phase":"add_to_cart", "reqs":reqs2, "resps":resps2})

        # 3) View Cart
        print("[STEP] View Cart")
        d.get("https://www.webstaurantstore.com/cart/"); z(1.0,1.6)
        reqs3, resps3 = collect(d); all_snapshots.append({"phase":"cart", "reqs":reqs3, "resps":resps3})

        # 4) Go to Checkout
        print("[STEP] Checkout")
        for by,sel in [(By.CSS_SELECTOR,"a[href*='/checkout'], button[name='checkout']"),
                       (By.XPATH,"//a[contains(.,'Checkout')] | //button[contains(.,'Checkout')]")]:
            try:
                ct = w.until(EC.element_to_be_clickable((by,sel))); ct.click(); print("[OK] Checkout clicked"); break
            except: pass
        z(1.2,1.8)
        reqs4, resps4 = collect(d); all_snapshots.append({"phase":"checkout", "reqs":reqs4, "resps":resps4})

        # 5) Sign-in probe (no credentials submitted)
        if DO_SIGNIN_PROBE:
            print("[STEP] Sign-in probe (no password submit)")
            # Try to open sign-in panel or focus email field if present
            # Common selectors; site may differ.
            try:
                # Try direct sign-in link if present
                for by,sel in [(By.CSS_SELECTOR,"a[href*='login'], a[href*='sign-in'], a[href*='signin']"),
                               (By.XPATH,"//a[contains(.,'Sign In') or contains(.,'Log In')]")]:
                    try:
                        link = d.find_element(by, sel)
                        d.execute_script("arguments[0].scrollIntoView({block:'center'})", link); z(0.2,0.6)
                        link.click(); z(1.0,1.4); break
                    except: pass
                # Try entering a placeholder email if field exists to trigger auth endpoints
                for by,sel in [(By.CSS_SELECTOR,"input[type='email']"),
                               (By.CSS_SELECTOR,"input[name*='email' i]"),
                               (By.XPATH,"//input[contains(@name,'email') or @type='email']")]:
                    try:
                        email = w.until(EC.presence_of_element_located((by,sel)))
                        email.clear(); email.send_keys("probe+no-login@example.com"); z(0.2,0.4)
                        # Click a continue/next/sign-in button if present (should still fail without password)
                        for by2,sel2 in [
                            (By.XPATH,"//button[contains(.,'Continue') or contains(.,'Sign In') or contains(.,'Next')]"),
                            (By.CSS_SELECTOR,"button[type='submit']")
                        ]:
                            try:
                                b = d.find_element(by2, sel2); b.click(); z(1.0,1.4); break
                            except: pass
                        break
                    except: pass
            except: pass
            reqs5, resps5 = collect(d); all_snapshots.append({"phase":"signin", "reqs":reqs5, "resps":resps5})

        # ---- Classification across all phases ----
        combined_reqs = {}
        combined_resps = {}
        for snap in all_snapshots:
            combined_reqs.update(snap["reqs"])
            combined_resps.update(snap["resps"])

        # Bucketize
        buckets = {k: [] for k in ACTIONS}
        for rid, rq in combined_reqs.items():
            url = rq["url"]; method = rq["method"]
            for action, pats in ACTIONS.items():
                if matches(url, pats):
                    buckets[action].append({**rq, "rid": rid})
                    break

        # Print best per action
        print("\n===== SUMMARY: Proper requests per action =====")
        for action in ["fetch_product","add_to_cart","view_cart","go_to_checkout","sign_in"]:
            best = pick_best(buckets[action])
            print(f"\n### {action}")
            if not best:
                print("not observed"); 
                continue
            rid = best["rid"]; resp = combined_resps.get(rid, {})
            body, b64 = get_body(d, rid)
            print_request(action, best, resp, body, b64)

        # Save full dump
        dump = {
            "created_at": datetime.utcnow().isoformat()+"Z",
            "phases": [
                {"phase": s["phase"], "reqs": s["reqs"], "resps": s["resps"]}
                for s in all_snapshots
            ]
        }
        with open(DUMP_FILE, "w", encoding="utf-8") as f:
            json.dump(dump, f, indent=2)
        print(f"\n[INFO] Full capture saved to {DUMP_FILE}")

    except Exception:
        traceback.print_exc()
    finally:
        d.quit()

if __name__ == "__main__":
    main()