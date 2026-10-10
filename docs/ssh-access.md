# SSH access

SSH is a second way into the same game process. It is off unless `ssh.enabled` is true. Classic web play is unchanged when it is off.

One process is one mode. A classic Auth0 process and a door local-auth process each run their own listener. SSH has no name-based routing, so the two modes use two ports.

| Mode | Suggested port | Player command |
| --- | --- | --- |
| Classic MUD | 2222 | `ssh -t -p 2222 veilspan.com` |
| Door | 2223 | `ssh -t -p 2223 door.veilspan.com` |

Port 22 stays the host's admin sshd. The game ports are above 1024. Opening them in the firewall is a later deploy step.

## Config

Put an `ssh:` block in the game-mode YAML. Every flag defaults off. Zero numbers are filled in at startup with the values below.

```yaml
ssh:
  enabled: false
  listen: "127.0.0.1:2222"          # required when enabled
  host_key_path: "/var/lib/talesmud/ssh/host_ed25519"  # required when enabled
  public_host: "veilspan.com"       # empty keeps the play link as /play
  public_port: 2222
  max_connections: 100
  max_per_ip: 5
  new_conns_per_ip_per_min: 20
  auth_fail_ban: {failures: 10, window: 10m, ban: 15m}
  auth_timeout: 30s
  idle_timeout: 30m                 # warns one minute before
  max_session: 0                    # 0 = no cap for accounts
  keys:
    enabled: false
    max_per_account: 10
    web_manage: true                # omit to leave web list/revoke on. A paste does not link a key.
  device:
    enabled: false
    ttl: 10m
    activate_url: "https://veilspan.com/activate"
    max_pending_per_ip: 3           # IPv6 counts a /64 as one address
    max_pending: 100                # pending codes for the whole process
    lookups_per_user_per_10m: 10
    lookups_per_ip_per_10m: 20
  guest:
    enabled: false                  # also needs guests allowed on the process
    max_concurrent: 10
    per_ip_per_hour: 5
    max_session: 30m                # warns five minutes before
  door:
    splash: ""                      # raw CP437 .ans inside the world pack
    activate_screen: ""
    charset_default: auto
    letterbox_fill: ""
  mud:
    history: 20
```

Environment overrides, when set: `SSH_ENABLED`, `SSH_LISTEN`, `SSH_HOST_KEY_PATH`, `SSH_PUBLIC_HOST`, `SSH_PUBLIC_PORT`, `SSH_ACTIVATE_URL`, `SSH_GUEST_ENABLED`, `SSH_DEVICE_ENABLED`. They turn features on or off and set limits. They never name a user, a key, or a token.

Per-address SSH caps, auth-fail bans, guest caps, and device-code caps treat an IPv6 /64 as one address. IPv4 stays one address. Pending device codes also stop at `max_pending` (default 100) for the whole process. The connection cap stays global. The address shown on the confirm page is the peer address, not the /64.

`ssh.enabled: false` does not open a listener. `GET /api/ssh/info` then returns `{"enabled": false}`.

A process with `ssh.enabled: true` and a missing listen address or host-key path stops before it serves HTTP.

## Host key

The listener generates one ed25519 key the first time it starts, if the file is absent. The directory is created mode `0700`. The file is created mode `0600` and is refused later if any group or world bit is set. Do not commit the key.

Startup logs the full public fingerprint, and `GET /api/ssh/info` returns the same value in `host_key_fingerprints`. The format is OpenSSH SHA256:

```text
SHA256:base64-unpadded-hash
```

User key fingerprints in logs are shortened to `SHA256:ab12…wxyz`. The SSH lobby shortens a device code to `BC**-****`. The HTTP access log replaces `access_token`, `code`, and `activate` query values with `[REDACTED]`. The play client stores `activate` and removes that parameter from the page URL before the login redirect. The log does not contain key blobs, device-code secrets, or session tokens.

Players can pin the host with:

```text
ssh -tt -p 2222 -o StrictHostKeyChecking=accept-new veilspan.com
```

Compare that fingerprint with the startup line or `/api/ssh/info` before trusting the prompt.

## How a player signs in

