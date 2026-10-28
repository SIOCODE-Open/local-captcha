package main

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strings"
)

const anchorPageTemplate = `<!DOCTYPE html>
<html lang="en" class="{{if eq .Theme "dark"}}dark{{end}}">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>SIOCODE Local Captcha</title>
  <style>
    :root {
      --sioblue: #002266;
      --sioblue-hover: #001a4d;
      --sioblue-tint: #f0f4ff;
      --bg-card: #ffffff;
      --border-color: #e2e8f0;
      --text-fg: #020817;
      --text-muted: #64748b;
      --accent-color: #002266;
      --checkbox-bg: #ffffff;
      --checkbox-border: #002266;
      --checkbox-hover-bg: #f0f4ff;
      --success-color: #16a34a;
      --error-color: #dc2626;
      --error-bg: #fef2f2;
      --error-border: #fecaca;
      --input-bg: #ffffff;
    }

    html.dark {
      --bg-card: #020817;
      --border-color: #1e293b;
      --text-fg: #f8fafc;
      --text-muted: #94a3b8;
      --accent-color: #3b82f6;
      --checkbox-bg: #0f172a;
      --checkbox-border: #3b82f6;
      --checkbox-hover-bg: #1e293b;
      --success-color: #16a34a;
      --error-color: #f87171;
      --error-bg: #450a0a;
      --error-border: #7f1d1d;
      --input-bg: #0f172a;
    }

    * {
      box-sizing: border-box;
      margin: 0;
      padding: 0;
    }

    body {
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
      background: transparent;
      color: var(--text-fg);
      user-select: none;
      -webkit-font-smoothing: antialiased;
    }

    .captcha-card {
      background-color: var(--bg-card);
      border: 1px solid var(--border-color);
      border-radius: 8px;
      padding: 12px 14px;
      width: 100%;
      max-width: 302px;
      box-shadow: 0 1px 3px rgba(0, 34, 102, 0.08);
      display: flex;
      flex-direction: column;
      gap: 10px;
      transition: border-color 0.15s ease, box-shadow 0.15s ease;
    }

    .captcha-card.compact {
      max-width: 162px;
      padding: 10px;
    }

    .captcha-main-row {
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 10px;
    }

    .compact .captcha-main-row {
      flex-direction: column;
      align-items: flex-start;
      gap: 8px;
    }

    .anchor-left {
      display: flex;
      align-items: center;
      gap: 12px;
    }

    #recaptcha-anchor {
      width: 28px;
      height: 28px;
      border-radius: 6px;
      border: 2px solid var(--checkbox-border);
      background-color: var(--checkbox-bg);
      cursor: pointer;
      display: inline-flex;
      align-items: center;
      justify-content: center;
      padding: 0;
      outline: none;
      transition: all 0.15s ease;
      position: relative;
      flex-shrink: 0;
    }

    #recaptcha-anchor:hover {
      background-color: var(--checkbox-hover-bg);
      box-shadow: 0 0 0 3px rgba(0, 34, 102, 0.12);
    }

    #recaptcha-anchor:focus-visible {
      box-shadow: 0 0 0 3px rgba(0, 34, 102, 0.28);
    }

    #recaptcha-anchor.loading {
      border-color: var(--accent-color);
      pointer-events: none;
    }

    #recaptcha-anchor.checked {
      border-color: var(--success-color);
      background-color: var(--success-color);
    }

    #recaptcha-anchor.error {
      border-color: var(--error-color);
    }

    .spinner {
      display: none;
      width: 16px;
      height: 16px;
      border: 2px solid rgba(0, 34, 102, 0.2);
      border-top-color: var(--accent-color);
      border-radius: 50%;
      animation: spin 0.65s linear infinite;
    }

    #recaptcha-anchor.loading .spinner {
      display: block;
    }

    .checkmark {
      display: none;
      width: 16px;
      height: 16px;
      stroke: #ffffff;
      stroke-width: 2.5;
      fill: none;
      stroke-linecap: round;
      stroke-linejoin: round;
    }

    #recaptcha-anchor.checked .checkmark {
      display: block;
    }

    @keyframes spin {
      to { transform: rotate(360deg); }
    }

    .anchor-label {
      font-size: 14px;
      font-weight: 500;
      color: var(--text-fg);
      cursor: pointer;
      line-height: 1.2;
    }

    .brand-badge {
      display: flex;
      flex-direction: column;
      align-items: flex-end;
      text-align: right;
      flex-shrink: 0;
    }

    .compact .brand-badge {
      align-items: flex-start;
      text-align: left;
      width: 100%;
      border-top: 1px solid var(--border-color);
      padding-top: 6px;
    }

    .brand-logo-row {
      display: flex;
      align-items: center;
      gap: 5px;
    }

    .brand-icon {
      width: 18px;
      height: 18px;
      border-radius: 4px;
      background-color: var(--sioblue);
      display: inline-flex;
      align-items: center;
      justify-content: center;
      color: #ffffff;
      font-weight: 800;
      font-size: 10px;
      letter-spacing: -0.5px;
    }

    .brand-title {
      font-size: 11px;
      font-weight: 700;
      letter-spacing: 0.02em;
      color: var(--accent-color);
    }

    .brand-subtitle {
      font-size: 9px;
      color: var(--text-muted);
      margin-top: 2px;
    }

    #recaptcha-challenge {
      border-top: 1px solid var(--border-color);
      padding-top: 10px;
      display: flex;
      flex-direction: column;
      gap: 8px;
    }

    #recaptcha-challenge.hidden {
      display: none;
    }

    .challenge-header {
      font-size: 11px;
      font-weight: 600;
      color: var(--text-muted);
      display: flex;
      align-items: center;
      justify-content: space-between;
    }

    .challenge-controls {
      display: flex;
      gap: 6px;
    }

    #challenge-answer {
      flex: 1;
      min-width: 0;
      height: 30px;
      padding: 0 8px;
      font-size: 12px;
      color: var(--text-fg);
      background-color: var(--input-bg);
      border: 1px solid var(--border-color);
      border-radius: 5px;
      outline: none;
    }

    #challenge-answer:focus {
      border-color: var(--sioblue);
      box-shadow: 0 0 0 2px rgba(0, 34, 102, 0.15);
    }

    #solve-challenge-btn {
      height: 30px;
      padding: 0 12px;
      font-size: 12px;
      font-weight: 600;
      color: #ffffff;
      background-color: var(--sioblue);
      border: none;
      border-radius: 5px;
      cursor: pointer;
      transition: background-color 0.15s ease;
      white-space: nowrap;
    }

    #solve-challenge-btn:hover {
      background-color: var(--sioblue-hover);
    }

    #recaptcha-token {
      font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
      font-size: 10px;
      color: var(--success-color);
      word-break: break-all;
      display: none;
    }

    #recaptcha-token.visible {
      display: block;
    }

    #recaptcha-error {
      font-size: 11px;
      color: var(--error-color);
      background-color: var(--error-bg);
      border: 1px solid var(--error-border);
      border-radius: 5px;
      padding: 6px 8px;
      display: none;
    }

    #recaptcha-error.visible {
      display: block;
    }
  </style>
</head>
<body>
  <div class="captcha-card {{if eq .Size "compact"}}compact{{end}}" data-mode="{{.EffectiveMode}}" data-sitekey="{{.SiteKey}}" data-size="{{.Size}}" data-theme="{{.Theme}}">
    <div class="captcha-main-row">
      <div class="anchor-left">
        <button type="button" id="recaptcha-anchor" role="checkbox" aria-checked="false" aria-labelledby="recaptcha-anchor-label" class="{{if eq .EffectiveMode "always_fail"}}error{{end}}">
          <span class="spinner"></span>
          <svg class="checkmark" viewBox="0 0 24 24">
            <polyline points="20 6 9 17 4 12"></polyline>
          </svg>
        </button>
        <label id="recaptcha-anchor-label" for="recaptcha-anchor" class="anchor-label">I'm not a robot</label>
      </div>

      <div class="brand-badge">
        <div class="brand-logo-row">
          <span class="brand-icon">S</span>
          <span class="brand-title">SIOCODE Captcha</span>
        </div>
        <span class="brand-subtitle">Protected by SIOCODE Local Captcha</span>
      </div>
    </div>

    <div id="recaptcha-challenge" class="{{if ne .EffectiveMode "challenge"}}hidden{{end}}">
      <div class="challenge-header">
        <span>Interactive Verification Challenge</span>
      </div>
      <div class="challenge-controls">
        <input type="text" id="challenge-answer" name="challenge_answer" placeholder="Enter challenge answer..." value="human" autocomplete="off" />
        <button type="button" id="solve-challenge-btn">Verify</button>
      </div>
    </div>

    <div id="recaptcha-token" data-token=""></div>
    <div id="recaptcha-error" class="{{if eq .EffectiveMode "always_fail"}}visible{{end}}">{{if eq .EffectiveMode "always_fail"}}{{.InitialErrorText}}{{end}}</div>
  </div>

  <script>
    (function() {
      var SITEKEY = {{.SiteKeyJSON}};
      var ACTION = {{.ActionJSON}};
      var WIDGET_ID = {{.WidgetIdJSON}};
      var MODE = {{.EffectiveModeJSON}};

      var anchorBtn = document.getElementById("recaptcha-anchor");
      var challengeBox = document.getElementById("recaptcha-challenge");
      var answerInput = document.getElementById("challenge-answer");
      var solveBtn = document.getElementById("solve-challenge-btn");
      var tokenEl = document.getElementById("recaptcha-token");
      var errorEl = document.getElementById("recaptcha-error");

      function notifyResize() {
        try {
          if (window.parent && window.parent !== window) {
            var h = document.body.scrollHeight;
            if (h > 78) {
              window.parent.postMessage({
                type: "local-captcha-resize",
                widgetId: WIDGET_ID,
                height: h + 8
              }, "*");
            }
          }
        } catch (e) {}
      }

      function showError(codes) {
        var errList = Array.isArray(codes) && codes.length > 0 ? codes : ["invalid-input-response"];
        anchorBtn.classList.remove("loading", "checked");
        anchorBtn.classList.add("error");
        anchorBtn.setAttribute("aria-checked", "false");
        errorEl.textContent = "Verification failed: " + errList.join(", ");
        errorEl.classList.add("visible");
        tokenEl.textContent = "";
        tokenEl.setAttribute("data-token", "");
        tokenEl.classList.remove("visible");

        try {
          if (window.parent && window.parent !== window) {
            window.parent.postMessage({
              type: "local-captcha-error",
              widgetId: WIDGET_ID,
              errorCodes: errList
            }, "*");
          }
        } catch (e) {}
        notifyResize();
      }

      function showSolved(token) {
        anchorBtn.classList.remove("loading", "error");
        anchorBtn.classList.add("checked");
        anchorBtn.setAttribute("aria-checked", "true");
        errorEl.textContent = "";
        errorEl.classList.remove("visible");
        challengeBox.classList.add("hidden");

        tokenEl.textContent = token;
        tokenEl.setAttribute("data-token", token);
        tokenEl.classList.add("visible");

        try {
          if (window.parent && window.parent !== window) {
            window.parent.postMessage({
              type: "local-captcha-solved",
              widgetId: WIDGET_ID,
              token: token
            }, "*");
          }
        } catch (e) {}
      }

      function requestSolve(challengeAnswer) {
        anchorBtn.classList.add("loading");
        anchorBtn.classList.remove("error");

        var payload = {
          sitekey: SITEKEY,
          action: ACTION
        };
        if (challengeAnswer !== undefined) {
          payload.challenge_answer = challengeAnswer;
        }

        fetch("/challenge/solve", {
          method: "POST",
          headers: {
            "Content-Type": "application/json"
          },
          body: JSON.stringify(payload)
        })
        .then(function(res) {
          return res.json();
        })
        .then(function(data) {
          if (!data || !data.success) {
            showError(data && data["error-codes"] ? data["error-codes"] : ["invalid-input-response"]);
            return;
          }
          showSolved(data.token);
        })
        .catch(function() {
          showError(["invalid-input-response"]);
        });
      }

      anchorBtn.addEventListener("click", function() {
        if (anchorBtn.classList.contains("checked") || anchorBtn.classList.contains("loading")) {
          return;
        }
        if (MODE === "challenge") {
          challengeBox.classList.remove("hidden");
          notifyResize();
          if (answerInput) {
            answerInput.focus();
          }
          return;
        }
        requestSolve(undefined);
      });

      solveBtn.addEventListener("click", function() {
        var ans = answerInput ? answerInput.value : "";
        requestSolve(ans);
      });

      if (answerInput) {
        answerInput.addEventListener("keydown", function(e) {
          if (e.key === "Enter") {
            e.preventDefault();
            solveBtn.click();
          }
        });
      }

      if (MODE === "challenge" || MODE === "always_fail") {
        setTimeout(notifyResize, 10);
      }
    })();
  </script>
</body>
</html>
`

