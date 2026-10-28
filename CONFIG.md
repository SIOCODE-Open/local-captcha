# Configuration Reference – SIOCODE Local Captcha

This document describes all configuration options for the SIOCODE Local Captcha service.

---

## Configuration File

The CAPTCHA provider is configured via a YAML file. By default, the server looks for `/config.yaml`, but you can specify a custom path using:

1. Command-line flag: `-c` or `--config-path`
2. Environment variable: `CONFIG_PATH`

### Configuration File Path

- **Command-Line Flag**: `-c <path>` or `--config-path <path>`
- **Environment Variable**: `CONFIG_PATH`
- **Default**: `/config.yaml`
- **Example**:
  - `local-captcha -c /app/my-config.yaml`
  - `local-captcha --config-path /app/my-config.yaml`
  - `CONFIG_PATH=/app/my-config.yaml local-captcha`

---

## Configuration Structure

### YAML Format

```yaml
port: 8080
base_url: http://localhost:8080
hostname: localhost
allowed_origins: "*"
mode: auto_pass
default_score: 0.9
token_ttl_seconds: 300
verify_once: true
allow_any_site: true
allow_any_secret: true
force_error_codes: []
response_delay_ms: 0
sites:
  - sitekey: "site-key-1"
    secret: "secret-key-1"
    hostname: "localhost"
    mode: "auto_pass"
    default_score: 0.9
  - sitekey: "site-key-fail"
    secret: "secret-key-fail"
    mode: "always_fail"
    error_codes:
      - "invalid-input-response"
magic_tokens:
  - token: "custom-pass-token"
    success: true
    score: 0.95
    action: "checkout"
    hostname: "shop.example.com"
```

---

## Configuration Options

### `port` (integer, optional)

The port on which the server will listen.

- **Type**: Integer
- **Default**: `8080`
- **Environment Variable Override**: `PORT`
- **Example**: `port: 8080`

If the `PORT` environment variable is set, it will override the value in the configuration file.

---

### `base_url` (string, optional)

The base URL used by the injectable JavaScript SDK (`/recaptcha/api.js` and `/captcha.js`) to construct URLs for the challenge iframe (`/recaptcha/anchor`) and the challenge solve endpoint (`/challenge/solve`).

- **Type**: String
- **Default**: `http://localhost:<port>`
- **Example**: `base_url: http://localhost:8080`

If not provided, defaults to `http://localhost:<port>` using the configured port. Set this when the browser accesses `local-captcha` through a custom hostname, reverse proxy, or mapped Docker port.

---

### `hostname` (string, optional)

The default hostname returned in the `hostname` field of `/recaptcha/api/siteverify` and `/challenge/solve` responses.

- **Type**: String
- **Default**: `"localhost"`
- **Example**: `hostname: "app.example.local"`

Can be overridden per site in `sites` or per request when solving/minting tokens.

---

### `allowed_origins` (string, optional)

The allowed origins for Cross-Origin Resource Sharing (CORS) requests.

- **Type**: String
- **Default**: `*` (all origins allowed)
- **Example**: `allowed_origins: http://localhost:3000`

This value configures CORS headers for cross-origin requests. The server will:
- Set the `Access-Control-Allow-Origin` header to the configured value
- Handle preflight `OPTIONS` requests automatically
- Allow credentials for cross-origin requests

For multiple origins, separate them with commas or use a wildcard `*` to allow all origins.

Common configurations:
- `*` - Allow all origins (default, convenient for local testing)
- `http://localhost:3000` - Allow single specific origin
- `http://localhost:3000,https://example.com` - Allow multiple specific origins

---

### `mode` (string, optional)

The global operating mode for CAPTCHA challenges and verification.

- **Type**: String (`"auto_pass"`, `"challenge"`, or `"always_fail"`)
- **Default**: `"auto_pass"`
- **Example**: `mode: "auto_pass"`

