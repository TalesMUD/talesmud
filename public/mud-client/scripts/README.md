# Verify responsive presentation

The fixture runner uses Chromium. Prefer the local protocol fixture for repeated checks; it creates no live guests. Use the live backend for the baseline and real game walks. It injects client messages to exercise long rewards, six queued quest cards, defeat, populated inventory, party controls, and pending loot rolls. After programmatic scrolling, the runner waits two animation frames for layout before hit testing. It saves screenshots, loaded asset hashes, bounds, and three hit tests per target at 375×812, 768×1024, 1440×900, 812×375, and 844×390. The phone uses touch and DPR 3. The other touch sizes use DPR 2.

Install Playwright and `ws` in a separate tools directory. Set `UX_PLAYWRIGHT_PACKAGE` and `UX_PROXY_PACKAGE` to that directory's `package.json` if the tools are not installed in this client.

Run these commands from the repository root. The proxy serves the local client build and forwards API and WebSocket traffic to `UX_UPSTREAM`. Set that URL to the game backend under test. It does not log credentials or request URLs.

```sh
npm --prefix public/mud-client run build
UX_UPSTREAM=https://your-game.example UX_PROXY_PACKAGE=/path/to/tools/package.json node public/mud-client/scripts/uxTopLocalServer.mjs
```

In a second terminal, run the proof.

```sh
UX_PLAYWRIGHT_PACKAGE=/path/to/tools/package.json UX_OUT=/tmp/ux-top-proof node public/mud-client/scripts/uxTopFixtures.mjs
```

`UX_PORT` changes the proxy's default port 8016. `UX_PUBLIC` changes its default asset directory `public/mud-client/public`. `UX_BASE` changes the runner's default URL `http://127.0.0.1:8016/play/?v=wowf1`. `UX_PHASE` names the output group and defaults to `local`. `UX_SIZES` selects comma-separated viewport IDs. `UX_PHASE=before` retains failed baseline hit tests and continues past blocked outcome dismissal.

The fixture messages do not change backend inventory, quests, or combat. Close the runner to close its guest sockets. Keep a real guest walk alongside these deterministic layout checks for dialog, shop, movement, and server combat behavior.

To run without a live backend, set `UX_FIXTURE=1` on both processes. The local protocol fixture provides a generic room, guide, merchant, character, and guest session. The runner also taps the NPC cards and checks dialog and shop portals.

```sh
UX_FIXTURE=1 UX_PORT=8017 UX_PROXY_PACKAGE=/path/to/tools/package.json node public/mud-client/scripts/uxTopLocalServer.mjs
UX_FIXTURE=1 UX_BASE='http://127.0.0.1:8017/play/?v=wowf1' UX_PLAYWRIGHT_PACKAGE=/path/to/tools/package.json UX_OUT=/tmp/ux-top-proof node public/mud-client/scripts/uxTopFixtures.mjs
```
