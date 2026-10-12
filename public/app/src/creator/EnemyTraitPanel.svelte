<script>
  import EntitySelectButton from "./EntitySelectButton.svelte";
  import { scriptColumns, lootTableColumns, itemTemplateColumns } from "./tableColumns.js";

  /** The NPC being edited. Fields are mutated in place, then the store is nudged. */
  export let npc;
  export let store;
  export let scripts = [];
  export let lootTables = [];
  export let itemTemplates = [];
  /** { tiers, namedOverrides } from GET /api/balance/enemy-scaling */
  export let scaling = null;

  const knownDifficulties = [
    { id: "trivial", name: "Trivial" },
    { id: "easy", name: "Easy" },
    { id: "normal", name: "Normal" },
    { id: "hard", name: "Hard" },
    { id: "boss", name: "Boss" },
  ];

  let fleeInput = "";
  let lowInput = "";
  let seenKey = "";

  $: trait = npc?.enemyTrait;
  $: base = trait?.baseStats || null;
  $: factor = lookupFactors(npc, scaling);
  $: difficultyChoices = difficultyOptions(trait?.difficulty);
  $: lowLine = lowHealthLine(trait, npc?.maxHitPoints);

  $: {
    const key = `${npc?.id || ""}:${trait ? "on" : "off"}`;
    if (key !== seenKey) {
      seenKey = key;
      fleeInput = percentText(trait?.fleeThreshold);
      lowInput = percentText(trait?.lowHealthThreshold);
    }
    ensureShape(trait);
  }

  function ensureShape(current) {
    if (!current) return;
    if (!current.goldDrop) current.goldDrop = { min: 0, max: 0 };
    if (!Array.isArray(current.guaranteedLoot)) current.guaranteedLoot = [];
  }

  function difficultyOptions(current) {
    if (!current || knownDifficulties.some((d) => d.id === current)) return knownDifficulties;
    return [...knownDifficulties, { id: current, name: `${current} (unknown tier)` }];
  }

  function lookupFactors(current, table) {
    if (!current?.enemyTrait || !table) return null;
    const tier = current.enemyTrait.difficulty || "";
    const overrides = table.namedOverrides || {};
    const name = String(current.name || "");
    const key = Object.keys(overrides).find((entry) => entry.toLowerCase() === name.toLowerCase());
    if (key) {
      return { ...overrides[key], tier, known: true, named: true };
    }
    const tiers = table.tiers || {};
    if (tier && Object.prototype.hasOwnProperty.call(tiers, tier)) {
      return { ...tiers[tier], tier, known: true, named: false };
    }
    return { hp: 1, attack: 1, defense: 1, tier, known: false, named: false };
  }

  // Match encoding/json and Go's int32(float64) truncation, including the
  // minimums ApplyEnemyMultipliers applies only when a tier is known.
  function scaled(value, multiplier, kind) {
    const v = Math.trunc(Number(value) * Number(multiplier));
    if (kind === "defense") return v < 0 ? 0 : v;
    return v < 1 ? 1 : v;
  }

  function syncEffective() {
    const currentFactor = lookupFactors(npc, scaling);
    if (!npc?.enemyTrait?.baseStats || !currentFactor) return;
    const stats = npc.enemyTrait.baseStats;
    let hp;
    let atk;
    let def;
    if (!currentFactor.known) {
      hp = Number(stats.maxHitPoints) || 0;
      atk = Number(stats.attackPower) || 0;
      def = Number(stats.defense) || 0;
    } else {
      hp = scaled(stats.maxHitPoints, currentFactor.hp, "hp");
      atk = scaled(stats.attackPower, currentFactor.attack, "attack");
      def = scaled(stats.defense, currentFactor.defense, "defense");
    }
    store.update((state) => {
      const el = state.selectedElement;
      if (!el?.enemyTrait?.baseStats) return state;
      const oldMax = Number(el.maxHitPoints) || 0;
      const current = Number(el.currentHitPoints) || 0;
      const full = oldMax <= 0 || current >= oldMax;
      el.maxHitPoints = hp;
      el.enemyTrait.attackPower = atk;
      el.enemyTrait.defense = def;
      if (full || current > hp) el.currentHitPoints = hp;
      return state;
    });
  }

  function percentText(value) {
    const n = Number(value);
    if (!n) return "";
    const pct = Math.round(n * 1000) / 10;
    return String(pct);
  }

  function setFraction(field, raw) {
    const n = Number(raw);
    trait[field] = !raw || !Number.isFinite(n) || n <= 0 ? 0 : Math.min(1, n / 100);
  }

  function fmtFactor(value) {
    const n = Number(value);
    if (!Number.isFinite(n)) return "?";
    return String(Math.round(n * 1000) / 1000);
  }

  function lowHealthLine(current, maxHp) {
    const raw = Number(current?.lowHealthThreshold) || 0;
    const frac = raw > 0 ? Math.min(1, raw) : 0.3;
    const hp = Math.max(0, Math.floor((Number(maxHp) || 0) * frac));
    const pct = Math.round(frac * 1000) / 10;
    return `≤ ${pct}% (${hp} HP)`;
  }

  function setScript(field, id) {
    trait[field] = id || "";
  }

  function addGuaranteed() {
    trait.guaranteedLoot = [...(trait.guaranteedLoot || []), ""];
  }

  function setGuaranteed(index, id) {
    const next = [...trait.guaranteedLoot];
    next[index] = id || "";
    trait.guaranteedLoot = next;
  }

  function removeGuaranteed(index) {
    trait.guaranteedLoot = trait.guaranteedLoot.filter((_, i) => i !== index);
  }

  $: tierLabel = factor?.tier ? factor.tier : "unset";
