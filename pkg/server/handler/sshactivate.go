package handler

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/talesmud/talesmud/pkg/gamemode"
)

// Activate serves the local sign-in page. A classic Auth0 process redirects
// into the play client, which already holds the Auth0 session.
func Activate(local bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !local {
			code := strings.TrimSpace(c.Query("code"))
			target := "/play/"
			if code != "" {
				target = "/play/?activate=" + url.QueryEscape(code)
			}
			c.Redirect(http.StatusFound, target)
			return
		}
		setActivateHeaders(c)
		_, _, tokenKey := gamemode.ClientPage()
		c.Data(http.StatusOK, "text/html; charset=utf-8", activatePage(tokenKey))
	}
}

func setActivateHeaders(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	c.Header("X-Frame-Options", "DENY")
	c.Header("Content-Security-Policy", "frame-ancestors 'none'")
	c.Header("Referrer-Policy", "no-referrer")
	c.Header("X-Content-Type-Options", "nosniff")
}

func activatePage(tokenKey string) []byte {
	key, err := json.Marshal(tokenKey)
	if err != nil || tokenKey == "" {
		key = []byte(`"talesmudDoorToken"`)
	}
	return []byte(strings.ReplaceAll(activateHTML, "__TOKEN_KEY__", string(key)))
}

const activateHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>TalesMUD sign-in</title>
<style>
body { font-family: sans-serif; max-width: 40rem; margin: 2rem auto; padding: 0 1rem; }
label { display: block; margin-top: 0.8rem; }
input { width: 100%; box-sizing: border-box; }
button { margin-top: 0.8rem; margin-right: 0.4rem; }
#detail { white-space: pre-wrap; }
</style>
</head>
<body>
<h1>TalesMUD sign-in</h1>
<form id="login">
<label>Username <input id="username" autocomplete="username"></label>
<label>Password <input id="password" type="password" autocomplete="current-password"></label>
<p id="loginerr"></p>
<button type="submit">Sign in</button>
</form>
<div id="panel" hidden>
<label>Code <input id="code" autocomplete="off"></label>
<button type="button" onclick="lookupCode()">Look up</button>
<p id="detail"></p>
<div id="decide" hidden>
<button type="button" onclick="decide('/api/ssh/device/confirm')">Confirm</button>
<button type="button" onclick="decide('/api/ssh/device/deny')">Deny</button>
</div>
<h2>SSH keys</h2>
<div id="keys"></div>
<form id="addkey">
<label>Public key <input id="pubkey" autocomplete="off"></label>
<label>Label <input id="label" autocomplete="off"></label>
<button type="submit">Add key</button>
</form>
<p id="keyerr"></p>
</div>
<script>
const tokenKey = __TOKEN_KEY__;
const codeInput = document.getElementById("code");
const params = new URLSearchParams(location.search);
if (params.get("code")) codeInput.value = params.get("code");
function token() { return localStorage.getItem(tokenKey) || ""; }
function showLogin(on) {
  document.getElementById("login").hidden = !on;
  document.getElementById("panel").hidden = on;
}
async function api(method, path, body) {
  const headers = {"Authorization": "Bearer " + token()};
  let payload;
  if (body) {
    headers["Content-Type"] = "application/json";
    payload = JSON.stringify(body);
  }
  const res = await fetch(path, {method: method, headers: headers, body: payload});
  const data = await res.json().catch(function () { return {}; });
  if (!res.ok) throw new Error(data.error || "request failed");
  return data;
}
document.getElementById("login").addEventListener("submit", async function (ev) {
  ev.preventDefault();
  const res = await fetch("/api/auth/login", {
    method: "POST",
    headers: {"Content-Type": "application/json"},
    body: JSON.stringify({
      username: document.getElementById("username").value,
      password: document.getElementById("password").value
    })
  });
  const data = await res.json().catch(function () { return {}; });
  if (!res.ok) {
    document.getElementById("loginerr").textContent = data.error || "login failed";
    return;
  }
  localStorage.setItem(tokenKey, data.token || "");
  showLogin(false);
  loadKeys();
});
let csrf = "";
async function lookupCode() {
  document.getElementById("detail").textContent = "";
  try {
    const data = await api("POST", "/api/ssh/device/lookup", {user_code: codeInput.value});
    csrf = data.csrf || "";
    let ago = "";
    if (data.created) {
      const mins = Math.max(0, Math.round((Date.now() - new Date(data.created).getTime()) / 60000));
      ago = mins + " min ago";
    }
    document.getElementById("detail").textContent =
      "An SSH session from " + (data.ip || "unknown") + " (" + ago + ") wants to sign in as you. " +
      "Mode " + (data.mode || "") + ". Client " + (data.client_version || "") + ". " +
      "Key " + (data.key || "none") + ". Only confirm if you started it yourself.";
    document.getElementById("decide").hidden = false;
  } catch (err) {
    document.getElementById("detail").textContent = err.message;
    document.getElementById("decide").hidden = true;
  }
}
async function decide(path) {
  try {
    await api("POST", path, {user_code: codeInput.value, csrf: csrf});
    document.getElementById("detail").textContent = path.indexOf("deny") >= 0 ? "Denied." : "Confirmed. Return to your SSH window.";
    document.getElementById("decide").hidden = true;
  } catch (err) {
    document.getElementById("detail").textContent = err.message;
  }
}
async function loadKeys() {
  const box = document.getElementById("keys");
  box.textContent = "";
  try {
    const rows = await api("GET", "/api/ssh/keys");
    (rows || []).forEach(function (row) {
      const p = document.createElement("p");
      p.textContent = (row.label || row.key_type || "key") + " " + (row.fingerprint || "");
      const b = document.createElement("button");
      b.type = "button";
      b.textContent = "Revoke";
      b.onclick = function () { revoke(row.id); };
      p.appendChild(b);
      box.appendChild(p);
    });
  } catch (err) {
    box.textContent = "";
  }
}
async function revoke(id) {
  try {
    await api("DELETE", "/api/ssh/keys/" + encodeURIComponent(id));
    loadKeys();
  } catch (err) {
    document.getElementById("keyerr").textContent = err.message;
  }
}
document.getElementById("addkey").addEventListener("submit", async function (ev) {
  ev.preventDefault();
  document.getElementById("keyerr").textContent = "";
  try {
    await api("POST", "/api/ssh/keys", {
      public_key: document.getElementById("pubkey").value,
      label: document.getElementById("label").value
    });
    document.getElementById("pubkey").value = "";
    loadKeys();
  } catch (err) {
    document.getElementById("keyerr").textContent = err.message;
  }
});
if (token()) { showLogin(false); loadKeys(); } else { showLogin(true); }
</script>
</body>
</html>
`
