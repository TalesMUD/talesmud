# Product Requirements Document: Itemization System

**Version:** 1.0
**Date:** 2026-01-22
**Status:** Draft

---

## Executive Summary

This document outlines the comprehensive itemization system for TalesMUD, covering item management, player inventory, NPC loot mechanics, merchant trading, and room-based item interactions. The goal is to create a rich, interactive item economy that enhances gameplay through meaningful loot, trading, and inventory management.

---

## Table of Contents

1. [Current State Analysis](#current-state-analysis)
2. [Requirements Overview](#requirements-overview)
3. [Feature Specifications](#feature-specifications)
   - [Loot Tables](#1-loot-tables)
   - [Merchant System](#2-merchant-system)
   - [Room Items](#3-room-items)
   - [Player Inventory Commands](#4-player-inventory-commands)
   - [Equipment System](#5-equipment-system)
4. [Data Model Changes](#data-model-changes)
5. [Command Reference](#command-reference)
6. [Implementation Phases](#implementation-phases)

---

## Current State Analysis

### What's Implemented

#### Item Entity Model ✅
**File:** `pkg/entities/items/items.go`

The item system has a robust entity model:

| Field | Type | Description |
|-------|------|-------------|
| `Name` | string | Display name |
| `Description` | string | Item description |
| `Type` | ItemType | weapon, armor, consumable, currency, quest, collectible, crafting_material |
| `SubType` | ItemSubType | sword, bow, staff, chest, head, etc. |
| `Slot` | ItemSlot | Equipment location (13 slots) |
| `Quality` | ItemQuality | normal, magic, rare, legendary, mythic |
| `Level` | int32 | Item level (1-50) |
| `Properties` | map | Custom properties |
| `Attributes` | map | Stats (damage, armor, etc.) |
| `NoPickup` | bool | Cannot be picked up |
| `Tags` | []string | Item tags |

**Container Support:**
- `Closed`, `Locked`, `LockedBy` - Container state
- `Items` - Nested items for containers
- `MaxItems` - Container capacity

#### Template/Instance Pattern ✅
**File:** `pkg/service/items.go`

- Templates (`IsTemplate=true`) serve as blueprints
- Instances (`IsTemplate=false`) are unique game objects
- `InstanceSuffix` - 8-character UUID for targeting
- `CreateInstanceFromTemplate()` - Deep copy mechanism

#### Player Inventory Structure ✅
**File:** `pkg/entities/characters/character.go`

```go
type Character struct {
    Inventory     items.Inventory                // Array of items
    EquippedItems map[items.ItemSlot]*items.Item // Equipment slots
}
```

#### Enemy Trait (Partial) ⚠️
**File:** `pkg/entities/npcs/enemy_trait.go`

- `GoldDrop` - Min/max gold range
- `LootTableID` - Reference field (not implemented)
- `XPReward` - Experience on kill

#### Merchant Trait (Stub) ⚠️
**File:** `pkg/entities/npcs/merchant_trait.go`

```go
type MerchantTrait struct {
    MerchantType string // Only field, minimal implementation
}
```

#### Room Items Structure ✅
**File:** `pkg/entities/rooms/rooms.go`

- `Items []string` - Item IDs in room (exists but unused)

#### Existing Commands ⚠️
| Command | Status | Notes |
|---------|--------|-------|
| `inventory` / `i` | Partial | Lists names only, no grid view |
| `look` | Partial | TODO comment for room items |

#### Web UI & API ✅
- Full CRUD for items via REST API
- ItemsEditor.svelte for item management
- Template/instance distinction in UI

---

### What's Missing

| Feature | Priority | Complexity |
|---------|----------|------------|
| Loot Tables | High | Medium |
| Loot Drop Mechanics | High | Medium |
| Merchant Inventory | High | Medium |
| Buy/Sell Commands | High | Low |
| Pickup/Drop Commands | High | Low |
| Room Item Display | High | Low |
| Equip/Unequip Commands | Medium | Low |
| Inventory Capacity Enforcement | Medium | Low |
| Item Use (Consumables) | Medium | Medium |
| Container Commands | Low | Medium |
| Trading Between Players | Low | High |

---

## Requirements Overview

### Goals

1. **Rich Loot Economy** - NPCs drop meaningful, configurable loot
2. **Merchant Trading** - Buy/sell items with dynamic inventories
3. **World Interaction** - Items exist in rooms, can be picked up/dropped
4. **Inventory Management** - Clear, usable inventory system
5. **Scriptable** - Lua hooks for custom item behaviors

### Non-Goals (v1)

- Crafting system
- Item durability/degradation
- Auction house
- Mail/storage systems
- Item enchanting/socketing

---

## Feature Specifications

### 1. Loot Tables

#### Overview
Loot tables define what items NPCs can drop, with configurable drop rates and quantities.

#### Data Model

```go
// LootTable defines possible drops for an NPC or container
type LootTable struct {
    *entities.Entity
    Name        string      `json:"name"`
    Description string      `json:"description"`
    Entries     []LootEntry `json:"entries"`

    // Global modifiers
    GoldMultiplier float64 `json:"goldMultiplier"` // Applies to all gold drops
    DropBonus      float64 `json:"dropBonus"`      // Added to all drop chances
}

// LootEntry represents a single possible drop
type LootEntry struct {
    ItemTemplateID string  `json:"itemTemplateId"` // Reference to item template
    DropChance     float64 `json:"dropChance"`     // 0.0 - 1.0 (0% - 100%)
    MinQuantity    int32   `json:"minQuantity"`    // Minimum drop count
    MaxQuantity    int32   `json:"maxQuantity"`    // Maximum drop count
    Guaranteed     bool    `json:"guaranteed"`     // Always drops (ignores chance)

    // Optional conditions
    MinPlayerLevel int32    `json:"minPlayerLevel,omitempty"` // Requires player level
    RequiredTags   []string `json:"requiredTags,omitempty"`   // Requires player/quest tags
}
```

#### Enemy Integration

Update `EnemyTrait` in `pkg/entities/npcs/enemy_trait.go`:

```go
type EnemyTrait struct {
    // ... existing fields ...

    // Loot Configuration
    LootTableID    string  `json:"lootTableId,omitempty"`    // Primary loot table
    GuaranteedLoot []string `json:"guaranteedLoot,omitempty"` // Always-drop item template IDs
    MaxDrops       int32   `json:"maxDrops"`                 // Max items from table (0 = unlimited)
}
```

#### Drop Mechanics

When an enemy is defeated:

1. Roll for gold using `GoldDrop.Min` to `GoldDrop.Max`
2. Add `GuaranteedLoot` items (always drop)
3. Process `LootTable`:
   - For each entry, roll against `DropChance`
   - Check `MinPlayerLevel` and `RequiredTags` conditions
   - Generate quantity between `MinQuantity` and `MaxQuantity`
   - Stop if `MaxDrops` reached
4. Create item instances from templates
5. Place items in the room or player inventory (configurable)

#### Loot Events

```
Event: npc.death
Data: {
    npc: NPC,
    killer: Character,
    room: Room,
    droppedItems: []Item,
    droppedGold: int64
}
```

---

### 2. Merchant System

#### Overview
NPCs with `MerchantTrait` can buy/sell items. Merchants have inventory with optional stock limits.

#### Data Model

Update `MerchantTrait` in `pkg/entities/npcs/merchant_trait.go`:

```go
type MerchantTrait struct {
    MerchantType  string           `json:"merchantType"`   // general, weaponsmith, armorer, alchemist, etc.

    // Inventory
    Inventory     []MerchantItem   `json:"inventory"`      // Items for sale
    RestockTime   time.Duration    `json:"restockTime"`    // How often stock refreshes (0 = never)
    LastRestock   time.Time        `json:"lastRestock"`    // Last restock timestamp

    // Pricing
    BuyMultiplier  float64 `json:"buyMultiplier"`  // Multiplier when buying FROM player (e.g., 0.5 = 50%)
    SellMultiplier float64 `json:"sellMultiplier"` // Multiplier when selling TO player (e.g., 1.0 = 100%)

    // Restrictions
    AcceptedTypes  []ItemType `json:"acceptedTypes,omitempty"`  // Item types merchant will buy
    RejectedTags   []string   `json:"rejectedTags,omitempty"`   // Won't buy items with these tags
}

type MerchantItem struct {
    ItemTemplateID string `json:"itemTemplateId"` // Reference to item template

    // Pricing
    BasePrice      int64  `json:"basePrice"`      // Base price in gold (0 = use item default)
    PriceOverride  int64  `json:"priceOverride"`  // If set, ignores multipliers

    // Stock
    Quantity       int32  `json:"quantity"`       // Current stock (-1 = unlimited)
    MaxQuantity    int32  `json:"maxQuantity"`    // Stock cap on restock (-1 = unlimited)

    // Availability
    RequiredLevel  int32    `json:"requiredLevel,omitempty"`  // Player level to see/buy
    RequiredTags   []string `json:"requiredTags,omitempty"`   // Quest/reputation requirements
}
```

#### Price Calculation

**Selling to Player:**
```
finalPrice = (item.BasePrice OR merchantItem.BasePrice) * merchant.SellMultiplier
```

**Buying from Player:**
```
finalPrice = item.BasePrice * merchant.BuyMultiplier
```

If `PriceOverride` is set, it's used directly without multipliers.

#### Item Base Price

Add to `Item` struct:
```go
BasePrice int64 `json:"basePrice"` // Default price in gold
```

#### Merchant Commands

| Command | Syntax | Description |
|---------|--------|-------------|
| `list` | `list [category]` | Show merchant inventory |
| `buy` | `buy <item> [quantity]` | Purchase item |
| `sell` | `sell <item> [quantity]` | Sell item to merchant |
| `value` | `value <item>` | Check sell price |

#### Merchant Events

```
Event: merchant.buy    // Player buys from merchant
Event: merchant.sell   // Player sells to merchant
Data: {
    merchant: NPC,
    player: Character,
    item: Item,
    quantity: int,
    price: int64
}
```

---

### 3. Room Items

#### Overview
Items can exist on the ground in rooms. Players can see, pick up, and interact with them.

#### Data Model

Room already has `Items []string`. No changes needed to the room structure.

#### Display in Look Command

Update `look` command to show room items:

```
[Room Name]
[Room Description]

Exits: north, south, east

NPCs:
  - Guard Captain (hostile)

Items on the ground:
  - Rusty Sword
  - Gold Coins (x15)
  - Leather Pouch
```

#### Item Stacking

For display purposes, stackable items (currency, consumables) show quantity:
- `Gold Coins (x15)`
- `Health Potion (x3)`

Add to `Item`:
```go
Stackable bool  `json:"stackable"` // Can stack in inventory/room
Quantity  int32 `json:"quantity"`  // Current stack count
MaxStack  int32 `json:"maxStack"`  // Max stack size (0 = unlimited)
```

---

### 4. Player Inventory Commands

#### Inventory Display

**List View (default):**
```
=== Inventory (12/20 slots) ===
Weapons:
  [1] Rusty Sword (equipped)
  [2] Iron Dagger

Armor:
  [3] Leather Armor (equipped)
  [4] Cloth Hood

Consumables:
  [5] Health Potion (x5)
  [6] Mana Potion (x2)

Other:
  [7] Quest Letter
  [8] Gold Coins (x150)
```

**Grid View:**
```
=== Inventory (12/20 slots) ===
┌─────────────────┬─────────────────┐
│ [1] Rusty Sword │ [2] Iron Dagger │
├─────────────────┼─────────────────┤
│ [3] Leather Arm │ [4] Cloth Hood  │
├─────────────────┼─────────────────┤
│ [5] Health Pot  │ [6] Mana Potion │
├─────────────────┼─────────────────┤
│ [7] Quest Letter│ [8] Gold (x150) │
└─────────────────┴─────────────────┘
```

#### Command Reference

| Command | Aliases | Syntax | Description |
|---------|---------|--------|-------------|
| `inventory` | `i`, `inv` | `inventory [grid]` | Show inventory |
| `pickup` | `take`, `get` | `pickup <item>` | Pick up item from room |
| `drop` | | `drop <item>` | Drop item in room |
| `examine` | `look at`, `inspect` | `examine <item>` | Detailed item info |
| `use` | `consume` | `use <item>` | Use consumable item |
| `give` | | `give <item> to <player>` | Give item to another player |

---

### 5. Equipment System

#### Equipment Slots

```
┌─────────────────────────────────┐
│           [Head]                │
│                                 │
│  [Neck]                         │
│                                 │
│        [Chest]                  │
│                                 │
│ [Main Hand]    [Off Hand]       │
│                                 │
│        [Hands]                  │
│                                 │
│        [Legs]                   │
│                                 │
│ [Ring 1]       [Ring 2]         │
│                                 │
│        [Boots]                  │
└─────────────────────────────────┘
```

#### Equipment Commands

| Command | Syntax | Description |
|---------|--------|-------------|
| `equip` | `equip <item>` | Equip item to appropriate slot |
| `unequip` | `unequip <slot>` | Remove item from slot |
| `equipment` | `eq`, `gear` | Show equipped items |

#### Equip Display

```
=== Equipment ===
Head:      Empty
Neck:      Copper Amulet
Chest:     Leather Armor [8 armor, +1 agility]
Hands:     Empty
Legs:      Cloth Pants [2 armor]
Boots:     Empty
Main Hand: Rusty Sword [5 damage, 1.2 speed]
Off Hand:  Wooden Shield [15 block]
Ring 1:    Empty
Ring 2:    Ring of Minor Health [+10 HP]
```

---

## Data Model Changes

### New Entities

| Entity | File | Description |
|--------|------|-------------|
| `LootTable` | `pkg/entities/items/loottable.go` | Loot drop configuration |

### Modified Entities

| Entity | Changes |
|--------|---------|
| `Item` | Add `BasePrice`, `Stackable`, `Quantity`, `MaxStack` |
| `EnemyTrait` | Add `GuaranteedLoot`, `MaxDrops` |
| `MerchantTrait` | Full rewrite with inventory, pricing, restrictions |

### New Repositories

| Repository | Description |
|------------|-------------|
| `LootTablesRepository` | CRUD for loot tables |

### Database Schema

```sql
-- New table
CREATE TABLE loot_tables (
    id TEXT PRIMARY KEY,
    data JSON NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Index for quick lookup
CREATE INDEX idx_loot_tables_name ON loot_tables((json_extract(data, '$.name')));
```

---

## Command Reference

### Complete Command List

#### Inventory Commands
| Command | Aliases | Syntax | Priority |
|---------|---------|--------|----------|
| `inventory` | `i`, `inv` | `inventory [grid]` | High |
| `pickup` | `take`, `get` | `pickup <item>` | High |
| `drop` | - | `drop <item> [quantity]` | High |
| `examine` | `inspect` | `examine <item>` | High |
| `use` | `consume` | `use <item>` | Medium |
| `give` | - | `give <item> to <player>` | Low |

#### Equipment Commands
| Command | Aliases | Syntax | Priority |
|---------|---------|--------|----------|
| `equip` | `wear` | `equip <item>` | Medium |
| `unequip` | `remove` | `unequip <slot\|item>` | Medium |
| `equipment` | `eq`, `gear` | `equipment` | Medium |

#### Trading Commands
| Command | Aliases | Syntax | Priority |
|---------|---------|--------|----------|
| `list` | `shop` | `list [category]` | High |
| `buy` | - | `buy <item> [quantity]` | High |
| `sell` | - | `sell <item> [quantity]` | High |
| `value` | `price` | `value <item>` | Medium |

#### Interaction Commands
| Command | Aliases | Syntax | Priority |
|---------|---------|--------|----------|
| `admire` | - | `admire <item>` | Low |
| `open` | - | `open <container>` | Low |
| `close` | - | `close <container>` | Low |

---

## Implementation Phases

### Phase 1: Core Item Interactions (High Priority)
**Estimated Scope:** ~15 files

1. **Pickup/Drop Commands**
   - Implement `pickup` command
   - Implement `drop` command
   - Add items to room display in `look`
   - Handle item instance creation/cleanup

2. **Enhanced Inventory Display**
   - Grid view option
   - Group by category
   - Show quantities for stackable items
   - Slot numbering for quick reference

3. **Item Stacking**
   - Add `Stackable`, `Quantity`, `MaxStack` to Item
   - Merge logic for inventory
   - Split stack command

### Phase 2: Loot System (High Priority)
**Estimated Scope:** ~10 files

1. **Loot Table Entity**
   - Create LootTable entity
   - Repository and service layer
   - Web UI editor for loot tables

2. **Drop Mechanics**
   - Integrate with NPC death
   - Roll drop chances
   - Create item instances
   - Place in room

3. **Enemy Trait Enhancement**
   - Add loot configuration fields
   - Guaranteed loot support

### Phase 3: Merchant System (High Priority)
**Estimated Scope:** ~12 files

1. **Merchant Trait Rewrite**
   - Full inventory support
   - Pricing configuration
   - Stock management

2. **Trading Commands**
   - `list` - Show merchant inventory
   - `buy` - Purchase items
   - `sell` - Sell items
   - `value` - Check prices

3. **Web UI**
   - Merchant inventory editor
   - Pricing configuration

### Phase 4: Equipment System (Medium Priority)
**Estimated Scope:** ~8 files

1. **Equipment Commands**
   - `equip` command
   - `unequip` command
   - `equipment` display

2. **Slot Validation**
   - Check item slot compatibility
   - Handle two-handed weapons
   - Ring slot management

### Phase 5: Advanced Features (Lower Priority)
**Estimated Scope:** ~10 files

1. **Consumable Use**
   - `use` command
   - Effect application
   - Quantity decrement

2. **Container Support**
   - `open`/`close` commands
   - Nested item access
   - Lock/unlock mechanics

3. **Lua Integration**
   - Item pickup/drop events
   - Custom item use scripts
   - Loot modification hooks

---

## Lua Scripting Events

### New Events

| Event | Trigger | Data |
|-------|---------|------|
| `item.pickup` | Player picks up item | `{player, item, room}` |
| `item.drop` | Player drops item | `{player, item, room}` |
| `item.use` | Player uses item | `{player, item, target?}` |
| `item.equip` | Player equips item | `{player, item, slot}` |
| `item.unequip` | Player unequips item | `{player, item, slot}` |
| `npc.loot` | NPC generates loot | `{npc, player, items, gold}` |
| `merchant.buy` | Player buys from merchant | `{merchant, player, item, quantity, price}` |
| `merchant.sell` | Player sells to merchant | `{merchant, player, item, quantity, price}` |

### New Lua Functions

```lua
-- Inventory management
tales.items.addToInventory(characterId, itemId)
tales.items.removeFromInventory(characterId, itemId)
tales.items.getInventory(characterId)

-- Room items
tales.items.addToRoom(roomId, itemId)
tales.items.removeFromRoom(roomId, itemId)
tales.items.getRoomItems(roomId)

-- Loot tables
tales.loot.roll(lootTableId, playerLevel)
tales.loot.addDrop(lootTableId, itemTemplateId, chance)

-- Merchant
tales.merchant.getInventory(npcId)
tales.merchant.setPrice(npcId, itemTemplateId, price)
tales.merchant.restock(npcId)
```

---

## Web UI Requirements

### New Editors

1. **Loot Table Editor**
   - Create/edit loot tables
   - Drag-drop item templates
   - Drop chance sliders
   - Quantity range inputs
   - Preview loot rolls

2. **Merchant Inventory Editor** (in NPC Editor)
   - Add items to merchant inventory
   - Set prices and stock limits
   - Configure restock settings
   - Set buy/sell multipliers

### Enhanced Editors

1. **Item Editor**
   - Base price field
   - Stackable toggle
   - Max stack input

2. **NPC Editor**
   - Loot table dropdown (for enemies)
   - Merchant inventory section

---

## Success Metrics

1. **Loot Drops** - NPCs successfully drop configured loot on death
2. **Trading** - Players can buy/sell items with merchants
3. **Room Items** - Items visible in rooms, pickupable
4. **Inventory** - Clear, usable inventory display
5. **Equipment** - Players can equip/unequip items

---

## Appendix: File Reference

### Current Implementation Files

| Purpose | File |
|---------|------|
| Item Entity | `pkg/entities/items/items.go` |
| Inventory | `pkg/entities/items/inventory.go` |
| Repository | `pkg/repository/items_sqlite.go` |
| Service | `pkg/service/items.go` |
| Enemy Trait | `pkg/entities/npcs/enemy_trait.go` |
| Merchant Trait | `pkg/entities/npcs/merchant_trait.go` |
| Inventory Command | `pkg/mudserver/game/commands/inventory.go` |
| Look Command | `pkg/mudserver/game/commands/look.go` |
| HTTP Handler | `pkg/server/handler/items.go` |
| Lua Module | `pkg/scripts/runner/lua/modules/items.go` |
| UI Editor | `public/app/src/creator/ItemsEditor.svelte` |
| Room Entity | `pkg/entities/rooms/rooms.go` |
| Character | `pkg/entities/characters/character.go` |

### Files to Create

| Purpose | Proposed File |
|---------|---------------|
| Loot Table Entity | `pkg/entities/items/loottable.go` |
| Loot Table Repository | `pkg/repository/loottables_sqlite.go` |
| Loot Table Service | `pkg/service/loottables.go` |
| Loot Table Handler | `pkg/server/handler/loottables.go` |
| Pickup Command | `pkg/mudserver/game/commands/pickup.go` |
| Drop Command | `pkg/mudserver/game/commands/drop.go` |
| Equip Command | `pkg/mudserver/game/commands/equip.go` |
| Trade Commands | `pkg/mudserver/game/commands/trade.go` |
| Use Command | `pkg/mudserver/game/commands/use.go` |
| Loot Table UI | `public/app/src/creator/LootTableEditor.svelte` |
| Loot Table API | `public/app/src/api/loottables.js` |
