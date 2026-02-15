# World Builder Agent - Product Requirements Document

## Overview

A generative AI system for creating rich, interconnected MUD content including rooms, NPCs, items, enemies, quests, and discoverable secrets. The system leverages the existing TalesMUD REST API through an agent-based architecture.

## Goals

1. **Automate bulk content creation** - Generate entire areas (20-100+ rooms) with consistent themes
2. **Maintain narrative coherence** - Ensure storylines, NPC dialogue, and world details align
3. **Preserve game balance** - Respect economy, combat difficulty, and progression systems
4. **Enable iterative refinement** - Allow human review and AI-assisted editing
5. **Support multiple content types** - Rooms, NPCs, items, enemies, spawners, secrets, and eventually Lua scripts

## Content Types

| Type | Description | Complexity |
|------|-------------|------------|
| **Rooms** | Locations with descriptions, exits, and properties | Medium |
| **NPCs** | Non-hostile characters with dialogue trees | High |
| **Enemies** | Hostile entities with stats, loot tables, abilities | Medium |
| **Items** | Equipment, consumables, quest items, keys | Low-Medium |
| **Spawners** | Rules for enemy/NPC placement and respawn | Low |
| **Secrets** | Hidden rooms, triggers, special actions | High |
| **Quests** | Multi-step objectives with rewards | High |
| **Scripts** | Lua behaviors for complex interactions | Very High |

---

## Option A: Single Agent with MCP Tools

### Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    User / Operator                       │
│         "Create a haunted forest with 30 rooms"          │
└─────────────────────────┬───────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│                   World Builder Agent                    │
│                                                          │
│  - Receives high-level prompt                            │
│  - Plans content structure                               │
│  - Calls MCP tools iteratively                           │
│  - Maintains context across calls                        │
│  - Self-validates results                                │
└─────────────────────────┬───────────────────────────────┘
                          │ MCP Protocol
                          ▼
┌─────────────────────────────────────────────────────────┐
│                MCP Tool Server                           │
│                                                          │
│  Tools:                                                  │
│  ├── create_area(name, theme, description)               │
│  ├── create_room(area_id, name, desc, exits, props)      │
│  ├── link_rooms(room_a, room_b, direction)               │
│  ├── create_npc(room_id, name, desc, dialogue)           │
│  ├── create_enemy(room_id, template, level, loot)        │
│  ├── create_item(name, type, stats, location)            │
│  ├── create_spawner(room_id, entity_type, config)        │
│  ├── create_secret(room_id, trigger, reveal)             │
│  ├── get_area_summary(area_id)                           │
│  ├── get_templates(type)  # NPC/enemy/item templates     │
│  ├── validate_area(area_id)                              │
│  └── rollback_area(area_id)                              │
└─────────────────────────┬───────────────────────────────┘
                          │ REST API
                          ▼
┌─────────────────────────────────────────────────────────┐
│                   TalesMUD Server                        │
│                                                          │
│  Existing API endpoints for CRUD operations              │
│  SQLite database persistence                             │
└─────────────────────────────────────────────────────────┘
```

### MCP Tool Definitions

```typescript
// Core creation tools
interface CreateAreaParams {
  name: string;
  theme: string;
  description: string;
  suggestedLevel?: { min: number; max: number };
}

interface CreateRoomParams {
  areaId: string;
  name: string;
  description: string;
  exits: { direction: string; targetRoomId?: string }[];
  properties?: {
    isIndoors?: boolean;
    lighting?: "dark" | "dim" | "normal" | "bright";
    terrain?: string;
  };
}

interface CreateNPCParams {
  roomId: string;
  name: string;
  description: string;
  dialogue: DialogueNode[];
  role?: "merchant" | "quest_giver" | "info" | "ambient";
  inventory?: string[];  // item template IDs
}

interface CreateEnemyParams {
  roomId: string;
  templateId: string;
  levelOverride?: number;
  lootTable?: LootEntry[];
  spawnerId?: string;
}

