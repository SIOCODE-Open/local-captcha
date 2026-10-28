# API Reference – SIOCODE Local Captcha

This document describes all API endpoints exposed by the SIOCODE Local Captcha service.

---

## 🧠 Health Check

### `GET /healthz`

Returns a health check response.

**Response (`200 OK`):**
```json
{
  "status": "OK"
}
```

---

## 📜 Injectable JavaScript SDK

### `GET /recaptcha/api.js`
### `GET /captcha.js`

Serves the browser JavaScript SDK that defines `window.grecaptcha` and renders `.g-recaptcha` widgets. Both paths are identical aliases.

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `render` | string | No | `"onload"` (default: auto-renders `.g-recaptcha` elements), `"explicit"` (skips auto-rendering), or `<sitekey>` (reCAPTCHA v3 mode: sets default sitekey for `grecaptcha.execute` and skips auto-rendering) |
| `onload` | string | No | Name of a global `window` callback function to invoke once `grecaptcha` is initialized and any auto-rendering has completed |
| `hl` | string | No | Language code (accepted for compatibility with Google reCAPTCHA script URLs) |

**Response Headers:**
- `Content-Type`: `application/javascript; charset=utf-8`

**Response (`200 OK`):**

Executable JavaScript that exposes `window.grecaptcha` with the following API:

#### `window.grecaptcha` Methods

1. **`grecaptcha.ready(callback)`**
   - Executes `callback()` asynchronously once the SDK and DOM are ready.

2. **`grecaptcha.render(container, parameters)`**
   - Renders a CAPTCHA widget inside `container` (either a DOM element ID string or an `HTMLElement`).
   - `parameters` is an optional configuration object. Any parameter omitted from `parameters` is read from `data-*` attributes on the container element:
     - `sitekey` / `data-sitekey` (string): The site key (defaults to `"default-sitekey"`).
     - `callback` / `data-callback` (function or string): Function or global function name called with `token` when solved.
     - `expired-callback` / `expiredCallback` / `data-expired-callback` (function or string): Function or global function name called when a token expires or is reset.
     - `error-callback` / `errorCallback` / `data-error-callback` (function or string): Function or global function name called when challenge/verification fails.
     - `size` / `data-size` (string): `"normal"` (default), `"compact"`, or `"invisible"`.
     - `theme` / `data-theme` (string): `"light"` (default) or `"dark"`.
     - `action` / `data-action` (string): Action name (default `"submit"`).
   - Creates a `<textarea name="g-recaptcha-response" class="g-recaptcha-response">` inside the container and embeds an `<iframe src="{base_url}/recaptcha/anchor?...">`.
   - Returns an integer `widgetId` (`0`, `1`, `2`, ...).
   - Throws an `Error` if `container` cannot be found in the DOM.

3. **`grecaptcha.execute([sitekeyOrWidgetId], [options])`**
   - Supports both **reCAPTCHA v3** (`grecaptcha.execute(sitekey, { action: 'login' })` or `grecaptcha.execute({ action: 'login' })` using the `render=<sitekey>` query parameter) and **reCAPTCHA v2 Invisible** (`grecaptcha.execute(widgetId)` or `grecaptcha.execute()` targeting widget `0`).
   - Sends a `POST {base_url}/challenge/solve` request with `{"sitekey": ..., "action": ...}`.
   - Returns a `Promise<string>` that resolves with the issued `token` string.
   - If called for a rendered widget, updates the widget's `<textarea name="g-recaptcha-response">` value and invokes the widget's `callback(token)` if configured.
   - If `/challenge/solve` returns `"success": false`, invokes the widget's `error-callback` (if configured) and rejects the returned Promise with an `Error`.

4. **`grecaptcha.getResponse([widgetId])`**
   - Returns the current response token string for `widgetId` (defaults to `0`), or `""` if the widget has not been solved or has been reset.

5. **`grecaptcha.reset([widgetId])`**
   - Clears the stored token and `<textarea name="g-recaptcha-response">` value for `widgetId` (defaults to `0`), and reloads the widget iframe.

---

## 🧩 Interactive Challenge UI & Token Issuance

### `GET /recaptcha/anchor`
### `GET /challenge`

