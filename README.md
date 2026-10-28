# Local Captcha

A fully in-memory, configurable **local CAPTCHA provider** that mimics Google reCAPTCHA (v2 & v3) for local development and automated testing.

Supports:

- Drop-in injectable script (`GET /recaptcha/api.js` and `GET /captcha.js`) providing `window.grecaptcha`
- Google reCAPTCHA v2 Checkbox, v2 Invisible, and v3 (`grecaptcha.execute`) flows
- Interactive challenge iframe UI (`GET /recaptcha/anchor`, `GET /challenge`) & programmatic solve API (`POST /challenge/solve`)
- Standard reCAPTCHA verification backend API (`POST /recaptcha/api/siteverify`, `POST /siteverify`, `POST /verify`)
- Configurable auto-pass, interactive challenge, forced error modes, and custom v3 scores
- Testing & inspection API (`/testing/...`) for pre-seeding tokens, inspecting verifications, and mutating runtime behavior
- Dockerized and architecture-portable (x86_64 and arm64)

## 🚀 Purpose

This project simulates a **realistic CAPTCHA service** (compatible with Google reCAPTCHA v2 and v3) that you can run locally without any external dependencies or internet access. It's perfect for:

- Developing frontend forms and SPAs that use `grecaptcha` or `.g-recaptcha` widgets
- Testing backend verification logic (`siteverify`) including edge cases, low v3 scores, expired/duplicate tokens, and error codes
- Running end-to-end tests in CI/CD or air-gapped environments **without needing cloud keys or bypassing captcha code**

All state is **in-memory** and defined via a simple YAML config file.

## 🐳 Docker

Pre-built images available at:

```
docker pull siocode/local-captcha
```

Available for:
- `linux/amd64` (x86 64-bit)
- `linux/arm64` (ARM 64-bit)

## 🛠️ Configuration

YAML configuration is passed via `CONFIG_PATH` env var (or `-c` / `--config-path` flag) or defaults to `/config.yaml`.

See [`CONFIG.md`](./CONFIG.md) for full details.

### Example: `example/local-captcha.config.yaml`

```yaml
port: 8080
base_url: http://localhost:8080
hostname: localhost
mode: auto_pass # "auto_pass", "challenge", or "always_fail"
default_score: 0.9
token_ttl_seconds: 300
verify_once: true
allow_any_site: true
allow_any_secret: true
sites:
  - sitekey: "site-key-1"
    secret: "secret-key-1"
    mode: "auto_pass"
    default_score: 0.9
  - sitekey: "site-key-challenge"
    secret: "secret-key-challenge"
    mode: "challenge"
    default_score: 0.8
  - sitekey: "site-key-bot"
    secret: "secret-key-bot"
    mode: "auto_pass"
    default_score: 0.1
  - sitekey: "site-key-fail"
    secret: "secret-key-fail"
    mode: "always_fail"
    error_codes:
      - "invalid-input-response"
```

## 🔐 Endpoints

[📄 See `API.md` for a complete reference of the API endpoints](./API.md)

### 🔧 Health & Metadata

| Method | Path       | Description  |
| ------ | ---------- | ------------ |
| GET    | `/healthz` | Health check |

### 📜 Injectable Script & Challenge UI

| Method | Path                 | Description                                                |
| ------ | -------------------- | ---------------------------------------------------------- |
| GET    | `/recaptcha/api.js`  | Injectable JavaScript SDK (`window.grecaptcha`)            |
| GET    | `/captcha.js`        | Alias for `/recaptcha/api.js`                              |
| GET    | `/recaptcha/anchor`  | Interactive CAPTCHA widget / challenge HTML page (iframe)  |
| GET    | `/challenge`         | Alias for `/recaptcha/anchor`                              |
| POST   | `/challenge/solve`   | Issue or solve a CAPTCHA challenge and obtain a token      |
| POST   | `/recaptcha/reload`  | Alias for `/challenge/solve`                               |

### ✅ Verification API

| Method | Path                        | Description                                           |
| ------ | --------------------------- | ----------------------------------------------------- |
| POST   | `/recaptcha/api/siteverify` | Verify a CAPTCHA response token (form or JSON)        |
| POST   | `/siteverify`               | Alias for `/recaptcha/api/siteverify`                 |
| POST   | `/verify`                   | Alias for `/recaptcha/api/siteverify`                 |

### 🧪 Testing & Inspection API

| Method | Path                     | Description                                                 |
| ------ | ------------------------ | ----------------------------------------------------------- |
| POST   | `/testing/tokens`        | Pre-seed or mint a custom token with specific attributes    |
| GET    | `/testing/tokens`        | List all issued tokens and their current state              |
| GET    | `/testing/verifications` | List recorded verification attempts                         |
| GET    | `/testing/config`        | Get active runtime configuration                            |
| PUT    | `/testing/config`        | Dynamically update runtime configuration (mode, score, etc) |
| POST   | `/testing/reset`         | Clear tokens/verifications and restore initial config       |

## 🧩 Frontend & Backend Integration

### Frontend (HTML / JS)

Replace `https://www.google.com/recaptcha/api.js` with your local-captcha instance URL:

```html
<script src="http://localhost:8080/recaptcha/api.js" async defer></script>
<form action="/submit" method="POST">
  <div class="g-recaptcha" data-sitekey="site-key-1" data-callback="onSolved"></div>
  <button type="submit">Submit</button>
</form>
```

Or use reCAPTCHA v3 / programmatic execution:

```html
<script src="http://localhost:8080/recaptcha/api.js?render=site-key-1"></script>
<script>
  grecaptcha.ready(function () {
    grecaptcha.execute('site-key-1', { action: 'login' }).then(function (token) {
      console.log('Captcha token:', token);
    });
  });
</script>
```

### Backend Verification

Point your backend's reCAPTCHA verification URL to `http://localhost:8080/recaptcha/api/siteverify`:

```bash
curl -X POST http://localhost:8080/recaptcha/api/siteverify \
  -d "secret=secret-key-1&response=<token>"
```

Response:

```json
{
  "success": true,
  "challenge_ts": "2025-01-01T12:00:00Z",
  "hostname": "localhost",
  "score": 0.9,
  "action": "login"
}
```

## 🧪 Example Docker Compose

### `example/docker-compose.yaml`

```yaml
services:
  local-captcha:
    image: siocode/local-captcha
    ports:
      - "8080:8080"
    volumes:
      - ./local-captcha.config.yaml:/config.yaml:ro
```

Start it:

```bash
cd example
docker compose up
```

## 🏗️ Building

Build the container via the following command (builds for both X86 & ARM architectures):

```bash
docker buildx build --platform=linux/arm64,linux/amd64 -t siocode/local-captcha:latest .
```

## 🧪 Testing

End-to-end tests are in the `e2e/` folder using Mocha + Chai + Docker Compose. To run:

```bash
cd e2e
npm install
npm test
```

## 📜 License

MIT © 2026 SIOCODE
