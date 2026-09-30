# Veilspan `/app` Android download page (2026-09-30)

## Live URLs
- Page: https://veilspan.com/app  (also https://veilspan.com/app/)
- APK:  https://veilspan.com/app/downloads/veilspan-7cbb5df.apk
- sha256: `efa262db9d3ec469b96eea6b750842fcc7a424afd5a08b4e7e4b6215b5375aa2`

## How it is served
- `LANDING_PATH=./public/landing` on Veilspan (`talesmud.service` :8010).
- LandingMiddleware serves any file/dir under that tree from the OS filesystem
  (same path used for `/`, `/docs`, `/map`).
- No engine binary rebuild or SIGTERM was required for this page.
- APKs are **not** in git; rsync/scp onto the VPS only.

## VPS paths
- HTML: `/home/atla/dev/talesmud/public/landing/app/index.html`
- APK:  `/home/atla/dev/talesmud/public/landing/app/downloads/veilspan-7cbb5df.apk`
- Docs: `/home/atla/dev/talesmud/public/landing/app/downloads/README.md`

## Upload recipe (clawdbot)
```bash
scp -i ~/.ssh/veilspan_vps \
  ~/dev/veilspan-client/.director/ux-audit/veilspan-dev-7cbb5df.apk \
  atla@100.83.205.104:/home/atla/dev/talesmud/public/landing/app/downloads/veilspan-7cbb5df.apk
```

## Notes
- Debug / playtest-signed → Play Protect may warn (called out on the page).
- `public/app/` in the engine repo is the **admin SPA**, unrelated to this marketing `/app` route.
- nginx (`/home/atla/dev/nginx`) still proxies `/` → :8010; passwordless sudo is not available
  for an nginx reload, so APK MIME comes from Go `c.File` + `/etc/mime.types` (apk mapped).
- Door `:8020` (pid 758959) was not touched.
- veilspan-client stays private; this page only hosts the APK binary publicly.

## Branding
Gold/dark Veilspan vibe (Cinzel + Cormorant, amber `#e8a849` / gold `#c8a84e` on abyss `#06080c`).
