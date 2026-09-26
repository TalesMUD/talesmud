(function () {
  let TOKEN_KEY = "talesmudDoorToken";
  const form = document.getElementById("form");
  const err = document.getElementById("err");
  const emailRow = document.getElementById("emailRow");
  const tokenRow = document.getElementById("tokenRow");
  const toggle = document.getElementById("toggle");
  const forgot = document.getElementById("forgot");
  const submit = document.getElementById("submit");
  const passLabel = document.getElementById("passLabel");
  const userLabel = document.getElementById("userLabel");

  let mode = "login";
  let term = null;
  let ws = null;
  let inputMode = "hotkey";
  let line = "";

  function setMode(next) {
    mode = next;
    err.textContent = "";
    err.style.color = "#f66";
    const username = document.getElementById("username");
    const email = document.getElementById("email");
    const password = document.getElementById("password");
    const token = document.getElementById("token");
    const showUser = mode === "login" || mode === "register";
    const showEmail = mode === "register" || mode === "forgot";
    const showPassword = mode !== "forgot";
    const showToken = mode === "reset";
    username.hidden = !showUser;
    userLabel.hidden = !showUser;
    username.required = showUser;
    emailRow.hidden = !showEmail;
    email.required = showEmail;
    tokenRow.hidden = !showToken;
    token.required = showToken;
    password.hidden = !showPassword;
    passLabel.hidden = !showPassword;
    password.required = showPassword;
    if (mode === "register") {
      submit.textContent = "Create account";
      toggle.textContent = "I already have an account";
      passLabel.textContent = "Password";
    } else if (mode === "forgot") {
      submit.textContent = "Send reset token";
      toggle.textContent = "Back to sign in";
    } else if (mode === "reset") {
      submit.textContent = "Set new password";
      toggle.textContent = "Back to sign in";
      passLabel.textContent = "New password";
    } else {
      submit.textContent = "Enter";
      toggle.textContent = "Create an account";
      passLabel.textContent = "Password";
    }
  }

  toggle.addEventListener("click", function () {
    setMode(mode === "login" ? "register" : "login");
  });
  forgot.addEventListener("click", function () {
    setMode(mode === "forgot" ? "login" : "forgot");
  });

  form.addEventListener("submit", function (ev) {
    ev.preventDefault();
    err.textContent = "";
    const username = document.getElementById("username").value.trim();
    const email = document.getElementById("email").value.trim();
    const password = document.getElementById("password").value;
    const token = document.getElementById("token").value.trim();
    if (mode === "forgot") {
      post("/api/auth/forgot", { email: email }).then(function () {
        setMode("reset");
        err.style.color = "#9c9";
        err.textContent = "If that email is registered, a reset token was sent. This page never receives the token.";
      }).catch(showErr);
      return;
    }
    if (mode === "reset") {
      post("/api/auth/reset", { token: token, password: password }).then(function () {
        setMode("login");
        err.style.color = "#9c9";
        err.textContent = "Password updated. Sign in with the new one.";
      }).catch(showErr);
      return;
    }
    const path = mode === "register" ? "/api/auth/register" : "/api/auth/login";
    const body = mode === "register"
      ? { username: username, email: email, password: password }
      : { username: username, password: password };
    post(path, body).then(function (data) {
      if (!data.token) throw new Error("no session token");
      connect(data.token);
    }).catch(showErr);
  });

  function showErr(error) {
    err.style.color = "#f66";
    err.textContent = (error && error.message) || "Request failed";
  }

  function post(url, body) {
    return fetch(url, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body)
    }).then(function (res) {
      return res.json().catch(function () { return {}; }).then(function (data) {
        if (!res.ok) throw new Error(data.error || ("HTTP " + res.status));
        return data;
      });
    });
  }

  function connect(token) {
    localStorage.setItem(TOKEN_KEY, token);
    document.getElementById("auth").hidden = true;
    document.getElementById("stage").hidden = false;
    ensureTerm();
    const proto = location.protocol === "https:" ? "wss" : "ws";
    ws = new WebSocket(proto + "://" + location.host + "/ws?access_token=" + encodeURIComponent(token));
    ws.onmessage = function (ev) {
      let msg;
      try { msg = JSON.parse(ev.data); } catch (e) { return; }
      if (!msg || (msg.type !== "door_frame" && msg.type !== "doorFrame") || !msg.ansi) return;
      if (msg.accepts && msg.accepts[0] === "line") inputMode = "line";
      else if (msg.accepts && msg.accepts[0] === "hotkey") inputMode = "hotkey";
      else inputMode = msg.inputMode || "hotkey";
      if (inputMode === "hotkey") line = "";
      term.write(msg.ansi);
    };
    ws.onclose = function () {
      term.write("\r\n\x1b[1;31mConnection closed.\x1b[0m\r\n");
    };
    ws.onerror = function () {
      err.textContent = "WebSocket failed";
    };
  }

  function ensureTerm() {
    if (term) return;
    if (typeof Terminal === "undefined") {
      err.textContent = "xterm.js failed to load";
      return;
    }
    term = new Terminal({
      cols: 80,
      rows: 25,
      fontFamily: "Consolas, 'Courier New', monospace",
      fontSize: 16,
      theme: { background: "#000000", foreground: "#d0d0d0", cursor: "#33ff66" },
      cursorBlink: true,
      scrollback: 0,
      convertEol: false
    });
    term.open(document.getElementById("term"));
    fitFont();
    window.addEventListener("resize", fitFont);
    term.onData(onData);
  }

  function fitFont() {
    if (!term) return;
    const w = Math.max(320, window.innerWidth - 16);
    const h = Math.max(240, window.innerHeight - 16);
    const size = Math.max(10, Math.min(28, Math.floor(w / (80 * 0.62)), Math.floor(h / (25 * 1.15))));
    term.setOption("fontSize", size);
  }

  function onData(data) {
    if (!ws || ws.readyState !== 1) return;
    if (inputMode === "line") {
      if (data === "\r") {
        const text = line;
        line = "";
        term.write("\r\n");
        ws.send(JSON.stringify({ type: "door_key", key: text }));
        return;
      }
      if (data === "\x7f" || data === "\b") {
        if (line.length > 0) {
          line = line.slice(0, -1);
          term.write("\b \b");
        }
        return;
      }
      if (data.length === 1 && data >= " " && data <= "~" && line.length < 48) {
        line += data;
        term.write(data);
      }
      return;
    }
    if (data.length === 1 && /[a-zA-Z0-9?]/.test(data)) {
      ws.send(JSON.stringify({ type: "door_key", key: data }));
    }
  }

  function applyBranding(cfg) {
    cfg = cfg || {};
    if (cfg.tokenKey) TOKEN_KEY = cfg.tokenKey;
    if (cfg.title) {
      document.title = cfg.title;
      var heading = document.getElementById("title");
      if (heading) heading.textContent = cfg.title;
    }
    if (cfg.subtitle) {
      var sub = document.getElementById("sub");
      if (sub) sub.textContent = cfg.subtitle;
    }
    var forgotButton = document.getElementById("forgot");
    if (forgotButton) forgotButton.hidden = !cfg.forgotEnabled;
    var existing = localStorage.getItem(TOKEN_KEY);
    if (!existing) return;
    fetch("/api/auth/me", { headers: { Authorization: "Bearer " + existing } })
      .then(function (res) {
        if (!res.ok) {
          localStorage.removeItem(TOKEN_KEY);
          return;
        }
        connect(existing);
      })
      .catch(function () {
        localStorage.removeItem(TOKEN_KEY);
      });
  }

  fetch("/api/door/config").then(function (res) {
    if (!res.ok) return {};
    return res.json();
  }).then(applyBranding).catch(function () { applyBranding({}); });
})();