**Supported Modes:**
- `"auto_pass"` — Clicking the v2 checkbox or calling `grecaptcha.execute()` immediately issues a valid token without requiring extra user interaction.
- `"challenge"` — Clicking the v2 checkbox reveals an interactive challenge prompt where the user must click solve (or submit an answer) before a token is issued.
- `"always_fail"` — All challenge solve attempts and token verification requests fail with error codes (defaults to `["invalid-input-response"]`).

If an unrecognized value is provided, the server defaults to `"auto_pass"`. Individual sites in `sites` can override this mode.

---

### `default_score` (number, optional)

The default reCAPTCHA v3 score (`0.0` to `1.0`) assigned to newly issued tokens and returned by `/recaptcha/api/siteverify`.

- **Type**: Float
- **Default**: `0.9`
- **Example**: `default_score: 0.9`

Individual sites in `sites` can override `default_score` (e.g., to simulate bot traffic with `0.1`).

---

### `token_ttl_seconds` (integer, optional)

The lifetime in seconds for issued CAPTCHA tokens before they expire.

- **Type**: Integer
- **Default**: `300` (5 minutes)
- **Example**: `token_ttl_seconds: 300`

When a token is verified after its TTL has elapsed, `/recaptcha/api/siteverify` returns `"success": false` with `"error-codes": ["timeout-or-duplicate"]`.

---

### `verify_once` (boolean, optional)

Whether issued CAPTCHA tokens can only be verified once.

- **Type**: Boolean
- **Default**: `true`
- **Example**: `verify_once: true`

- When `true` (default, matching real Google reCAPTCHA behavior), verifying an already-used token returns `"success": false` with `"error-codes": ["timeout-or-duplicate"]`.
- When `false`, the same issued token can be verified multiple times until it expires.
- Note: Built-in and configured `magic_tokens` are reusable and are never consumed by `verify_once`.

---

### `allow_any_site` (boolean, optional)

Whether unknown `sitekey` values not explicitly listed in `sites` are accepted.

- **Type**: Boolean
- **Default**: `true`
- **Example**: `allow_any_site: true`