interface CreateSecretParams {
  roomId: string;
  triggerType: "examine" | "use_item" | "say_phrase" | "skill_check";
  triggerConfig: Record<string, unknown>;
  revealType: "hidden_exit" | "hidden_item" | "message" | "event";
  revealConfig: Record<string, unknown>;
}

// Context tools
interface GetAreaSummaryResult {
  rooms: { id: string; name: string; connections: string[] }[];
  npcs: { id: string; name: string; roomId: string }[];
  enemies: { id: string; template: string; roomId: string }[];
  items: { id: string; name: string; location: string }[];
  validationWarnings: string[];
}

// Validation tools
interface ValidateAreaResult {
  isValid: boolean;
  errors: string[];
  warnings: string[];
  suggestions: string[];
}
```

### Agent Prompt Structure

```markdown
# World Builder Agent

You are a MUD world builder for TalesMUD. Your role is to create rich,
interconnected game content that provides engaging player experiences.

## Your Capabilities
- Create areas with multiple connected rooms
- Design NPCs with meaningful dialogue
- Place enemies with appropriate difficulty
- Distribute loot and items fairly
- Hide secrets for players to discover

## Guidelines

### Room Design
- Each room needs a unique, evocative description (2-4 sentences)
- Descriptions should hint at exits and points of interest
- Vary room sizes and purposes (corridors, chambers, outdoor spaces)
- Include sensory details: sounds, smells, lighting

### NPC Design
- Give NPCs distinct personalities and speech patterns
- Dialogue should provide world lore, hints, or services
- Quest givers need clear objectives and rewards

### Enemy Placement
- Difficulty should increase as players go deeper
- Boss encounters need memorable descriptions
- Loot should match enemy type and difficulty

### Secrets
- Hide at least 1 secret per 10 rooms
- Secrets should reward exploration and cleverness
- Provide subtle hints in room descriptions

## Workflow
1. Analyze the request to understand theme, size, and purpose
2. Plan the layout (sketch room connections mentally)
3. Create the area container
4. Create rooms, starting from the entrance
5. Add NPCs and enemies
6. Place items and loot
7. Add secrets and special interactions
8. Validate the area
9. Report summary to user
```

### Advantages

- **Simplicity** - Single agent, straightforward debugging
- **Context efficiency** - One model maintains full area context
- **Lower latency** - No inter-agent communication overhead
- **Easier to start** - Can build incrementally

### Limitations

- **Prompt complexity** - Single prompt must cover all domains
- **Context limits** - Large areas may exceed context window
- **Style consistency** - Hard to maintain different "voices" for NPCs
- **No specialization** - Cannot use different models for different tasks

### When to Use

- Initial implementation / MVP
- Smaller areas (< 50 rooms)
- Rapid prototyping
- When simplicity is valued over capability

---

## Option B: Multi-Agent Architecture

### Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    User / Operator                       │
│         "Create a haunted forest with 30 rooms"          │
└─────────────────────────┬───────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│                  Director Agent                          │
│                                                          │
│  Responsibilities:                                       │
│  - Parse user intent                                     │
│  - Create area brief / design document                   │
│  - Orchestrate specialist agents                         │
│  - Review and approve outputs                            │
│  - Handle conflicts and revisions                        │
│  - Final validation and delivery                         │
└───────┬─────────┬─────────┬─────────┬─────────┬─────────┘
        │         │         │         │         │
        ▼         ▼         ▼         ▼         ▼
┌───────────┐ ┌───────────┐ ┌───────────┐ ┌───────────┐ ┌───────────┐
│Cartograph-│ │   NPC     │ │  Enemy    │ │   Loot    │ │  Secret   │
│   er      │ │ Designer  │ │ Designer  │ │  Master   │ │  Keeper   │
│           │ │           │ │           │ │           │ │           │
│ - Room    │ │ - Chars   │ │ - Mobs    │ │ - Items   │ │ - Hidden  │
│   layout  │ │ - Dialog  │ │ - Combat  │ │ - Economy │ │   content │
│ - Geography│ │ - Quests │ │ - Balance │ │ - Rewards │ │ - Puzzles │
│ - Exits   │ │ - Lore    │ │ - Spawns  │ │ - Drops   │ │ - Triggers│
└─────┬─────┘ └─────┬─────┘ └─────┬─────┘ └─────┬─────┘ └─────┬─────┘
      │             │             │             │             │
      └─────────────┴─────────────┴─────────────┴─────────────┘
                                  │
                                  ▼
                    ┌─────────────────────────┐
                    │    Shared MCP Server    │
                    │    (Same as Option A)   │
                    └─────────────┬───────────┘
                                  │
                                  ▼
                    ┌─────────────────────────┐
                    │     TalesMUD Server     │
                    └─────────────────────────┘
```