Serves the interactive HTML widget / challenge page meant to be embedded inside an `<iframe>` or opened directly in a browser. Both paths are identical aliases.

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `k` | string | No | Site key (Google reCAPTCHA standard parameter) |
| `sitekey` | string | No | Alias for `k` (if both are omitted, defaults to `"default-sitekey"`) |
| `size` | string | No | Widget size: `"normal"` (default), `"compact"`, or `"invisible"` |
| `theme` | string | No | Widget theme: `"light"` (default) or `"dark"` |
| `action` | string | No | Action name associated with the challenge (default: `"submit"`) |
| `widget_id` | string | No | Widget identifier echoed back in `window.parent.postMessage` events |

**Response (`200 OK`):**
- `Content-Type`: `text/html; charset=utf-8`
- Renders an interactive HTML page containing:
  - A checkbox/anchor button (`#recaptcha-anchor`)
  - A challenge section (`#recaptcha-challenge` with `#solve-challenge-btn` and `#challenge-answer` when the effective mode is `"challenge"`)
  - A token display element (`#recaptcha-token`) that contains the issued token once solved
  - An error display element (`#recaptcha-error`) if in `"always_fail"` mode
  - `window.parent.postMessage` integration:
    - On solve: posts `{ type: "local-captcha-solved", widgetId: "<widget_id>", token: "<token>" }` to `window.parent`
    - On error: posts `{ type: "local-captcha-error", widgetId: "<widget_id>", errorCodes: [...] }` to `window.parent`

