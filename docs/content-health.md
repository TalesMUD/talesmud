# Content health

`tales -check <importFolderName>` reads `import/<name>` and reports content problems without a database. The same rules run in the Creator on `GET /api/health`.

```bash
go run ./cmd/tales -check game
go run ./cmd/tales -check game -json
go run ./cmd/tales -check game -fail-on=warning
```

`-check` and `-import` cannot be used together. The default `-fail-on=error` exits 1 when any unmuted error remains. `-fail-on=warning` also exits 1 on unmuted warnings. Info hits and muted rules do not fail the run. The text report shows about 30 hits per rule. `-json` prints every hit.

Pack rules live in `import/<name>/data/rules/*.yaml`. A missing directory adds no pack rules. The server keeps the rules that were present at the last successful import and compares the database with that import baseline.

Muted rule ids are stored on server settings (`mutedHealthRuleIDs`). `PUT /api/health/mute` with `{"ruleId":"...","muted":true}` updates them. Creators call:

- `GET /api/health`
- `PUT /api/health/mute`
- `GET /api/health/drift`
- `GET /api/health/drift/export?type=&id=` (importer-format YAML)

`/api/diagnostics/world` and `/api/world/validation` keep their existing response shapes.

An item is obtainable when it sits in a reachable room, drops from a reachable NPC, is sold by a reachable merchant, or is granted by `giveItem` from a script the player can run. A quest reward counts too, once that quest can complete and its prerequisites can be finished before the quest that needs the item. The check repeats until it stops changing. A cycle does not make the item obtainable.

Reachability starts at the server setting `startRoomID` when that room exists, and otherwise at the engine default start room. `tales -check` does not open the database, so it uses the engine default.

The printed content commit is that folder's own git HEAD. A copy that sits inside another checkout does not inherit the parent commit. It reads the first line of `CONTENT_COMMIT` or `.content-commit` in the folder, and otherwise prints `unknown`.

`GET /api/health` also warns with `deploy-tree-dirty` when the server's working directory is its own git checkout and `git status` shows tracked changes. Each changed path is one hit. Paths under `import/` are skipped. The check is one `git status` per health run, with a short timeout, and it is skipped when the directory is not a checkout or `git` is missing. `tales -check` does not run this rule.

`GET /api/quests/:id/debug` uses this same quest check for one quest. Each character row names the open objective. The Creator opens it at `/creator/quests/debug?id=` and follows that id when the URL changes. Health quest hits link there. An admin can reset the open step.

## GitHub Actions

```yaml
name: content-health
on: [pull_request]
jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.24"
      - name: Engine
        uses: actions/checkout@v4
        with:
          repository: talesmud/talesmud
          path: engine
      - name: Check
        working-directory: engine
        run: |
          ln -s "$GITHUB_WORKSPACE" import/game
          go run ./cmd/tales -check game -fail-on=error
```

Point the symlink at the content checkout. The engine directory is the Tales module that contains `cmd/tales`.