### Agent Definitions

#### Director Agent

```markdown
# Director Agent

You are the creative director for TalesMUD world building. You orchestrate
a team of specialist agents to create cohesive game content.

## Your Responsibilities
1. Analyze user requests and create a design brief
2. Delegate tasks to specialist agents
3. Review outputs for consistency and quality
4. Resolve conflicts between agents
5. Ensure the final product meets requirements

## Design Brief Format
When creating an area, produce a brief like:

```yaml
area_name: "The Whispering Woods"
theme: "Haunted forest with ancient elven ruins"
level_range: [10, 15]
room_count: 30
key_locations:
  - Entrance clearing
  - Ruined watchtower
  - Sacred grove (boss)
  - Hidden elven tomb (secret)
story_hook: "Villagers report ghostly lights..."
atmosphere: "Eerie, melancholic, dangerous beauty"
special_requirements:
  - At least one friendly NPC (lost traveler)
  - Boss enemy: Corrupted treant
  - Quest: Lay the elven spirits to rest
```

## Orchestration Flow
1. Create design brief
2. Send to Cartographer: "Create layout per brief"
3. Send to NPC Designer: "Populate with characters"
4. Send to Enemy Designer: "Add threats and encounters"
5. Send to Loot Master: "Distribute rewards"
6. Send to Secret Keeper: "Add hidden content"
7. Review all outputs
8. Request revisions if needed
9. Validate and deliver
```

#### Cartographer Agent

```markdown
# Cartographer Agent

You design room layouts and geography for TalesMUD areas.

## Your Focus
- Logical spatial relationships
- Interesting navigation (not just linear)
- Varied room types and purposes
- Clear but evocative descriptions
- Proper exit connections

## Layout Patterns
- Hub and spoke: Central area with branches
- Linear with branches: Main path with side areas
- Maze: Complex interconnections
- Layered: Outdoor → building → dungeon
- Mixed: Combination of patterns

## Room Description Guidelines
- 2-4 sentences per room
- Include cardinal directions naturally
- Mention lighting, sounds, smells
- Hint at history or purpose
- Leave hooks for other agents (NPC spots, enemy areas)

## Output
Call create_room and link_rooms tools to build the layout.
Provide a room manifest to the Director when complete.
```

#### NPC Designer Agent

```markdown
# NPC Designer Agent

You create memorable non-player characters for TalesMUD.

## Your Focus
- Distinct personalities and voices
- Meaningful dialogue trees
- World-building through conversation
- Quest hooks and information
- Merchant inventories

## Character Archetypes
- The Reluctant Helper: Needs convincing
- The Eager Guide: Loves sharing knowledge
- The Mysterious Stranger: Cryptic hints
- The Merchant: Focused on trade
- The Quest Giver: Has problems to solve

## Dialogue Guidelines
- Match speech to character background
- Include greeting, main topics, farewell
- Provide 2-3 conversation branches
- Hide lore in optional dialogue
- Quest dialogue needs clear objectives

## Output
Call create_npc tool for each character.
Provide NPC manifest to Director when complete.
```

