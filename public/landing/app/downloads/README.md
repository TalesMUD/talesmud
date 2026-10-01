# Veilspan Android APK downloads

APKs are **not** committed to git (large binaries). Deploy to the VPS landing
tree and serve via `LANDING_PATH` (`./public/landing`) or nginx.

## VPS path

```
/home/atla/dev/talesmud/public/landing/app/downloads/
```

## Public URL pattern

```
https://veilspan.com/app/downloads/veilspan-<gitsha>.apk
```

## Upload from clawdbot

```bash
scp -i ~/.ssh/veilspan_vps \
  /path/to/veilspan-dev-<sha>.apk \
  atla@100.83.205.104:/home/atla/dev/talesmud/public/landing/app/downloads/veilspan-<sha>.apk
```

Current playtest build: `veilspan-ae4412c.apk`
sha256 `c8a08cefd6c504847a35bb95632118394a83e1227d4764aad7d659ee201638d4`