- When `true` (default), any `sitekey` can be used with `/recaptcha/anchor` and `/challenge/solve`, using the global `mode`, `default_score`, and `hostname`.
- When `false`, only `sitekey` values configured in `sites` (plus Google's official test sitekey) are accepted; unknown site keys return `400 Bad Request` with `"error-codes": ["invalid-sitekey"]`.

---

### `allow_any_secret` (boolean, optional)

Whether unknown `secret` values not explicitly listed in `sites` are accepted during `/recaptcha/api/siteverify`.

- **Type**: Boolean
- **Default**: `true`
- **Example**: `allow_any_secret: true`

- When `true` (default), any non-empty `secret` is accepted when verifying tokens whose `sitekey` is not bound to a specific site secret in `sites`. (If a token was issued for a site configured in `sites` that has a `secret`, that site's `secret` or Google's test secret is still required.)
- When `false`, `secret` must match a site configured in `sites` (or Google's official test secret); otherwise verification fails with `"error-codes": ["invalid-input-secret"]`.

---

### `force_error_codes` (array of strings, optional)

Forces all `/challenge/solve` and `/recaptcha/api/siteverify` calls to fail with the specified reCAPTCHA error codes.

- **Type**: Array of Strings
- **Default**: `[]` (no forced errors)
- **Example**:
  ```yaml
  force_error_codes:
    - "invalid-input-response"
  ```

Common reCAPTCHA error codes:
- `missing-input-secret` — The secret parameter is missing
- `invalid-input-secret` — The secret parameter is invalid or malformed
- `missing-input-response` — The response parameter is missing
- `invalid-input-response` — The response parameter is invalid or malformed
- `bad-request` — The request is invalid or malformed
- `timeout-or-duplicate` — The response is no longer valid: either is too old or has been used previously

---

### `response_delay_ms` (integer, optional)

Artificial latency in milliseconds added before responding to `/challenge/solve` and `/recaptcha/api/siteverify` requests.

- **Type**: Integer
- **Default**: `0` (no delay)
- **Example**: `response_delay_ms: 250`

Useful for testing frontend loading states and backend verification timeouts.

---

### `sites` (array, optional)

An array of site-specific configurations that associate a `sitekey` and `secret` with custom behavior, scores, or failure modes.

- **Type**: Array of SiteConfig objects
- **Default**: `[]` (empty array)

#### Site Object Properties

Each site in the array supports the following properties:

##### `sitekey` (string, required)

The public site key used by the frontend widget or `grecaptcha.execute()`.

- **Type**: String
- **Required**: Yes
- **Example**: `sitekey: "site-key-1"`

##### `secret` (string, required)

The shared secret key used by the backend when calling `/recaptcha/api/siteverify`.

- **Type**: String
- **Required**: Yes
- **Example**: `secret: "secret-key-1"`

##### `hostname` (string, optional)

Overrides the global `hostname` for tokens issued under this site.

- **Type**: String
- **Default**: Global `hostname`
- **Example**: `hostname: "portal.example.com"`

##### `mode` (string, optional)

Overrides the global `mode` (`"auto_pass"`, `"challenge"`, or `"always_fail"`) for this site.

- **Type**: String
- **Default**: Global `mode`
- **Example**: `mode: "challenge"`

##### `default_score` (number, optional)

Overrides the global `default_score` (`0.0` – `1.0`) for tokens issued under this site.

- **Type**: Float
- **Default**: Global `default_score`
- **Example**: `default_score: 0.1`

##### `error_codes` (array of strings, optional)

Custom error codes returned when this site's `mode` is `"always_fail"`.

- **Type**: Array of Strings
- **Default**: `["invalid-input-response"]` (when in `always_fail` mode)
- **Example**:
  ```yaml
  error_codes:
    - "invalid-input-response"
  ```

#### Sites Example

```yaml
sites:
  # Standard auto-passing site with high v3 score
  - sitekey: "site-key-1"
    secret: "secret-key-1"
    hostname: "localhost"
    mode: "auto_pass"
    default_score: 0.9

  # Interactive challenge site
  - sitekey: "site-key-challenge"
    secret: "secret-key-challenge"
    mode: "challenge"
    default_score: 0.8

  # Low-score site for simulating suspected bot traffic
  - sitekey: "site-key-bot"
    secret: "secret-key-bot"
    mode: "auto_pass"
    default_score: 0.1

  # Always-failing site for testing error handling
  - sitekey: "site-key-fail"
    secret: "secret-key-fail"
    mode: "always_fail"
    error_codes:
      - "invalid-input-response"
```

---

### `magic_tokens` (array, optional)

Custom static response tokens that always produce a deterministic verification result on `/recaptcha/api/siteverify` without needing to be issued or pre-seeded first. Magic tokens do not expire and are not consumed by `verify_once`.

- **Type**: Array of MagicTokenConfig objects
- **Default**: Empty/not set (built-in magic tokens are always active)

#### MagicToken Object Properties

##### `token` (string, required)

The response token string to match in `/recaptcha/api/siteverify`.

- **Type**: String
- **Required**: Yes
- **Example**: `token: "my-custom-pass-token"`

##### `success` (boolean, required)

Whether verification of this token succeeds (`true`) or fails (`false`).

- **Type**: Boolean
- **Required**: Yes
- **Example**: `success: true`

##### `score` (number, optional)

The v3 score (`0.0` – `1.0`) returned when `success` is `true`.

- **Type**: Float
- **Default**: Site/global `default_score`
- **Example**: `score: 0.95`

##### `action` (string, optional)

The action name returned when `success` is `true`.

- **Type**: String
- **Default**: `"submit"`
- **Example**: `action: "login"`

##### `hostname` (string, optional)

The hostname returned when `success` is `true`.

- **Type**: String
- **Default**: Site/global `hostname`
- **Example**: `hostname: "app.example.com"`

##### `error_codes` (array of strings, optional)

The error codes returned in `error-codes` when `success` is `false`.

- **Type**: Array of Strings
- **Default**: `["invalid-input-response"]` (when `success` is `false`)
- **Example**:
  ```yaml
  error_codes:
    - "timeout-or-duplicate"
  ```

#### MagicTokens Example

```yaml
magic_tokens:
  - token: "custom-vip-token"
    success: true
    score: 0.99
    action: "checkout"
    hostname: "shop.example.com"
  - token: "custom-expired-token"
    success: false
    error_codes:
      - "timeout-or-duplicate"
```

---

## Built-in Magic Tokens & Google Test Keys

### Built-in Magic Tokens

In addition to any custom `magic_tokens` defined in your configuration file, the following built-in magic tokens are always available for immediate backend testing:

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

### Official Google reCAPTCHA Test Keys

Google's official test credentials are always accepted by `local-captcha`, even when `allow_any_site: false` or `allow_any_secret: false`:

- **Test Sitekey**: `6LeIxAcTAAAAAJcZVRqyHh71UMIEGNQ_MXjiZKhI`
- **Test Secret**: `6LeIxAcTAAAAAGG-vFI1TnRWxMZNFuojJ4WifJWe`

---

## Environment Variables

### `CONFIG_PATH`

Path to the YAML configuration file.

- **Default**: `/config.yaml`
- **Example**: `CONFIG_PATH=/app/config/captcha.yaml`

### `PORT`

Server port (overrides the `port` setting in the configuration file).

- **Default**: `8080` (if not specified in config file)
- **Example**: `PORT=9000`

---

## Complete Configuration Example

```yaml
# Server configuration
port: 8080
base_url: http://localhost:8080
hostname: localhost
allowed_origins: "*"

# Default CAPTCHA behavior
mode: auto_pass             # "auto_pass", "challenge", or "always_fail"
default_score: 0.9          # Default reCAPTCHA v3 score (0.0 - 1.0)
token_ttl_seconds: 300      # Token lifetime in seconds (5 minutes)
verify_once: true           # Single-use tokens
allow_any_site: true        # Accept any sitekey not listed in sites
allow_any_secret: true      # Accept any secret not listed in sites
force_error_codes: []       # Optional forced error codes for all requests
response_delay_ms: 0        # Artificial delay in milliseconds

# Site-specific configurations
sites:
  - sitekey: "site-key-1"
    secret: "secret-key-1"
    hostname: "localhost"
    mode: "auto_pass"
    default_score: 0.9

  - sitekey: "site-key-challenge"
    secret: "secret-key-challenge"
    hostname: "localhost"
    mode: "challenge"
    default_score: 0.8

  - sitekey: "site-key-bot"
    secret: "secret-key-bot"
    hostname: "localhost"
    mode: "auto_pass"
    default_score: 0.1

  - sitekey: "site-key-fail"
    secret: "secret-key-fail"
    mode: "always_fail"
    error_codes:
      - "invalid-input-response"

# Custom magic tokens for deterministic backend testing
magic_tokens:
  - token: "custom-pass-token"
    success: true
    score: 0.95
    action: "login"
    hostname: "localhost"

  - token: "custom-fail-token"
    success: false
    error_codes:
      - "invalid-input-response"
```

---

## Docker Configuration

When running in Docker, mount your configuration file to `/config.yaml` or set the `CONFIG_PATH` environment variable:

```yaml
services:
  local-captcha:
    image: siocode/local-captcha
    ports:
      - "8080:8080"
    volumes:
      - ./my-config.yaml:/config.yaml:ro
    environment:
      - PORT=8080
```

Or with a custom path:

```yaml
services:
  local-captcha:
    image: siocode/local-captcha
    ports:
      - "9000:9000"
    volumes:
      - ./configs/captcha.yaml:/app/captcha-config.yaml:ro
    environment:
      - CONFIG_PATH=/app/captcha-config.yaml
      - PORT=9000
```

---

## Notes

- **All data is in-memory**: Issued tokens, pre-seeded tokens, verification logs, and runtime config changes are stored in memory and will be reset when the server restarts.
- **Runtime reconfiguration**: Active settings can be inspected and modified on the fly via `GET /testing/config`, `PUT /testing/config`, and reset to initial startup values via `POST /testing/reset`.
- **No persistence**: This service is designed for local development and automated testing, not production use.

---

## 📄 License

MIT © 2026 SIOCODE
