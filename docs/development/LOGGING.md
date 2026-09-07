# TalesMUD logging

## Where logs go

| Sink | Path / unit | Retention |
|------|-------------|-----------|
| Application file (primary) | `logs/talesmud.log` under the process working directory (`WorkingDirectory` in systemd; on veilspan: `/home/atla/dev/talesmud/logs/talesmud.log`) | **7 days** via lumberjack (`MaxAge=7`, `MaxSize=50MB`, compressed backups) |
| stderr / journald | `journalctl -u talesmud` | Distro journal vacuum (not app-controlled) |
| Optional logrotate | `config/logrotate.talesmud` → `/etc/logrotate.d/talesmud` | 7 daily copies (safety net; copytruncate) |

Override file path with `LOG_FILE` (see `.env.example`).

## Structured fields

Prefer `logrus.WithField` / `WithFields`. WebSocket lifecycle logs include:

- `WS connect` / `WS upgrade` / `WS replace-existing` / `WS read error` / `WS send error` / `WS close`
- `UserJoined` / `UserQuit` / `selectcharacter`
- Auth failures: `Auth failure` with `ip`, `reason` (never log raw `access_token`)

Gin access logs redact `access_token=` query values as `[REDACTED]`.

## VPS wiring

1. Ensure `logs/` exists under the service working directory (created automatically on start).
2. Set `LOG_FILE=logs/talesmud.log` in `.env` if desired (default is already that path).
3. Optional: `sudo cp config/logrotate.talesmud /etc/logrotate.d/talesmud`
4. After deploy: `ls -la logs/` and `tail -n 50 logs/talesmud.log`
