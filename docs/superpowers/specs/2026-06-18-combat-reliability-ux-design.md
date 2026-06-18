# Combat Reliability and Player Combat UX Design

## Goal

Make combat understandable, balanced, and stable across the backend combat loop and the Svelte game client.

This work reopens deferred combat balance work and treats combat as the current focus. The target end state is not just cleaner UI text: backend outcomes must be reliable, balance tests must run against the same rules as runtime combat, and the frontend must present combat state clearly enough for players to choose targets, understand HP/status changes, and see rewards or penalties.

## Current Evidence

- `pkg/mudserver/game/combat/combat_sim_test.go` has failing balance assertions when run with Go 1.24.12. The failures cluster around Mage and Druid win rates against trivial, level 2, and same-level normal enemies.
- Running the same tests with the default `go` on PATH fails before compilation because it is Go 1.18.1 and cannot parse the repo's `go 1.24.0` and `toolchain` directive. Verification commands must use `/usr/local/go/bin/go` or another Go 1.24+ binary.
- The balance loader in `pkg/mudserver/game/balance/difficulty.go` searches relative paths only up to `../..`, so package tests under `pkg/mudserver/game/combat` miss `config/combat_balance.yaml` and silently use defaults.
- Combat already has an auto-attack and queued-action model in `pkg/entities/combat/instance.go` and `pkg/mudserver/game/game_combat.go`: `AutoAttackTargetID`, `QueuedAction`, `QueuedTargetID`, and `QueuedSkillID`.
- Combat messages sent through `pkg/mudserver/game/messages` are mostly human-readable strings. `public/mud-client/src/game/Client.js` forwards those strings to the terminal and overlay, but `public/mud-client/src/game/MUDXPlusStore.js` does not retain structured combat participants, targets, statuses, outcomes, rewards, or penalties.
- Mobile controls in `public/mud-client/src/game/mobile/MobileActionBar.svelte` expose generic Attack, Defend, Flee, and Status buttons during combat, but they do not provide an enemy target picker or current target feedback.

## Product Baseline

Combat balance will be tuned around solo, skill-enabled combat for starter classes.

Reasoning:
- The simulator already creates characters with available class skills equipped via `simutil.CreateCharacter`.
- The live combat loop already processes skills, cooldowns, mana, status effects, queued skills, and auto-attacks.
- Balancing every class around basic weapon attacks would make caster/support classes less distinct and would work against the current skills system.

Basic attacks still need to be valid fallbacks when a skill is unavailable, on cooldown, lacks mana, or has no valid target.

## Backend Design

### Balance Test Reliability

The combat balance loader must find `config/combat_balance.yaml` from package test working directories and normal runtime working directories. The loader should prefer the explicit config path and then walk upward from the current working directory until it finds the repo config file.

Balance tests should be split into two categories:
- Regression tests with stable pass/fail thresholds for expected player-facing outcomes.
- Exploratory analysis tests that are skipped by default or converted into simulator CLI documentation, so normal test runs do not depend on print-heavy what-if output.

The regression suite should prove:
- Starter classes can defeat trivial enemies at level 1 with the configured skill-enabled baseline.
- Level 5 characters can reliably defeat level 2 enemies.
- Same-level normal enemies are challenging but not class-breaking.
- Bosses remain meaningfully harder than normal enemies.

### Combat Outcome Hardening

Combat cleanup should be idempotent from the game layer's perspective: once an instance reaches victory, defeat, fled, or timeout, reward/penalty handling should run once, participant combat flags should clear, NPC state should match survival state, and the instance should be removed from the manager.

Victory handling should:
- Award XP and gold only to living, non-fled players.
- Use configured enemy rewards when present and level/difficulty formulas otherwise.
- Drop loot in the origin room when room data is available.
- Notify players with a structured reward summary plus the existing readable victory text.
- Send character and inventory updates after XP, gold, level, and inventory effects are persisted.

Defeat handling should:
- Apply the configured death gold loss percent instead of hardcoding the penalty.
- Respawn each defeated player with the configured death respawn HP percent, minimum 1 HP.
- Preserve player room/bind behavior as currently implemented unless a bind-point-specific bug is found during implementation.
- Notify players with a structured penalty/respawn summary plus readable defeat text.
- Ensure later cleanup does not overwrite respawn HP back to zero.

Fled and timeout handling should:
- Clear player and NPC combat flags.
- Keep living NPCs alive and idle unless a flee-specific NPC state is intentionally triggered.
- Avoid rewards and death penalties.

### Action Semantics

The implementation should make these rules explicit in code, tests, and messages:

- `attack <target>` during combat switches the persistent auto-attack target and queues an attack for the next player turn.
- `attack` without a target only works when exactly one living enemy exists; otherwise it lists valid targets.
- `defend` queues a defensive stance for the next player turn. The defense bonus applies until that combatant's next turn is advanced and is then cleared.
- `flee` queues a flee attempt for the next player turn. Success marks the combatant as fled; failure consumes the turn.
- Status effects tick at the start of the affected combatant's turn. DoT/HoT, stun, buff/debuff duration, dodge, and mana shield behavior must be visible in combat logs and snapshots.

### Structured Combat Snapshot

Backend combat messages should continue carrying `message` text for terminal compatibility, and add a structured snapshot that clients can consume without parsing terminal text.