</script>

{#if trait}
  <p class="text-xs text-slate-500 dark:text-slate-400">
    {#if base}
      These inputs edit the content base. Combat uses the effective stats, which are recalculated on save.
    {:else}
      No content base is stored. These inputs edit the stored combat stats, which combat uses as entered.
    {/if}
  </p>

  <section class="space-y-3">
    <h4 class="section-label">Combat stats</h4>
    <div class="grid grid-cols-2 md:grid-cols-3 gap-4">
      <div class="space-y-1.5">
        <label class="label-caps" for="enemy-creature-type">Creature Type</label>
        <select id="enemy-creature-type" class="input-base text-xs" bind:value={trait.creatureType}>
          <option value="beast">Beast</option>
          <option value="humanoid">Humanoid</option>
          <option value="undead">Undead</option>
          <option value="elemental">Elemental</option>
          <option value="construct">Construct</option>
          <option value="demon">Demon</option>
          <option value="dragon">Dragon</option>
          <option value="aberration">Aberration</option>
        </select>
      </div>
      <div class="space-y-1.5">
        <label class="label-caps" for="enemy-combat-style">Combat Style</label>
        <select id="enemy-combat-style" class="input-base text-xs" bind:value={trait.combatStyle}>
          <option value="melee">Melee</option>
          <option value="ranged">Ranged</option>
          <option value="magic">Magic</option>
          <option value="swarm">Swarm</option>
          <option value="brute">Brute</option>
          <option value="agile">Agile</option>
        </select>
        <p class="help">Swarm pulls other swarm copies in the room into the fight.</p>
      </div>
      <div class="space-y-1.5">
        <label class="label-caps" for="enemy-difficulty">Difficulty</label>
        <select id="enemy-difficulty" class="input-base text-xs" bind:value={trait.difficulty} on:change={syncEffective}>
          {#each difficultyChoices as d}
            <option value={d.id}>{d.name}</option>
          {/each}
        </select>
        <p class="help">Tier selects the combat multipliers. An unknown tier is stored as written.</p>
      </div>
    </div>

    {#if base && scaling && factor && !factor.known}
      <div class="rounded-md border border-amber-500/40 bg-amber-500/10 px-3 py-2 text-xs text-amber-200">
        unknown difficulty tier — no scaling applied
      </div>
    {/if}

    {#if base}
      <div class="grid grid-cols-3 gap-4">
        <div class="space-y-1.5">
          <label class="label-caps" for="enemy-base-hp">Content (base) HP</label>
          <input id="enemy-base-hp" class="input-base text-xs text-center" type="number" bind:value={base.maxHitPoints} on:input={syncEffective} />
        </div>
        <div class="space-y-1.5">
          <label class="label-caps" for="enemy-attack">Content (base) Attack</label>
          <input id="enemy-attack" class="input-base text-xs text-center" type="number" bind:value={base.attackPower} on:input={syncEffective} />
        </div>
        <div class="space-y-1.5">
          <label class="label-caps" for="enemy-defense">Content (base) Defense</label>
          <input id="enemy-defense" class="input-base text-xs text-center" type="number" bind:value={base.defense} on:input={syncEffective} />
        </div>
      </div>
      <div class="rounded-lg border border-slate-700/60 bg-slate-800/40 p-3 space-y-1">
        <p class="text-[10px] font-bold uppercase tracking-wider text-slate-400">
          Effective after {tierLabel} scaling (stored, used in combat)
        </p>
        <p class="text-sm text-slate-200 font-mono">
          HP {npc.maxHitPoints} · Attack {trait.attackPower} · Defense {trait.defense}
        </p>
        {#if factor}
          <p class="help">
            {#if factor.named}
              Named override replaces the {tierLabel} factors.
            {/if}
            HP ×{fmtFactor(factor.hp)}, Attack ×{fmtFactor(factor.attack)}, Defense ×{fmtFactor(factor.defense)}.
          </p>
        {:else}
          <p class="help">Multiplier table is still loading. The numbers above are the stored combat stats.</p>
        {/if}
      </div>
    {:else}
      <div class="grid grid-cols-3 gap-4">
        <div class="space-y-1.5">
          <label class="label-caps" for="enemy-stored-hp">Stored max HP</label>
          <input id="enemy-stored-hp" class="input-base text-xs text-center" type="number" bind:value={npc.maxHitPoints} />
          <p class="help">Also shown as max hit points above. Current HP is edited there.</p>
        </div>
        <div class="space-y-1.5">
          <label class="label-caps" for="enemy-attack">Attack Power</label>
          <input id="enemy-attack" class="input-base text-xs text-center" type="number" bind:value={trait.attackPower} />
          <p class="help">Stored combat attack. Unset content uses this number as entered.</p>
        </div>
        <div class="space-y-1.5">
          <label class="label-caps" for="enemy-defense">Defense</label>
          <input id="enemy-defense" class="input-base text-xs text-center" type="number" bind:value={trait.defense} />
          <p class="help">Stored combat defense. Engine default for a new trait here is 2.</p>
        </div>
      </div>
    {/if}

    <div class="grid grid-cols-2 md:grid-cols-3 gap-4">
      <div class="space-y-1.5">
        <label class="label-caps" for="enemy-xp">XP Reward</label>
        <input id="enemy-xp" class="input-base text-xs text-center" type="number" bind:value={trait.xpReward} />
        <p class="help">Experience granted on kill. Unset is 0.</p>
      </div>
      <div class="space-y-1.5">
        <label class="label-caps" for="enemy-attack-speed">Attack Speed</label>
        <input id="enemy-attack-speed" class="input-base text-xs text-center" type="number" step="0.1" min="0" bind:value={trait.attackSpeed} />
        <p class="help">Unset or 0: one swing per attack action (engine default). 1 is also one swing. 2 is two swings. Below 1 the enemy swings once, then holds. Positive values clamp to 0.25–3.</p>
      </div>
    </div>
  </section>

  <section class="space-y-3 pt-2 border-t border-slate-700/50">
    <h4 class="section-label">Behaviour</h4>
    <div class="grid grid-cols-2 md:grid-cols-3 gap-4">
      <div class="space-y-1.5">
        <label class="label-caps" for="enemy-aggro-radius">Aggro Radius</label>
        <input id="enemy-aggro-radius" class="input-base text-xs text-center" type="number" min="0" bind:value={trait.aggroRadius} />
        <p class="help">0 = unset (engine default). On-sight aggro is the checkbox and only covers this room. This radius is not a leash.</p>
      </div>
      <div class="space-y-1.5">
        <label class="label-caps" for="enemy-flee">Flee Threshold</label>
        <div class="flex items-center gap-2">
          <input
            id="enemy-flee"
            class="input-base text-xs text-center"
            type="number"
            min="0"
            max="100"
            step="1"
            placeholder="0"
            bind:value={fleeInput}
            on:input={() => setFraction("fleeThreshold", fleeInput)}
          />
          <span class="text-xs text-slate-500">%</span>
        </div>
        <p class="help">0 = never flees (engine default). Percent of max HP.</p>
      </div>
    </div>
    <div class="flex flex-col gap-2">
      <label class="flex items-center gap-2 text-xs cursor-pointer">
        <input type="checkbox" class="rounded border-slate-300 dark:border-slate-600" bind:checked={trait.aggroOnSight} />
        <span class="label-caps">Aggressive (attacks on sight)</span>
      </label>
      <p class="help">Default off. Starts a fight when a player enters this room, or when this NPC arrives in a room that already has players. The ruleset can disable it.</p>
      <label class="flex items-center gap-2 text-xs cursor-pointer">
        <input type="checkbox" class="rounded border-slate-300 dark:border-slate-600" bind:checked={trait.callForHelp} />
        <span class="label-caps">Call for help</span>
      </label>
      <p class="help">Default off. Pack pulls follow combat style swarm. This flag alone does not pull the room.</p>
    </div>
  </section>

  <section class="space-y-3 pt-2 border-t border-slate-700/50">
    <h4 class="section-label">Loot & rewards</h4>
    <div class="grid grid-cols-2 gap-4">
      <div class="space-y-1.5">
        <label class="label-caps" for="enemy-gold-min">Gold min</label>
        <input id="enemy-gold-min" class="input-base text-xs text-center" type="number" min="0" bind:value={trait.goldDrop.min} />
      </div>
      <div class="space-y-1.5">
        <label class="label-caps" for="enemy-gold-max">Gold max</label>
        <input id="enemy-gold-max" class="input-base text-xs text-center" type="number" min="0" bind:value={trait.goldDrop.max} />
      </div>
    </div>
    <p class="help">Inclusive gold range on kill. Unset is 0–0.</p>
    <div class="space-y-1.5">
      <div class="label-caps">Loot table</div>
      <EntitySelectButton
        value={trait.lootTableId || ""}
        elements={lootTables}
        columns={lootTableColumns}
        title="Select Loot Table"
        placeholder="None"
        on:change={(e) => trait.lootTableId = e.detail}
      />
      <p class="help">Rolled on kill. Shows the table name and id. None means no loot table.</p>
    </div>
    <div class="space-y-2">
      <div class="flex items-center justify-between">
        <span class="label-caps">Guaranteed loot</span>
        <button class="text-xs text-primary hover:underline" type="button" on:click={addGuaranteed}>+ Add item</button>
      </div>
      {#if trait.guaranteedLoot.length === 0}
        <p class="help">No guaranteed drops. These item templates always drop, in addition to the loot table.</p>
      {:else}
        {#each trait.guaranteedLoot as itemId, index}
          <div class="grid grid-cols-[1fr_auto] gap-2 items-center">
            <EntitySelectButton
              value={itemId}
              elements={itemTemplates}
              columns={itemTemplateColumns}
              title="Select Item Template"
              placeholder="Select item template..."
              on:change={(e) => setGuaranteed(index, e.detail)}
            />
            <button type="button" class="text-xs text-red-400 hover:text-red-300 px-2" on:click={() => removeGuaranteed(index)}>Remove</button>
          </div>
        {/each}
      {/if}
    </div>
    <div class="space-y-1.5">
      <label class="label-caps" for="enemy-max-drops">Max drops</label>
      <input id="enemy-max-drops" class="input-base text-xs text-center w-28" type="number" min="0" bind:value={trait.maxDrops} />
      <p class="help">0 = no limit on loot-table rolls (engine default). Guaranteed drops are not capped by this.</p>
    </div>
  </section>

  <section class="space-y-3 pt-2 border-t border-slate-700/50">
    <h4 class="section-label">Lua hooks</h4>
    <div class="space-y-1.5">
      <div class="label-caps">On aggro</div>
      <EntitySelectButton
        value={trait.onAggroScript || ""}
        elements={scripts}
        columns={scriptColumns}
        title="Select On Aggro Script"
        placeholder="None"
        on:change={(e) => setScript("onAggroScript", e.detail)}
      />
      <p class="help">None skips the hook. Runs once when this NPC enters a fight.</p>
    </div>
    <div class="space-y-1.5">
      <div class="label-caps">On death</div>
      <EntitySelectButton
        value={trait.onDeathScript || ""}
        elements={scripts}
        columns={scriptColumns}
        title="Select On Death Script"
        placeholder="None"
        on:change={(e) => setScript("onDeathScript", e.detail)}
      />
      <p class="help">None skips the hook. Runs once on death, before loot and XP.</p>
    </div>
    <div class="space-y-1.5">
      <div class="label-caps">On flee</div>
      <EntitySelectButton
        value={trait.onFleeScript || ""}
        elements={scripts}
        columns={scriptColumns}
        title="Select On Flee Script"
        placeholder="None"
        on:change={(e) => setScript("onFleeScript", e.detail)}
      />
      <p class="help">None skips the hook. Runs once when this NPC first chooses to flee.</p>
    </div>
    <div class="space-y-1.5">
      <div class="label-caps">On low health</div>
      <EntitySelectButton
        value={trait.onLowHealthScript || ""}
        elements={scripts}
        columns={scriptColumns}
        title="Select On Low Health Script"
        placeholder="None"
        on:change={(e) => setScript("onLowHealthScript", e.detail)}
      />
      <p class="help">None skips the hook. Runs once when HP first drops below the line and the NPC is still alive. A killing blow from above the line does not run it.</p>
    </div>
    <div class="space-y-1.5">
      <label class="label-caps" for="enemy-low-health">Low health threshold</label>
      <div class="flex items-center gap-3">
        <input
          id="enemy-low-health"
          class="input-base text-xs text-center w-24"
          type="number"
          min="0"
          max="99"
          step="1"
          placeholder="30"
          bind:value={lowInput}
          on:input={() => setFraction("lowHealthThreshold", lowInput)}
        />
        <span class="text-xs text-slate-400">{lowLine}</span>
      </div>
      <p class="help">Unset or 0 uses 30% (engine default). The HP figure uses the effective max.</p>
    </div>
  </section>
{/if}

<style>
  .section-label {
    font-size: 10px;
    font-weight: 700;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: #94a3b8;
  }

  .help {
    font-size: 10px;
    line-height: 1.35;
    color: #64748b;
  }
</style>