type anchorTemplateData struct {
	SiteKey           string
	Size              string
	Theme             string
	Action            string
	WidgetId          string
	EffectiveMode     string
	InitialErrorText  string
	SiteKeyJSON       template.JS
	ActionJSON        template.JS
	WidgetIdJSON      template.JS
	EffectiveModeJSON template.JS
}

func GET_recaptcha_anchor(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	siteKey := strings.TrimSpace(q.Get("k"))
	if siteKey == "" {
		siteKey = strings.TrimSpace(q.Get("sitekey"))
	}

	AppContext.mu.RLock()
	cfg := CloneConfig(AppContext.Config)
	AppContext.mu.RUnlock()

	if siteKey == "" {
		if !cfg.AllowAnySite {
			writeJSON(w, http.StatusBadRequest, ChallengeSolveResponse{
				Success:    false,
				Error:      "invalid-sitekey",
				ErrorCodes: []string{"invalid-sitekey"},
			})
			return
		}
		siteKey = "default-sitekey"
	}

	if !isSiteKeyAllowed(cfg, siteKey) {
		writeJSON(w, http.StatusBadRequest, ChallengeSolveResponse{
			Success:    false,
			Error:      "invalid-sitekey",
			ErrorCodes: []string{"invalid-sitekey"},
		})
		return
	}

	size := q.Get("size")
	if size != "compact" && size != "invisible" {
		size = "normal"
	}

	theme := q.Get("theme")
	if theme != "dark" {
		theme = "light"
	}

	action := q.Get("action")
	if action == "" {
		action = "submit"
	}

	widgetId := q.Get("widget_id")
	if widgetId == "" {
		widgetId = "0"
	}

	site := findSiteBySiteKey(cfg, siteKey)
	effectiveMode := resolveEffectiveMode(cfg, site)

	initialErrorText := ""
	if effectiveMode == ModeAlwaysFail {
		errCodes := []string{"invalid-input-response"}
		if len(cfg.ForceErrorCodes) > 0 {
			errCodes = cfg.ForceErrorCodes
		} else if site != nil && len(site.ErrorCodes) > 0 {
			errCodes = site.ErrorCodes
		}
		initialErrorText = "Verification failed: " + strings.Join(errCodes, ", ")
	}

	siteKeyJSON, _ := json.Marshal(siteKey)
	actionJSON, _ := json.Marshal(action)
	widgetIdJSON, _ := json.Marshal(widgetId)
	modeJSON, _ := json.Marshal(effectiveMode)

	tmpl, err := template.New("anchor").Parse(anchorPageTemplate)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := anchorTemplateData{
		SiteKey:           siteKey,
		Size:              size,
		Theme:             theme,
		Action:            action,
		WidgetId:          widgetId,
		EffectiveMode:     effectiveMode,
		InitialErrorText:  initialErrorText,
		SiteKeyJSON:       template.JS(siteKeyJSON),
		ActionJSON:        template.JS(actionJSON),
		WidgetIdJSON:      template.JS(widgetIdJSON),
		EffectiveModeJSON: template.JS(modeJSON),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = tmpl.Execute(w, data)
}