The snapshot should include:
- `instanceId`, `round`, `state`, `currentTurnId`, `currentTurnName`, and `turnTimeRemaining`.
- `players` and `enemies`, each with `id`, `type`, `name`, `isAlive`, `hasFled`, `maxHp`, `currentHp`, `maxMana`, `currentMana`, `defenseBonus`, and `statusEffects`.
- Current player's `autoAttackTargetId`, `queuedAction`, `queuedTargetId`, and `queuedSkillId`.
- `validTargets`, containing living enemies available to the current player.
- `lastAction`, derived from the latest combat log entry.
- `outcome` on combat end: `victory`, `defeat`, `fled`, or `timeout`.
- `rewards` on victory: XP, gold, loot names, defeated enemy names.
- `penalties` on defeat: gold lost and respawn HP.

The snapshot should be generated from the authoritative `CombatInstance` after each relevant state change. Text should remain the source for the terminal log; the snapshot should be the source for controls and compact HUDs.

## Frontend Design

### Store State

`MUDXPlusStore.js` should gain a `combat` object with:
- `active`
- `instanceId`
- `round`
- `state`
- `players`
- `enemies`
- `validTargets`
- `autoAttackTargetId`
- `queuedAction`
- `queuedTargetId`
- `statusEffects`
- `lastAction`
- `outcome`
- `rewards`
- `penalties`

`Client.js` should update this store state from structured combat fields on `combatStart`, `combatTurn`, `combatAction`, `combatStatus`, and `combatEnd`. Existing terminal rendering and overlay messages should remain.

### Combat Log Clarity

Combat log rendering should distinguish:
- Player damage and enemy damage.
- Misses, dodges, critical hits, blocks, and failed flees.
- Status applications, ticks, expirations, and stun skips.
- Queued actions and target changes.
- Victory rewards and defeat penalties.

This can be implemented by classifying `lastAction` metadata when available and falling back to existing text for older messages.

### Enemy Target Controls

Enemy controls should come from `combat.validTargets` during combat, not from stale room NPC state. Each enemy target button should show:
- Enemy name.
- HP value and bar.
- Status effect badges.
- Selected target state.
- A command action that sends `attack <enemy name>` or an ID-backed command if the backend adds ID targeting.

The room `EntityPanel` can continue to show NPCs outside combat. During combat, target selection should be a dedicated combat control so players understand they are switching focus rather than starting a new encounter.

### HP, Status, Reward Feedback

Player HP/mana should continue syncing through `characterUpdate`, and the combat snapshot should provide combat-specific participant HP for all combatants.

The UI should show:
- Player HP and mana changes immediately after combat actions.
- Status effect names and durations for the player and enemies.
- Queued action confirmation.
- Victory reward summary.
- Defeat penalty and respawn HP summary.

### Mobile Combat Controls

Mobile controls should keep 44px minimum touch targets and add:
- A combat target sheet or compact horizontal target row.
- Attack, Defend, Flee, and Status buttons that reflect queued action state.
- Current target and low-HP feedback in the mobile header or action area.
- Reward/penalty feedback in a compact combat result panel or overlay message.

The mobile UI should avoid requiring typed commands for normal combat decisions.

## Testing Strategy

### Backend

Use test-driven implementation for behavior changes:
- Write failing tests for config path discovery.
- Write failing regression tests for balance outcomes after config path discovery is fixed.
- Write targeted engine tests for defend bonus clearing, flee success/failure, target switching, status effect tick/expiration, and stun turn skipping.
- Write game-controller-level tests or focused integration tests for victory rewards, defeat respawn/gold loss, and cleanup idempotency.

Simulator tests with random combat should use enough iterations for useful confidence but stable thresholds that are not brittle. Deterministic engine tests should cover rules that do not require probability.

### Frontend

Frontend changes should include:
- Store/unit tests where the current toolchain supports them, especially for combat snapshot updates.
- Build verification for `public/mud-client`.
- Browser smoke checks of desktop and mobile combat controls if a local dev server can be run.

## Documentation Updates

Because this changes user-facing combat behavior and message contracts:
- Update `PROJECT.md` for user-facing combat UX and behavior changes.
- Update `FEATURES.md` for structured combat snapshots, action semantics, and status/reward feedback.
- Update `ARCHITECTURE.md` if the WebSocket combat message contract or combat cleanup flow changes materially.
- Update `docs/player-guide/06-combat.md` and `docs/player-guide/11-mobile.md` for combat controls, target switching, defend, flee, status effects, rewards, and mobile usage.

## Non-Goals

- No party combat expansion beyond preserving existing multi-player instance behavior.
- No new combat classes, abilities, or itemization systems unless required to make existing skill-enabled balance reliable.
- No complete visual redesign of the game client.
- No migration away from the terminal-style combat log.

## Acceptance Criteria

- Combat balance tests run with Go 1.24+ and no longer miss `config/combat_balance.yaml`.
- Deferred combat balance failures are resolved under the skill-enabled solo baseline.
- Regression tests cover combat outcome rules: victory, defeat/respawn, reward/penalty, flee, defend, target switching, and status effects.
- Existing terminal combat messages still render.
- Structured combat snapshots are available to the frontend on combat lifecycle messages.
- Desktop combat UI exposes current enemies, target controls, HP/status state, queued action feedback, and rewards/penalties.
- Mobile combat UI exposes target controls and combat actions without typed commands for the normal path.
- Relevant project, feature, architecture, and player guide documentation is updated.
