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

Current playtest build: `veilspan-1adb986.apk`
sha256 `c40a7f105f468153b5f42666032afebf8db4bf7ab80e6531727b39bae4d03d77`
