# Character Templates System

This document describes the character template system in TalesMUD, which provides pre-built character archetypes for streamlined character creation.

## Overview

Character templates allow game designers to create pre-configured character builds that players can select during character creation. Each template defines race, class, attributes, backstory, and starting equipment.

## Architecture

```
CharacterTemplate Entity (stored in DB)
       |
   CharacterTemplatesService (CRUD)
       |
   CharacterTemplatesHandler (REST API)
       |
   CharactersService.CreateFromTemplate()
       |
   Player Character Entity
```

## Entity Structure

Character templates are defined in `pkg/entities/characters/charactertemplates.go`:

```go
type CharacterTemplate struct {
    *entities.Entity    // ID, timestamps

    // Display / Lore
    Name        string  // Human-readable name (e.g., "Forest Ranger")
    Description string  // Brief description
    Backstory   string  // Optional lore text
    OriginArea  string  // Starting region (e.g., "Silverwood Forest")
    Archetype   string  // Category (warrior, rogue, mage, etc.)

    // Gameplay Base
    Race  Race          // Dwarf, Human, Elve
    Class Class         // Warrior, Ranger, Hunter, Rogue, Mage, Cleric

    Level            int32
    CurrentHitPoints int32
    MaxHitPoints     int32

    Attributes    Attributes     // STR, DEX, CON, INT, WIS, CHA
    StartingItems []StartingItem // Equipment given on creation

    // Meta
    Source  string    // "db" or "system"
    Created time.Time
    Updated time.Time
}
```

### StartingItem Structure

```go
type StartingItem struct {
    Slot             ItemSlot  // Where to equip (or "inventory")
    ItemTemplateID   string    // ID of the item template to instantiate
    ItemTemplateName string    // Display name (convenience for UI)
}
```

## Available Races

| Race | Description |
|------|-------------|
| `Dwarf` | Sturdy mountain folk |
| `Human` | Versatile and adaptable |
| `Elve` | Graceful forest dwellers |

## Available Classes

| Class | Armor Type | Combat Type |
|-------|------------|-------------|
| `Warrior` | Heavy | Melee |
| `Ranger` | Medium | Ranged |
| `Hunter` | Medium | Ranged |
| `Rogue` | Light | Melee |
| `Mage` | Cloth | Magic |
| `Cleric` | Medium | Magic |

## Attributes

Character templates define the 6-attribute system:

| Attribute | Abbreviation | Description |
|-----------|--------------|-------------|
| Strength | STR | Physical power, melee damage |
| Dexterity | DEX | Agility, ranged attacks, defense |
| Constitution | CON | Health, stamina, resilience |
| Intelligence | INT | Magic power, knowledge |
| Wisdom | WIS | Perception, willpower |
| Charisma | CHA | Social skills, leadership |

## Equipment Slots

Templates can specify starting items for any of these slots:

| Slot | Description |
|------|-------------|
| `head` | Helmets, hats |
| `chest` | Armor, robes |
| `legs` | Pants, leggings |
| `boots` | Footwear |
| `neck` | Necklaces, amulets |
| `ring1` | First ring slot |
| `ring2` | Second ring slot |
| `hands` | Gloves, gauntlets |
| `mainHand` | Primary weapon |
| `offHand` | Shield, off-hand weapon |
| `inventory` | Added to bag (not equipped) |

## REST API

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/character-templates` | List all templates |
| GET | `/api/character-templates/:id` | Get template by ID |
| POST | `/api/character-templates` | Create template |
| PUT | `/api/character-templates/:id` | Update template |
| DELETE | `/api/character-templates/:id` | Delete template |

### Example Response

```json
{
  "id": "template-123",
  "name": "Forest Ranger",
  "description": "A skilled archer who protects the wild places.",
  "backstory": "Trained by the elven wardens of Silverwood...",
  "originArea": "Silverwood Forest",
  "archetype": "ranger",
  "race": "elve",
  "class": "ranger",
  "level": 1,
  "currentHitPoints": 18,
  "maxHitPoints": 18,
  "attributes": {
    "strength": 12,
    "dexterity": 16,
    "constitution": 14,
    "intelligence": 10,
    "wisdom": 14,
    "charisma": 10
  },
  "startingItems": [
    {
      "slot": "mainHand",
      "itemTemplateId": "item-longbow-basic",
      "itemTemplateName": "Hunting Bow"
    },
    {
      "slot": "chest",
      "itemTemplateId": "item-leather-armor",
      "itemTemplateName": "Leather Armor"
    },
    {
      "slot": "inventory",
      "itemTemplateId": "item-health-potion",
      "itemTemplateName": "Minor Health Potion"
    }
  ],
  "source": "db"
}
```

## Character Creation Flow

1. Player requests available templates via `/api/character-templates`
2. Player selects a template
3. Player can customize name and other allowed fields
4. Backend creates character from template:
   - Copies race, class, attributes, HP from template
   - For each starting item:
     - Creates item instance from item template
     - Equips to specified slot (or adds to inventory)
5. Character is persisted and ready to play

## System vs Custom Templates

Templates have a `source` field:

| Source | Description |
|--------|-------------|
| `system` | Built-in templates that ship with the game |
| `db` | Custom templates created by world builders |

System templates may be loaded from preset files on server startup.

## Design Guidelines

### Attribute Distribution

For balanced templates, consider these guidelines:

- **Total points**: ~75-80 for level 1
- **Primary stat**: 15-17
- **Secondary stats**: 12-14
- **Dump stats**: 8-10

### Example Builds

**Warrior** (STR-focused):
```
STR: 16, DEX: 12, CON: 14, INT: 8, WIS: 10, CHA: 10
```

**Mage** (INT-focused):
```
STR: 8, DEX: 12, CON: 12, INT: 16, WIS: 14, CHA: 10
```

**Rogue** (DEX-focused):
```
STR: 10, DEX: 16, CON: 12, INT: 12, WIS: 10, CHA: 12
```

### Starting Equipment

- Give at least a basic weapon appropriate to the class
- Include one piece of armor
- Consider 1-2 consumables (health potions, etc.)
- Avoid overpowering new characters

## Files

| File | Purpose |
|------|---------|
| `pkg/entities/characters/charactertemplates.go` | Entity definition |
| `pkg/entities/characters/charactertemplates_presets.go` | System presets |
| `pkg/repository/charactertemplates.go` | Repository interface |
| `pkg/repository/charactertemplates_sqlite.go` | SQLite implementation |
| `pkg/server/handler/charactertemplates.go` | REST handler |

## Future Enhancements

- Template categories/filtering in UI
- Template previews with rendered character art
- Racial/class ability preview
- Template recommendations based on playstyle quiz
- Template popularity tracking
