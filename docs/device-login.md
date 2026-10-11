# Device login for installable clients

Classic servers expose public `POST /api/device/start` and `POST /api/device/poll`.
Clients may use any descriptive client name; there is no client registration or
framework dependency. Door processes retain their existing local/SSH login and
do not mount these public app routes.

Set `ssh.device.activate_url` or `SSH_ACTIVATE_URL` to the public URL of the
existing `/activate` page, for example `https://game.example/activate`. This is
also the allowed Origin for browser lookup/confirm/deny. An absent or invalid
URL makes start return 503. Public app login works even with the SSH listener or
SSH device lobby disabled. Reverse proxies must forward `/activate`,
`/api/device/*`, `/api/ssh/device/*` and `/ws`.

## Client contract

Start with JSON:

```json
{"client":"my-terminal","client_version":"0.1.0","proto":1}
```

The 200 response is:

```json
{"user_code":"BCDF-GHJK","device_code":"<256-bit-random-hex-secret>","verification_uri":"https://game.example/activate","verification_uri_complete":"https://game.example/activate?code=BCDF-GHJK","expires_in":600,"interval":5}
```

Only the human `user_code` goes in the browser. Keep `device_code` private. It
is a distinct random secret; both codes are stored as hashes in the same
in-process `pkg/devicecode` store used by SSH. Restart drops outstanding codes.
TTL follows `ssh.device.ttl`, capped at ten minutes for app codes.

Wait at least `interval` seconds, then poll with JSON:

```json
{"device_code":"<secret-from-start>"}
```

Pending/error responses are HTTP 400 with one of:

```json
{"error":"authorization_pending"}
{"error":"slow_down"}
{"error":"access_denied"}
{"error":"expired_token"}
```

`slow_down` means add five seconds to the polling interval. Early polls increase
the server's per-device interval by five seconds and restart that wait. Poll IP
limits also return `slow_down`. Denial and expiry end the flow. Invalid, unknown,
expired and already-consumed secrets all return `expired_token`, with no account
details. A transient issuer failure returns 503 `{"error":"unavailable"}` and
keeps the confirmation available for a later poll.

After the account owner explicitly confirms, the next permitted poll returns
HTTP 200 once:

```json
{"token":"<signed-account-session-JWT>","expires_in":1800}
```

Store the bearer privately (directory 0700, token file 0600). Send it as
`Authorization: Bearer <token>` on character HTTP requests and the `/ws` upgrade.
There is no refresh token; expiry requires another login. Both public responses
use `Cache-Control: no-store`. Codes and tokens must never be logged.

## Browser confirmation

The existing `/activate` page says “Connect this terminal/app”. It supports SSH
and app codes through the same authenticated `/api/ssh/device/lookup`, `confirm`
and `deny` routes. Lookup obtains the per-account CSRF nonce; confirm/deny require
that nonce plus an Origin matching the configured activation URL. Lookup never
confirms automatically. Guests and banned accounts cannot confirm or deny.
App lookup adds `client_kind: "app"` and a descriptive `client` string to the
existing IP, creation time, mode, client version, key and CSRF response. Client
metadata grants no authority. SSH lookup responses and SSH key/lobby behavior
stay unchanged.

## Session signing and limits

`pkg/authsession` is an explicit server issuer, separate from Auth0, guest tokens
and Door's local issuer. It signs HS256 JWTs containing `iss:
"talesmud-server-session"`, `aud: "talesmud-api"`, `kind: "device-session"`, the
confirmed account's entity ID as `sub`, `iat`, mandatory `exp` (30 minutes), and
its existing `SessionVersion` as `sv`. Middleware verifies signature, algorithm,
issuer, audience, expiry and session version, loads the existing account without
creating one, and rejects guests and bans.

Signing reuses `authlocal.ResolveSecret`: `SESSION_SECRET` wins, then game-mode
`session_secret`, then `secret_path` (default `data/session.key`, generated on
first use with mode 0600). Configure a persistent server secret/key path; rotating
the key invalidates these sessions. No Auth0 or guest secret is reused implicitly.
This adds no database schema or dependency.

Default public limits per process:

- Start: 10 attempts per IP and 100 globally per ten minutes, including starts
  refused by pending caps. Shared SSH/app pending caps follow existing
  `ssh.device.max_pending_per_ip` (3) and `max_pending` (100).
- Poll: 60 requests per IP per minute, including bad secrets, plus the growing
  interval per device secret. Changing IP does not reset a device's interval.
- IPv6 addresses share a /64 bucket. HTTP addresses follow the existing trusted
  proxy configuration; public clients cannot choose their own rate-limit bucket
  through forwarded headers from an untrusted peer.

## Verification

`pkg/devicecode/poll_test.go` uses a controlled clock to check interval backoff,
limits, TTL, denial and concurrent single consumption. `pkg/authsession` tests
reject invalid signatures/claims, guests, bans and revoked sessions.
`pkg/server/device_flow_test.go` runs a local HTTP server with temporary SQLite
and a local JWKS service: a test-signed Auth0 identity confirms, the returned
account token reads its character list, and the real MUD `/ws` handler upgrades
and sends a welcome message. It also checks guest rejection, Origin/CSRF,
denial, expiry, HTTP limits and middleware rejection. Existing SSH tests remain
the regression coverage. A separate local startup smoke also exercised the real
`NewApp().Run()` wiring with the SSH listener and lobby disabled, temporary
SQLite, a test-signed Auth0 identity and local JWKS: start/pending/confirm/poll,
character HTTP access, `/ws` upgrade and a welcome message all passed.
These checks do not exercise a production Auth0 login,
browser clicks, reverse proxy, deployment or VPS.
