# Finish Dirty Merchant And Quest Work Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Finish the pre-existing dirty feature work around merchant trading, quest areas, collect quest initialization, quest-abandon refresh, and inventory sell UI without broad refactors.

**Architecture:** Keep the existing command/service/UI structure. Add focused backend tests for quantity accounting and runtime merchant stock behavior, then make the smallest code changes to satisfy them. Keep Creator and client UI conventions aligned and update docs for the completed behavior.

**Tech Stack:** Go 1.24 via `/usr/local/go/bin/go`, SQLite-backed service facade, Svelte/Rollup mud client, Svelte/Vite creator app.

---

### Task 1: Backend Quest Collect Quantity Accounting

**Files:**
- Modify: `pkg/service/quests.go`
- Test: `pkg/service/quests_test.go`

- [x] **Step 1: Write failing test**
  Add a test where a character accepts a collect quest after already holding a stackable item with `Quantity: 3`; expected objective progress is `Current: 3`, not `1`.

- [x] **Step 2: Run red test**
  Run: `/usr/local/go/bin/go test ./pkg/service -run TestAcceptQuestPrefillsCollectObjectiveWithStackQuantity`
  Expected: FAIL showing current progress is `1`.

- [x] **Step 3: Implement quantity-aware counting**
  In `AcceptQuest`, when an inventory item matches the collect objective target, add `item.Quantity` when positive, otherwise add `1`.

- [x] **Step 4: Run green test**
  Run: `/usr/local/go/bin/go test ./pkg/service -run TestAcceptQuestPrefillsCollectObjectiveWithStackQuantity`
  Expected: PASS.

### Task 2: Merchant Runtime Stock Safety

**Files:**
- Modify: `pkg/mudserver/game/commands/trade.go`
- Test: `pkg/mudserver/game/commands/trade_test.go`

- [x] **Step 1: Write failing tests**
  Add command tests proving stackable `buy` quantities are delivered as one stack, merchant stock is decremented, and `list` restocks the instance after `RestockMinutes` has elapsed.

- [x] **Step 2: Run red tests**
  Run: `/usr/local/go/bin/go test ./pkg/mudserver/game/commands -run 'TestBuyDecrementsMerchantStock|TestListRestocksMerchantInventory'`
  Expected: FAIL if stock changes are not applied through the NPC manager update boundary.

- [x] **Step 3: Implement stackable capacity fix**
  Keep existing runtime instance mutation and fix stackable buy capacity so quantity is not capped by open slot count when a stack can hold it.

- [x] **Step 4: Run green tests**
  Run: `/usr/local/go/bin/go test ./pkg/mudserver/game/commands -run 'TestBuyDecrementsMerchantStock|TestListRestocksMerchantInventory'`
  Expected: PASS.

### Task 3: Client Inventory Sell UI Polish

**Files:**
- Modify: `public/mud-client/src/game/widgets/InventoryWidget.svelte`

- [x] **Step 1: Fix rough edges**
  Keep the sell action hidden for quest, bound, and equipped items. Ensure the sell quantity input is numeric, clamped from `1` to the available quantity, and closes the detail overlay cleanly when the popup opens.

- [x] **Step 2: Build mud client**
  Run: `/home/atla/.nvm/versions/node/v20.20.0/bin/node ./node_modules/rollup/dist/bin/rollup -c` from `public/mud-client`.
  Expected: exit 0 and updated `public/bundle.js`.

### Task 4: Creator Quest Area And Docs

**Files:**
- Modify: `public/app/src/creator/QuestsEditor.svelte`
- Modify: `PROJECT.md`
- Modify: `ARCHITECTURE.md`
- Modify: `FEATURES.md`

- [x] **Step 1: Verify Creator behavior**
  Build creator app with `/home/atla/.nvm/versions/node/v20.20.0/bin/node node_modules/vite/bin/vite.js build` from `public/app`.

- [x] **Step 2: Update docs**
  Document quest `area`, merchant trading/restock behavior, and inventory sell UI at the appropriate level in project docs.

### Task 5: Final Verification

**Files:**
- No additional production files expected.

- [x] **Step 1: Backend verification**
  Run `/usr/local/go/bin/go test $(/usr/local/go/bin/go list ./... | grep -v '/pkg/mudserver/game/combat$')`.

- [x] **Step 2: Vet**
  Run `/usr/local/go/bin/go vet ./...`.

- [x] **Step 3: Frontend verification**
  Run creator and mud-client builds.

- [x] **Step 4: Known deferred failure**
  Do not claim `/usr/local/go/bin/go test ./...` passes until the separate combat balance work is approved and completed.