#### Enemy Designer Agent

```markdown
# Enemy Designer Agent

You create combat encounters and enemy placements for TalesMUD.

## Your Focus
- Appropriate difficulty scaling
- Thematic enemy selection
- Interesting combat variety
- Boss encounter design
- Spawner configuration

## Difficulty Guidelines
- Entrance: Easy enemies, singles or pairs
- Middle: Medium enemies, small groups
- Deep/hidden: Hard enemies, larger groups
- Boss: Challenging, unique mechanics

## Enemy Placement
- Guard patrol routes
- Ambush positions
- Lair locations
- Wandering monsters
- Boss arenas

## Output
Call create_enemy and create_spawner tools.
Provide encounter manifest to Director when complete.
```

#### Loot Master Agent

```markdown
# Loot Master Agent

You design item distribution and economy for TalesMUD areas.

## Your Focus
- Balanced reward pacing
- Thematic item selection
- Risk/reward alignment
- Economy preservation
- Unique item placement

## Loot Tiers
- Common: Consumables, basic gear
- Uncommon: Useful items, minor upgrades
- Rare: Significant upgrades, limited
- Epic: Boss drops only
- Unique: One per area, special

## Placement Guidelines
- Ground loot: Common items
- Container loot: Uncommon items
- Enemy drops: Scaled to difficulty
- Boss drops: Rare/Epic items
- Secret areas: Best rewards

## Output
Call create_item tool and configure loot tables.
Provide loot manifest to Director when complete.
```

#### Secret Keeper Agent

```markdown
# Secret Keeper Agent

You design hidden content and discoveries for TalesMUD areas.

## Your Focus
- Rewarding exploration
- Fair but challenging puzzles
- Hidden narrative content
- Secret areas and items
- Environmental storytelling

## Secret Types
- Hidden exits: Requires examination or item
- Hidden items: Concealed in descriptions
- Triggered events: Say phrase, use item
- Skill checks: Perception, knowledge
- Environmental: Move object, solve puzzle

## Design Guidelines
- Hint at secrets in room descriptions
- Reward matches difficulty of discovery
- Don't gate required content behind secrets
- Create "aha!" moments
- Layer secrets (secret leads to secret)

## Output
Call create_secret tool for each hidden element.
Provide secrets manifest to Director (marked spoiler).
```

### Framework Options

#### Mastra.ai

```typescript
import { Agent, Workflow } from '@mastra/core';

const director = new Agent({
  name: 'Director',
  model: 'claude-opus-4-5-20251101',
  instructions: directorPrompt,
  tools: [delegateToAgent, reviewOutput, mcpTools],
});

const cartographer = new Agent({
  name: 'Cartographer',
  model: 'claude-sonnet-4-20250514',  // Cheaper for structured output
  instructions: cartographerPrompt,
  tools: [createRoom, linkRooms],
});

const workflow = new Workflow({
  name: 'CreateArea',
  steps: [
    { agent: director, action: 'createBrief' },
    { agent: cartographer, action: 'createLayout', parallel: false },
    { agent: npcDesigner, action: 'populateNPCs', parallel: true },
    { agent: enemyDesigner, action: 'placeEnemies', parallel: true },
    { agent: lootMaster, action: 'distributeLoot', parallel: true },
    { agent: secretKeeper, action: 'hideSecrets', parallel: false },
    { agent: director, action: 'reviewAndFinalize' },
  ],
});
```

#### LangGraph Alternative