Password authentication is not offered. `exec`, subsystems, forwarding, agent forwarding, and X11 are refused. The server banner is `SSH-2.0-TalesMUD`.

- **Guest.** `ssh -t -p 2222 -l guest host`. Allowed only when `ssh.guest.enabled` is on and the process still allows guests. The account is a guest player: it cannot confirm or deny a device code, link a key, open a creator or admin route, or change its web profile (`PUT /api/user` is 403). The same connection does not become another account. It ends at `ssh.guest.max_session` (default 30 minutes) and on the idle timer.
- **Linked key.** `ssh -t -p 2222 -i ~/.ssh/id_ed25519 host`. A key that is already on the account enters the game with no extra question. Creators and admins use this same path.
- **Device code.** Any other user name opens a lobby. The screen shows a code like `BCDF-GHJK` and the activate URL. The player opens that URL, signs in on the web, and presses Confirm. Nothing is confirmed automatically.
  - Door (`auth: local`) serves `GET /activate` as a small sign-in page.
  - Classic (`auth: auth0`) answers `GET /activate?code=...` with a redirect to `/play/?activate=...`. The play client looks the code up and waits for Confirm or Deny.
- **Linking.** After a confirmed device login, the lobby asks whether to remember the computer. The offer is the key that signed. Yes stores a pending link and asks the player to reconnect. The next connection checks that same signature, names the account, and writes the key only when the whole answer is Y. N, a timeout, or quitting drops the pending link and disconnects. That session does not enter the account. Declining the first remember question still enters the account that confirmed the device code and does not store a link. A word that merely contains y is not yes. One fingerprint cannot be claimed by a second account while the first offer is live.
- **Web keys.** A signed-in, non-guest account can list and revoke keys that were linked from SSH. Classic uses the play-client account menu ("SSH keys"). Door uses the same `/api/ssh/keys` routes. `POST /api/ssh/keys` does not link a pasted key. A paste is not a signature, so it does not occupy the fingerprint or sign anyone in. The key is stored only after the device-link reconnect proves the signature and the owner presses Y. Deleting a key, or banning the account, closes that account's live SSH sessions in this process.

Account sessions use `ssh.max_session` when that value is greater than zero. Zero leaves them uncapped. Idle still applies.

A newer web or SSH session replaces the older one. The SSH side prints `Session moved to another client.`

## HTTP

Public:

- `GET /api/ssh/info` — `{enabled, host, port, host_key_fingerprints, guest_enabled, activate_url}` with the full host fingerprint.
- `GET /activate` — local page, or a redirect into `/play/?activate=` on a classic process.

Signed in (guest and banned accounts are refused):

- `GET /api/ssh/keys`
- `POST /api/ssh/keys` refuses a pasted key. Link the key from an SSH sign-in.
- `DELETE /api/ssh/keys/:id`
- `POST /api/ssh/device/lookup` with `{user_code}`
- `POST /api/ssh/device/confirm` and `POST /api/ssh/device/deny` with `{user_code, csrf}`

The lookup response shows the address, age, client version, and a shortened key fingerprint. Confirm and deny need the `csrf` value from that lookup, and an `Origin` header whose scheme and host match `activate_url`. A missing Origin is refused. The request still uses the bearer token. Key changes are limited to 20 per 10 minutes per account.

## World pack art

Door splash and activate screens are optional files named by `ssh.door.splash` and `ssh.door.activate_screen`. They are raw CP437 and must stay inside the world pack. Their CSI is kept. The activate screen may contain `{{CODE}}`, `{{URL}}`, and `{{EXPIRES}}`. The engine ships a plain fallback and does not embed a world's art. Player and creator text is sanitized before it is painted. A door frame drops OSC, DCS, C1, and bidi controls. A one-line label also drops CR and LF. SGR color in the frame is kept.

## Still to do before production

- Open TCP 2222 and 2223 on the host firewall.
- Set `ssh:` (or the `SSH_*` variables) on each service, with a host-key path outside the repo.
- Confirm the reverse proxy forwards `/activate` and `/api/ssh/*`.
- Decide whether guests are on for each mode. They stay off until that is set.
- No Auth0 dashboard change is required while activation stays under `/play`.
- An independent review of the auth paths is still outstanding. This document does not certify the feature.
