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

Current playtest build: `veilspan-5b4cfca.apk`
sha256 `7730d2acf9b4cbe1d8f2703ea5b38f6b1621631908aa08339e0fab52e8ce5091`