```python
from langgraph.graph import StateGraph, END

class AreaState(TypedDict):
    brief: DesignBrief
    rooms: List[Room]
    npcs: List[NPC]
    enemies: List[Enemy]
    items: List[Item]
    secrets: List[Secret]
    status: str

workflow = StateGraph(AreaState)
workflow.add_node("director_plan", director_plan)
workflow.add_node("create_layout", cartographer_create)
workflow.add_node("add_npcs", npc_designer_create)
workflow.add_node("add_enemies", enemy_designer_create)
workflow.add_node("add_loot", loot_master_create)
workflow.add_node("add_secrets", secret_keeper_create)
workflow.add_node("director_review", director_review)

workflow.add_edge("director_plan", "create_layout")
workflow.add_edge("create_layout", "add_npcs")
workflow.add_conditional_edges(
    "director_review",
    should_revise,
    {"revise": "create_layout", "approve": END}
)
```

### Advantages

- **Specialization** - Each agent masters its domain
- **Parallel execution** - NPCs, enemies, loot can run simultaneously
- **Better quality** - Focused prompts produce better results
- **Scalability** - Can add new specialist agents easily
- **Model flexibility** - Use expensive models for creative work, cheap for structured
- **Reviewable** - Director provides human-like oversight

### Limitations

- **Complexity** - More moving parts to debug
- **Latency** - Sequential steps add up
- **Coordination** - Agents may produce conflicting content
- **Cost** - More API calls
- **Context sharing** - Must explicitly pass information between agents

### When to Use

- Large areas (50+ rooms)
- High-quality content requirements
- When you need distinct NPC voices
- Complex narrative content
- Production use cases

---

## Implementation Roadmap

### Phase 1: Foundation (Option A - Single Agent)

1. **Create MCP Tool Server**
   - Wrap existing TalesMUD REST API
   - Add validation endpoints
   - Add rollback capability

2. **Build Single Agent**
   - Design comprehensive prompt
   - Test with small areas (5-10 rooms)
   - Iterate on prompt based on output quality

3. **Integration**
   - CLI tool for operators
   - Optional: MUD Creator UI integration

### Phase 2: Enhancement

4. **Add Context Tools**
   - get_existing_areas (for consistency)
   - get_templates (items, enemies, NPCs)
   - get_world_lore (setting information)

5. **Improve Validation**
   - Connectivity checks
   - Balance analysis
   - Narrative coherence scoring

### Phase 3: Multi-Agent (Option B)

6. **Extract Specialist Agents**
   - Start with Cartographer (most independent)
   - Add NPC Designer
   - Continue with others

7. **Build Orchestration**
   - Choose framework (Mastra, LangGraph, custom)
   - Implement Director agent
   - Add review/revision loops

8. **Optimize**
   - Parallel execution where possible
   - Caching for templates and context
   - Cost optimization (model selection)

### Phase 4: Advanced Features

9. **Script Generation**
   - Lua script templates
   - Agent that writes custom behaviors
   - Sandboxed testing

10. **Player Feedback Loop**
    - Track player engagement with generated content
    - Feed metrics back to improve generation

---

## Success Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| Room quality | 4/5 rating | Human review sample |
| Connectivity | 100% | Automated validation |
| Generation time | < 5 min for 30 rooms | Timing |
| Balance accuracy | Within 10% of manual | Difficulty analysis |
| Secret discovery rate | 30-50% of players | Analytics |
| Player engagement | Equal to manual content | Session metrics |

---

## Open Questions

1. **Consistency across areas** - How do we ensure generated areas fit with existing world lore?
2. **Player testing** - Should we A/B test generated vs manual content?
3. **Iteration workflow** - How do designers request changes to generated content?
4. **Version control** - How do we track and rollback generated content?
5. **Quality threshold** - What's the minimum acceptable quality for auto-publishing?

---

## Appendix: Technology Comparison

| Framework | Pros | Cons | Best For |
|-----------|------|------|----------|
| **Custom MCP** | Full control, simple | Build from scratch | Option A |
| **Mastra.ai** | TypeScript, good DX | Newer, less docs | Node.js stack |
| **LangGraph** | Mature, flexible | Python only | Complex workflows |
| **CrewAI** | Easy agent definition | Less control | Quick prototypes |
| **AutoGen** | Microsoft backing | Heavy, complex | Research |