**Errors:**
- `400 Bad Request` - If `allow_any_site` is `false` and the provided `k`/`sitekey` is not registered in `sites` (and is not Google's official test sitekey `6LeIxAcTAAAAAJcZVRqyHh71UMIEGNQ_MXjiZKhI`).

---

### `POST /challenge/solve`
### `POST /recaptcha/reload`

Programmatically solves a challenge and issues a new CAPTCHA response token. Used by `grecaptcha.execute()`, the `/recaptcha/anchor` iframe UI, and automated tests. Both paths are identical aliases.

**Supported Content-Types:**
- `application/json`
- `application/x-www-form-urlencoded`
- Query parameters

**Request Parameters (JSON body, form fields, or query params):**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `sitekey` | string | No | Site key (alias `k` is also accepted; defaults to `"default-sitekey"`) |
| `action` | string | No | Action name for v3/enterprise tracking (default: `"submit"`) |
| `challenge_answer` | string | No | Answer to an interactive challenge. Any value is accepted except `"fail"`, which forces a challenge failure |
| `score` | number | No | Optional score override (`0.0` – `1.0`) for the issued token. Defaults to site's `default_score` or global `default_score` |
| `hostname` | string | No | Optional hostname override for the issued token. Defaults to site's `hostname` or global `hostname` |

**Example Request (`application/json`):**
```json
{
  "sitekey": "site-key-1",
  "action": "login"
}
```

**Success Response (`200 OK`):**
```json
{
  "success": true,
  "token": "tok_550e8400e29b41d4a716446655440000",
  "sitekey": "site-key-1",
  "action": "login",
  "score": 0.9,
  "hostname": "localhost",
  "challenge_ts": "2025-01-01T12:00:00Z",
  "expires_in": 300
}
```

**Failure Response (`200 OK` — when site/global `mode` is `"always_fail"`, `force_error_codes` is non-empty, or `challenge_answer` is `"fail"`):**
```json
{
  "success": false,
  "error-codes": [
    "invalid-input-response"
  ]
}
```

**Errors:**
- `400 Bad Request` - If `allow_any_site` is `false` and `sitekey` is missing or not found in configured `sites` (and is not Google's official test sitekey `6LeIxAcTAAAAAJcZVRqyHh71UMIEGNQ_MXjiZKhI`):
  ```json
  {
    "success": false,
    "error": "invalid-sitekey",
    "error-codes": [
      "invalid-sitekey"
    ]
  }
  ```

---

## ✅ CAPTCHA Verification API

### `POST /recaptcha/api/siteverify`
### `POST /siteverify`
### `POST /verify`

Verifies a CAPTCHA response token, mimicking Google's `https://www.google.com/recaptcha/api/siteverify` endpoint. All three paths are identical aliases. `GET` requests with query parameters are also supported on all three paths for compatibility with legacy reCAPTCHA clients.

**Supported Content-Types:**
- `application/x-www-form-urlencoded` (standard Google reCAPTCHA format)
- `application/json`
- Query parameters (`?secret=...&response=...&remoteip=...`)

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `secret` | string | Yes | The shared secret key between your site and local-captcha |
| `response` | string | Yes | The user response token provided by the client-side integration |
| `remoteip` | string | No | The user's IP address (optional, recorded in verification history) |

**Example Request (`application/x-www-form-urlencoded`):**
```http
POST /recaptcha/api/siteverify HTTP/1.1
Content-Type: application/x-www-form-urlencoded

secret=secret-key-1&response=tok_550e8400e29b41d4a716446655440000&remoteip=127.0.0.1
```

**Example Request (`application/json`):**
```json
{
  "secret": "secret-key-1",
  "response": "tok_550e8400e29b41d4a716446655440000",
  "remoteip": "127.0.0.1"
}
```

**Success Response (`200 OK`):**
```json
{
  "success": true,
  "challenge_ts": "2025-01-01T12:00:00Z",
  "hostname": "localhost",
  "score": 0.9,
  "action": "login"
}
```

**Failure Response (`200 OK`):**

Following the Google reCAPTCHA specification, verification failures always return HTTP `200 OK` with `"success": false` and `"error-codes"`:

```json
{
  "success": false,
  "error-codes": [
    "invalid-input-response"
  ]
}
```

### Verification Evaluation Rules & Error Codes

Every call to `siteverify` is recorded in the verification log (inspectable via `GET /testing/verifications`) and evaluated in the following exact order:

1. **Missing Parameters**:
   - If both `secret` and `response` are empty: returns `{"success": false, "error-codes": ["missing-input-secret", "missing-input-response"]}`.
   - If `secret` is empty: returns `{"success": false, "error-codes": ["missing-input-secret"]}`.
   - If `response` is empty: returns `{"success": false, "error-codes": ["missing-input-response"]}`.
2. **Global Forced Errors / Always-Fail Mode**:
   - If runtime `force_error_codes` is non-empty: returns `{"success": false, "error-codes": <force_error_codes>}`.
   - If runtime `mode` is `"always_fail"`: returns `{"success": false, "error-codes": ["invalid-input-response"]}`.
3. **Secret Validation**:
   - Checks if `secret` matches any configured site in `sites` or Google's official test secret (`6LeIxAcTAAAAAGG-vFI1TnRWxMZNFuojJ4WifJWe`).
   - If `allow_any_secret` is `false` and `secret` does not match a configured site (or Google's test secret): returns `{"success": false, "error-codes": ["invalid-input-secret"]}`.
   - If `secret` matches a configured site whose `mode` is `"always_fail"`: returns `{"success": false, "error-codes": <site.error_codes or ["invalid-input-response"]>}`.
4. **Built-in & Configured Magic Tokens**:
   - Magic tokens allow deterministic backend testing without issuing tokens first. They are not consumed by `verify_once`:
     | Magic Token (`response`) | Result (`success`) | `score` | `action` | `error-codes` |
     |--------------------------|--------------------|---------|----------|---------------|
     | `"always-pass-token"` | `true` | Site/global `default_score` | `"submit"` | *(omitted)* |
     | `"valid-token"` | `true` | Site/global `default_score` | `"submit"` | *(omitted)* |
     | `"low-score-token"` | `true` | `0.1` | `"submit"` | *(omitted)* |
     | `"high-score-token"` | `true` | `1.0` | `"submit"` | *(omitted)* |
     | `"always-fail-token"` | `false` | *(omitted)* | *(omitted)* | `["invalid-input-response"]` |
     | `"invalid-token"` | `false` | *(omitted)* | *(omitted)* | `["invalid-input-response"]` |
     | `"expired-token"` | `false` | *(omitted)* | *(omitted)* | `["timeout-or-duplicate"]` |
     | `"duplicate-token"` | `false` | *(omitted)* | *(omitted)* | `["timeout-or-duplicate"]` |
     | `"bad-request-token"` | `false` | *(omitted)* | *(omitted)* | `["bad-request"]` |
   - Custom entries in `magic_tokens` from configuration are also evaluated here.
5. **Issued / Pre-seeded Token Validation**:
   - Looks up `response` in the in-memory store of issued tokens.
   - If the token does not exist: returns `{"success": false, "error-codes": ["invalid-input-response"]}`.
   - If the token's `sitekey` matches a configured site in `sites` with a defined `secret`, and the request's `secret` does not match that site's `secret` (and is not Google's test secret `6LeIxAcTAAAAAGG-vFI1TnRWxMZNFuojJ4WifJWe`): returns `{"success": false, "error-codes": ["invalid-input-secret"]}`.
   - If the token's site has `mode: "always_fail"`: returns `{"success": false, "error-codes": <site.error_codes or ["invalid-input-response"]>}`.
   - If the token was pre-seeded with `error_codes`: returns `{"success": false, "error-codes": <token.error_codes>}`.
   - If the token is expired (`expired == true` or current time is past `expires_at`): returns `{"success": false, "error-codes": ["timeout-or-duplicate"]}`.
   - If `verify_once` is `true` and the token has already been verified (`used == true`): returns `{"success": false, "error-codes": ["timeout-or-duplicate"]}`.
   - Marks the token as `used = true` and returns `{"success": true, "challenge_ts": ..., "hostname": ..., "score": ..., "action": ...}`.

### Standard Google reCAPTCHA Test Keys

For drop-in compatibility with existing test suites, Google's official test credentials are always recognized even when `allow_any_site: false` or `allow_any_secret: false`:
- **Test Sitekey**: `6LeIxAcTAAAAAJcZVRqyHh71UMIEGNQ_MXjiZKhI`
- **Test Secret**: `6LeIxAcTAAAAAGG-vFI1TnRWxMZNFuojJ4WifJWe`

---

## 🧪 Testing & Inspection API

These endpoints allow automated tests to pre-seed tokens, inspect verification history, dynamically change runtime behavior, and reset server state.

### `POST /testing/tokens`

Pre-seeds or mints a token with custom properties for backend verification tests.

**Content-Type:** `application/json`

**Request Body (all fields optional):**

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `token` | string | Auto-generated (`tok_<id>`) | Custom token string |
| `sitekey` | string | `"default-sitekey"` | Site key associated with the token |
| `action` | string | `"submit"` | Action returned on verification |
| `score` | number | Site/global `default_score` | Score (`0.0` – `1.0`) returned on verification |
| `hostname` | string | Site/global `hostname` | Hostname returned on verification |
| `ttl_seconds` | integer | `token_ttl_seconds` | Lifetime of the token in seconds |
| `expired` | boolean | `false` | If `true`, creates the token as already expired |
| `used` | boolean | `false` | If `true`, creates the token as already used/consumed |
| `error_codes` | array of strings | `[]` | If non-empty, verifying this token fails with these error codes |

**Example Request:**
```json
{
  "token": "custom-test-token-123",
  "sitekey": "site-key-1",
  "action": "checkout",
  "score": 0.3,
  "hostname": "shop.example.local"
}
```

**Response (`201 Created`):**
```json
{
  "token": "custom-test-token-123",
  "sitekey": "site-key-1",
  "action": "checkout",
  "score": 0.3,
  "hostname": "shop.example.local",
  "created_at": "2025-01-01T12:00:00Z",
  "expires_at": "2025-01-01T12:05:00Z",
  "used": false,
  "expired": false
}
```

**Errors:**
- `400 Bad Request` - If request body is invalid JSON

---

### `GET /testing/tokens`

Lists all issued and pre-seeded tokens currently stored in memory.

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `sitekey` | string | No | Filter tokens by `sitekey` |
| `used` | string | No | Filter by used state (`"true"` or `"false"`) |

**Response (`200 OK`):**
```json
{
  "tokens": [
    {
      "token": "tok_550e8400e29b41d4a716446655440000",
      "sitekey": "site-key-1",
      "action": "login",
      "score": 0.9,
      "hostname": "localhost",
      "created_at": "2025-01-01T12:00:00Z",
      "expires_at": "2025-01-01T12:05:00Z",
      "used": true,
      "expired": false
    }
  ]
}
```

---

### `GET /testing/verifications`

Returns a chronological list of all verification attempts made against `/recaptcha/api/siteverify`, `/siteverify`, or `/verify`.

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `response` | string | No | Filter verification attempts by the submitted `response` token |
| `success` | string | No | Filter by verification result (`"true"` or `"false"`) |

**Response (`200 OK`):**
```json
{
  "verifications": [
    {
      "secret": "secret-key-1",
      "response": "tok_550e8400e29b41d4a716446655440000",
      "remoteip": "127.0.0.1",
      "success": true,
      "score": 0.9,
      "action": "login",
      "hostname": "localhost",
      "error_codes": [],
      "timestamp": "2025-01-01T12:00:05Z"
    }
  ]
}
```

---

### `GET /testing/config`

Returns the current active runtime configuration.

**Response (`200 OK`):**
```json
{
  "port": 8080,
  "base_url": "http://localhost:8080",
  "hostname": "localhost",
  "allowed_origins": "*",
  "mode": "auto_pass",
  "default_score": 0.9,
  "token_ttl_seconds": 300,
  "verify_once": true,
  "allow_any_site": true,
  "allow_any_secret": true,
  "force_error_codes": [],
  "response_delay_ms": 0,
  "sites": [
    {
      "sitekey": "site-key-1",
      "secret": "secret-key-1",
      "mode": "auto_pass",
      "default_score": 0.9
    }
  ]
}
```

---

### `PUT /testing/config`
### `POST /testing/config`

Dynamically updates the active runtime configuration without restarting the container. Only the fields provided in the request JSON are modified. Both `PUT` and `POST` are supported.

**Content-Type:** `application/json`

**Request Body (all fields optional):**

| Field | Type | Default (at startup) | Description |
|-------|------|----------------------|-------------|
| `mode` | string | `"auto_pass"` | Global challenge/verification mode: `"auto_pass"`, `"challenge"`, or `"always_fail"` |
| `default_score` | number | `0.9` | Default v3 score (`0.0` – `1.0`) assigned to issued tokens |
| `token_ttl_seconds` | integer | `300` | Token lifetime in seconds |
| `verify_once` | boolean | `true` | Whether issued tokens are single-use (subsequent verifications fail with `"timeout-or-duplicate"`) |
| `allow_any_site` | boolean | `true` | Whether unknown `sitekey` values not listed in `sites` are accepted |
| `allow_any_secret` | boolean | `true` | Whether unknown `secret` values are accepted for non-site-bound tokens |
| `force_error_codes` | array of strings | `[]` | Force all verifications to fail with these error codes (set to `[]` to clear) |
| `response_delay_ms` | integer | `0` | Artificial response delay in milliseconds applied to solve and verify endpoints |
| `hostname` | string | `"localhost"` | Default `hostname` returned in verification responses |
| `sites` | array of objects | `[]` | Replace the active list of site configurations (`sitekey`, `secret`, `hostname`, `mode`, `default_score`, `error_codes`) |

**Example Request:**
```json
{
  "mode": "always_fail",
  "force_error_codes": ["invalid-input-response"]
}
```

**Response (`200 OK`):**

Returns the full updated runtime configuration object (same schema as `GET /testing/config`).

**Errors:**
- `400 Bad Request` - If request body is invalid JSON or `mode` is not one of `"auto_pass"`, `"challenge"`, `"always_fail"`

---

### `POST /testing/reset`

Clears all issued/pre-seeded tokens and recorded verification attempts from memory, and restores the runtime configuration to the initial values loaded from the YAML configuration file at startup.

**Response (`200 OK`):**
```json
{
  "status": "OK",
  "message": "State reset"
}
```

---

## 📋 Response Codes Summary

| Code | Description |
|------|-------------|
| 200 | Success - Request completed (note: `siteverify` returns `200 OK` even when `"success": false`) |
| 201 | Created - Token pre-seeded via `POST /testing/tokens` |
| 204 | No Content - CORS preflight `OPTIONS` response |
| 400 | Bad Request - Malformed JSON/form payload, invalid `mode`, or disallowed `sitekey` when `allow_any_site: false` |
| 404 | Not Found - Route does not exist |

---

## 📄 License

MIT © 2026 SIOCODE
