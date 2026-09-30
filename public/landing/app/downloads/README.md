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

Current playtest build: `veilspan-25fe7a8.apk`
sha256 `2e80ed58da6e21649f3b72dbd58f7dcf60a125b72da0583a93fd7e2473b03d28`
