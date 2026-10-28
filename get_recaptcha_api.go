package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func GET_recaptcha_api_js(w http.ResponseWriter, r *http.Request) {
	AppContext.mu.RLock()
	configuredBaseUrl := AppContext.Config.BaseUrl
	AppContext.mu.RUnlock()

	baseUrl := strings.TrimRight(configuredBaseUrl, "/")
	if baseUrl == "" {
		scheme := "http"
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		host := r.Host
		if host == "" {
			host = "localhost:8080"
		}
		baseUrl = fmt.Sprintf("%s://%s", scheme, host)
	}

	renderParam := r.URL.Query().Get("render")
	if renderParam == "" {
		renderParam = "onload"
	}
	onloadParam := r.URL.Query().Get("onload")
	hlParam := r.URL.Query().Get("hl")

	baseUrlJSON, _ := json.Marshal(baseUrl)
	renderJSON, _ := json.Marshal(renderParam)
	onloadJSON, _ := json.Marshal(onloadParam)
	hlJSON, _ := json.Marshal(hlParam)

	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	script := fmt.Sprintf(`/* SIOCODE Local Captcha JS SDK */
(function(window, document) {
  "use strict";

  var CONFIGURED_BASE_URL = %s;
  var RENDER_MODE = %s;
  var ONLOAD_CALLBACK = %s;
  var HL_LANG = %s;

  var BASE_URL = CONFIGURED_BASE_URL;
  try {
    if (document && document.currentScript && document.currentScript.src) {
      var parsedUrl = new URL(document.currentScript.src, window.location ? window.location.href : undefined);
      if (parsedUrl.origin && parsedUrl.origin !== "null") {
        BASE_URL = parsedUrl.origin;
      }
    }
  } catch (e) {}

  var DEFAULT_SITEKEY = (RENDER_MODE !== "onload" && RENDER_MODE !== "explicit") ? RENDER_MODE : "default-sitekey";

  var widgets = {};
  var nextWidgetId = 0;
  var readyQueue = [];
  var isReady = false;

  function resolveCallback(cb) {
    if (typeof cb === "function") {
      return cb;
    }
    if (typeof cb === "string" && cb.length > 0 && typeof window[cb] === "function") {
      return window[cb];
    }
    return null;
  }

  function invokeCallback(cb, arg) {
    var fn = resolveCallback(cb);
    if (fn) {
      try {
        fn(arg);
      } catch (err) {
        console.error("grecaptcha callback error:", err);
      }
    }
  }

  function resolveContainer(container) {
    if (!container) {
      return null;
    }
    if (typeof container === "string") {
      return document.getElementById(container);
    }
    if (typeof container === "object" && container.nodeType === 1) {
      return container;
    }
    return null;
  }

  function buildAnchorUrl(params, widgetId) {
    var q = [];
    q.push("k=" + encodeURIComponent(params.sitekey));
    q.push("sitekey=" + encodeURIComponent(params.sitekey));
    q.push("size=" + encodeURIComponent(params.size));
    q.push("theme=" + encodeURIComponent(params.theme));
    q.push("action=" + encodeURIComponent(params.action));
    q.push("widget_id=" + encodeURIComponent(String(widgetId)));
    if (HL_LANG) {
      q.push("hl=" + encodeURIComponent(HL_LANG));
    }
    return BASE_URL + "/recaptcha/anchor?" + q.join("&");
  }

  function render(container, parameters) {
    var el = resolveContainer(container);
    if (!el) {
      throw new Error("reCAPTCHA placeholder element must be an element or id");
    }

    if (el.getAttribute("data-LocalCaptchaWidgetId") !== null) {
      return parseInt(el.getAttribute("data-LocalCaptchaWidgetId"), 10);
    }

    var opts = parameters || {};
    var sitekey = opts.sitekey || el.getAttribute("data-sitekey") || DEFAULT_SITEKEY || "default-sitekey";
    var callback = opts.callback || el.getAttribute("data-callback") || null;
    var expiredCallback = opts["expired-callback"] || opts.expiredCallback || el.getAttribute("data-expired-callback") || null;
    var errorCallback = opts["error-callback"] || opts.errorCallback || el.getAttribute("data-error-callback") || null;
    var size = opts.size || el.getAttribute("data-size") || "normal";
    var theme = opts.theme || el.getAttribute("data-theme") || "light";
    var action = opts.action || el.getAttribute("data-action") || "submit";

    var widgetId = nextWidgetId++;
    el.setAttribute("data-LocalCaptchaWidgetId", String(widgetId));

    el.innerHTML = "";

    var textarea = document.createElement("textarea");
    textarea.id = widgetId === 0 ? "g-recaptcha-response" : "g-recaptcha-response-" + widgetId;
    textarea.name = "g-recaptcha-response";
    textarea.className = "g-recaptcha-response";
    textarea.style.cssText = "width: 250px; height: 40px; border: 1px solid #e2e8f0; margin: 10px 25px; padding: 0px; resize: none; display: none;";

    var iframe = document.createElement("iframe");
    var iframeSrc = buildAnchorUrl({
      sitekey: sitekey,
      size: size,
      theme: theme,
      action: action
    }, widgetId);
    iframe.src = iframeSrc;
    iframe.title = "SIOCODE Local Captcha";
    iframe.setAttribute("frameborder", "0");
    iframe.setAttribute("scrolling", "no");

    if (size === "invisible") {
      iframe.width = "0";
      iframe.height = "0";
      iframe.style.cssText = "display: none; width: 0; height: 0; border: 0;";
    } else if (size === "compact") {
      iframe.width = "164";
      iframe.height = "144";
      iframe.style.cssText = "width: 164px; height: 144px; border: 0; overflow: hidden;";
    } else {
      iframe.width = "304";
      iframe.height = "78";
      iframe.style.cssText = "width: 304px; height: 78px; border: 0; overflow: hidden;";
    }

    el.appendChild(iframe);
    el.appendChild(textarea);

    widgets[widgetId] = {
      id: widgetId,
      container: el,
      textarea: textarea,
      iframe: iframe,
      iframeSrc: iframeSrc,
      sitekey: sitekey,
      callback: callback,
      expiredCallback: expiredCallback,
      errorCallback: errorCallback,
      size: size,
      theme: theme,
      action: action,
      token: ""
    };

    return widgetId;
  }

  function execute(sitekeyOrWidgetId, options) {
    var targetWidget = null;
    var sitekey = DEFAULT_SITEKEY || "default-sitekey";
    var action = "submit";

    if (typeof sitekeyOrWidgetId === "number") {
      targetWidget = widgets[sitekeyOrWidgetId] || null;
      if (targetWidget) {
        sitekey = targetWidget.sitekey;
        action = (options && options.action) || targetWidget.action || "submit";
      }
    } else if (typeof sitekeyOrWidgetId === "string") {
      if (widgets.hasOwnProperty(sitekeyOrWidgetId)) {
        targetWidget = widgets[Number(sitekeyOrWidgetId)];
        sitekey = targetWidget.sitekey;
        action = (options && options.action) || targetWidget.action || "submit";
      } else {
        sitekey = sitekeyOrWidgetId;
        action = (options && options.action) || "submit";
      }
    } else if (typeof sitekeyOrWidgetId === "object" && sitekeyOrWidgetId !== null) {
      options = sitekeyOrWidgetId;
      if (RENDER_MODE !== "onload" && RENDER_MODE !== "explicit") {
        sitekey = RENDER_MODE;
      } else if (widgets[0]) {
        targetWidget = widgets[0];
        sitekey = targetWidget.sitekey;
      }
      action = (options && options.action) || (targetWidget && targetWidget.action) || "submit";
    } else if (sitekeyOrWidgetId === undefined || sitekeyOrWidgetId === null) {
      if (widgets[0]) {
        targetWidget = widgets[0];
        sitekey = targetWidget.sitekey;
        action = (options && options.action) || targetWidget.action || "submit";
      } else if (RENDER_MODE !== "onload" && RENDER_MODE !== "explicit") {
        sitekey = RENDER_MODE;
        action = (options && options.action) || "submit";
      }
    }

    return fetch(BASE_URL + "/challenge/solve", {
      method: "POST",
      headers: {
        "Content-Type": "application/json"
      },
      body: JSON.stringify({
        sitekey: sitekey,
        action: action
      })
    })
    .then(function(res) {
      return res.json();
    })
    .then(function(data) {
      if (!data || !data.success) {
        var errCodes = (data && data["error-codes"]) ? data["error-codes"] : ["invalid-input-response"];
        if (targetWidget) {
          targetWidget.token = "";
          if (targetWidget.textarea) {
            targetWidget.textarea.value = "";
          }
          invokeCallback(targetWidget.errorCallback, errCodes);
        }
        throw new Error(errCodes.join(", "));
      }

      var token = data.token;
      if (targetWidget) {
        targetWidget.token = token;
        if (targetWidget.textarea) {
          targetWidget.textarea.value = token;
        }
        invokeCallback(targetWidget.callback, token);
      }
      return token;
    });
  }

  function getResponse(widgetId) {
    var id = (widgetId === undefined || widgetId === null) ? 0 : Number(widgetId);
    var w = widgets[id];
    if (!w) {
      return "";
    }
    return w.token || "";
  }

  function reset(widgetId) {
    var id = (widgetId === undefined || widgetId === null) ? 0 : Number(widgetId);
    var w = widgets[id];
    if (!w) {
      return;
    }
    w.token = "";
    if (w.textarea) {
      w.textarea.value = "";
    }
    if (w.iframe) {
      w.iframe.src = w.iframeSrc;
    }
  }

  function ready(callback) {
    if (typeof callback !== "function") {
      return;
    }
    if (isReady) {
      setTimeout(callback, 0);
    } else {
      readyQueue.push(callback);
    }
  }

  if (typeof window.addEventListener === "function") {
    window.addEventListener("message", function(event) {
      var data = event.data;
      if (!data || typeof data !== "object") {
        return;
      }
      if (data.type === "local-captcha-solved") {
        var id = (data.widgetId !== undefined && data.widgetId !== "") ? Number(data.widgetId) : 0;
        var w = widgets[id] || widgets[0];
        if (w) {
          w.token = data.token || "";
          if (w.textarea) {
            w.textarea.value = w.token;
          }
          invokeCallback(w.callback, w.token);
        }
      } else if (data.type === "local-captcha-error") {
        var errId = (data.widgetId !== undefined && data.widgetId !== "") ? Number(data.widgetId) : 0;
        var errW = widgets[errId] || widgets[0];
        if (errW) {
          errW.token = "";
          if (errW.textarea) {
            errW.textarea.value = "";
          }
          invokeCallback(errW.errorCallback, data.errorCodes || ["invalid-input-response"]);
        }
      } else if (data.type === "local-captcha-resize") {
        var rId = (data.widgetId !== undefined && data.widgetId !== "") ? Number(data.widgetId) : 0;
        var rW = widgets[rId] || widgets[0];
        if (rW && rW.iframe && data.height) {
          rW.iframe.style.height = data.height + "px";
          rW.iframe.height = String(data.height);
        }
      }
    });
  }

  var grecaptchaObj = {
    ready: ready,
    render: render,
    execute: execute,
    getResponse: getResponse,
    reset: reset,
    enterprise: {
      ready: ready,
      render: render,
      execute: execute,
      getResponse: getResponse,
      reset: reset
    }
  };

  window.grecaptcha = grecaptchaObj;

  function initDom() {
    if (RENDER_MODE === "onload") {
      var elements = document.querySelectorAll(".g-recaptcha");
      for (var i = 0; i < elements.length; i++) {
        try {
          render(elements[i]);
        } catch (e) {}
      }
    }

    isReady = true;
    while (readyQueue.length > 0) {
      var cb = readyQueue.shift();
      setTimeout(cb, 0);
    }

    if (ONLOAD_CALLBACK && typeof window[ONLOAD_CALLBACK] === "function") {
      try {
        window[ONLOAD_CALLBACK]();
      } catch (e) {
        console.error("grecaptcha onload error:", e);
      }
    }
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", initDom);
  } else {
    setTimeout(initDom, 0);
  }
})(window, document);
`, string(baseUrlJSON), string(renderJSON), string(onloadJSON), string(hlJSON))

	_, _ = w.Write([]byte(script))
}
