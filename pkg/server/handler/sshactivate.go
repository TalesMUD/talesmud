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
		c.Data(http.StatusOK, "text/html; charset=utf-8", activatePage(tokenKey, gamemode.SignupOpen()))
	}
}

func setActivateHeaders(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	c.Header("X-Frame-Options", "DENY")
	c.Header("Content-Security-Policy", "frame-ancestors 'none'")
	c.Header("Referrer-Policy", "no-referrer")
	c.Header("X-Content-Type-Options", "nosniff")
}

func activatePage(tokenKey string, signup bool) []byte {
	key, err := json.Marshal(tokenKey)
	if err != nil || tokenKey == "" {
		key = []byte(`"talesmudDoorToken"`)
	}
	html := activateHTML
	if signup {
		html = strings.Replace(html, "<!--REGISTER-->", activateRegisterForm, 1)
		html = strings.Replace(html, "/*REGISTER_CSS*/", activateRegisterCSS, 1)
		html = strings.Replace(html, "//REGISTER_SCRIPT", activateRegisterScript, 1)
	}
	return []byte(strings.ReplaceAll(html, "__TOKEN_KEY__", string(key)))
}

const activateHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>TalesMUD sign-in</title>
<style>
body { font-family: sans-serif; max-width: 52rem; margin: 2rem auto; padding: 0 1rem; }
label { display: block; margin-top: 0.8rem; }
input { width: 100%; box-sizing: border-box; }
button { margin-top: 0.8rem; margin-right: 0.4rem; }
#detail { white-space: pre-wrap; }
#gates { display: flex; gap: 2rem; flex-wrap: wrap; align-items: flex-start; }
#gates form { flex: 1 1 16rem; }
/*REGISTER_CSS*/
</style>
</head>
<body>
<h1>TalesMUD sign-in</h1>
<div id="gates">
<form id="login">
<h2>Sign in</h2>
<label>Username <input id="username" autocomplete="username"></label>
<label>Password <input id="password" type="password" autocomplete="current-password"></label>
<p id="loginerr"></p>
<button type="submit">Sign in</button>
</form>
<!--REGISTER-->
</div>
<div id="panel" hidden>
<label>Code <input id="code" autocomplete="off"></label>
<button type="button" onclick="lookupCode()">Look up</button>
<p id="detail"></p>
<p id="keyline"></p>
<div id="decide" hidden>
<button type="button" onclick="decide('/api/ssh/device/confirm')">Confirm</button>
<button type="button" onclick="decide('/api/ssh/device/deny')">Deny</button>
</div>
<h2>SSH keys</h2>
<p>Keys are linked from an SSH sign-in. This page can revoke one.</p>
<div id="keys"></div>
<p id="keyerr"></p>
</div>
<script>
const tokenKey = __TOKEN_KEY__;
const codeInput = document.getElementById("code");
const params = new URLSearchParams(location.search);
if (params.get("code")) codeInput.value = params.get("code");
function token() { return localStorage.getItem(tokenKey) || ""; }
function showLogin(on) {
  document.getElementById("gates").hidden = !on;
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
//REGISTER_SCRIPT
let csrf = "";
async function lookupCode() {
  document.getElementById("detail").textContent = "";
  document.getElementById("keyline").textContent = "";
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
      "Only confirm if you started it yourself.";
    document.getElementById("keyline").textContent = "Key " + (data.key || "none");
    document.getElementById("decide").hidden = false;
  } catch (err) {
    document.getElementById("detail").textContent = err.message;
    document.getElementById("keyline").textContent = "";
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
if (token()) { showLogin(false); loadKeys(); } else { showLogin(true); }
</script>
</body>
</html>
`

const activateRegisterCSS = ``

const activateRegisterForm = `<form id="register">
<h2>Create account</h2>
<label>Username <input id="newuser" autocomplete="username"></label>
<label>Email <input id="newemail" type="email" autocomplete="email"></label>
<label>Password <input id="newpass" type="password" autocomplete="new-password"></label>
<p id="regerr"></p>
<button type="submit">Create account</button>
</form>`

const activateRegisterScript = `document.getElementById("register").addEventListener("submit", async function (ev) {
  ev.preventDefault();
  const res = await fetch("/api/auth/register", {
    method: "POST",
    headers: {"Content-Type": "application/json"},
    body: JSON.stringify({
      username: document.getElementById("newuser").value,
      email: document.getElementById("newemail").value,
      password: document.getElementById("newpass").value
    })
  });
  const data = await res.json().catch(function () { return {}; });
  if (!res.ok) {
    document.getElementById("regerr").textContent = data.error || "Could not create the account";
    return;
  }
  document.getElementById("regerr").textContent = "";
  document.getElementById("loginerr").textContent = data.message || "Sign in to continue.";
  var user = document.getElementById("username");
  var createdName = document.getElementById("newuser");
  if (user && createdName) user.value = createdName.value;
  if (user) user.focus();
});
if (params.get("signup") === "1") {
  var created = document.getElementById("newuser");
  if (created) created.focus();
}`
