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

Current playtest build: `veilspan-d43d2d0.apk`
sha256 `19072d6678cd5326661522ba02033bd98462da25c47ece60fde7664b2cc76652`
