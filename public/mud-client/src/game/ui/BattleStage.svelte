<script>
  import { onDestroy } from 'svelte';
  import { hashedAvatar } from '../portraitSrc.js';
  import { skillDisplayName, isConsumableItem } from '../hudPrefs.js';

  export let store;
  export let sendMessage;

  let panel = null; // null | 'skills' | 'items'
  let nowMs = Date.now();
  let tickTimer = null;
  let fxKey = 0;

  $: phase = $store.combatPhase || ($store.inCombat ? 'active' : 'idle');
  $: visible = phase === 'active' || phase === 'ending';
  $: enemies = $store.combatEnemies || [];
  $: players = $store.combatPlayers || [];
  $: targetId = $store.combatTargetId;
  $: turn = $store.combatTurn;
  $: log = $store.combatLog || [];
  $: outcome = $store.combatOutcome;
  $: endMessage = $store.combatEndMessage || '';
  $: fx = $store.combatFx;
  $: character = $store.character;
  $: stats = $store.characterStats || {};
  $: inventory = $store.inventory || [];
  $: equippedSkills = stats.equippedSkills || character?.equippedSkills || [];
  $: selfId = character?.id || '';
  $: selfCombatant = players.find((p) => p.id === selfId) || players[0] || null;

  $: selfHp = selfCombatant?.hp ?? stats.currentHitPoints ?? 0;
  $: selfMaxHp = selfCombatant?.maxHp ?? stats.maxHitPoints ?? 1;
  $: selfMana = stats.currentMana ?? 0;
  $: selfMaxMana = stats.maxMana ?? 0;
  $: selfName = selfCombatant?.name || character?.name || 'You';
  $: selfClass = typeof character?.class === 'object'
    ? (character.class.name || '')
    : (character?.class || '');

  $: isMyTurn = !!(turn && selfId && turn.actorId === selfId);
  $: deadlineMs = isMyTurn ? (turn?.deadlineMs || 0) : 0;
  $: timerLeftMs = deadlineMs > 0 ? Math.max(0, deadlineMs - nowMs) : 0;
  $: timerPct = deadlineMs > 0
    ? Math.max(0, Math.min(100, (timerLeftMs / 10000) * 100))
    : 0;

  $: if (fx?.at) fxKey = fx.at;

  $: if (visible && deadlineMs > 0) startTick();
  else stopTick();

  $: consumables = inventory.filter(isConsumableItem);
  $: skillEntries = (equippedSkills || []).map((s) => {
    if (typeof s === 'string') return { id: s, name: skillDisplayName(s) };
    const id = s.id || s.skillId || '';
    const name = s.name || skillDisplayName(id);
    return { id, name };
  }).filter((s) => s.name);

  function startTick() {
    if (tickTimer) return;
    tickTimer = setInterval(() => { nowMs = Date.now(); }, 200);
  }
  function stopTick() {
    if (tickTimer) {
      clearInterval(tickTimer);
      tickTimer = null;
    }
  }

  onDestroy(stopTick);

  function combatantPortrait(c, fallbackKey) {
    const p = (c && c.portrait) || '';
    if (p) {
      if (p.startsWith('/') || p.startsWith('http') || p.startsWith('img/')) return p;
      return `/api/portraits/${p.replace(/\.png$/i, '')}.png`;
    }
    return hashedAvatar(fallbackKey || (c && c.name) || 'hero');
  }

  function hpPct(hp, maxHp) {
    const m = maxHp > 0 ? maxHp : 1;
    return Math.max(0, Math.min(100, (hp / m) * 100));
  }

  function hpColor(pct) {
    if (pct > 60) return '#ef4444'; // enemy bars stay red per mock
    if (pct > 30) return '#f59e0b';
    return '#dc2626';
  }

  function playerHpColor(pct) {
    if (pct > 60) return '#22c55e';
    if (pct > 30) return '#f59e0b';
    return '#ef4444';
  }

  function selectEnemy(enemy) {
    if (!enemy || phase !== 'active') return;
    if (store.setCombatTarget) store.setCombatTarget(enemy.id);
    const name = enemy.name || '';
    if (name && sendMessage) sendMessage(`attack ${name}`);
  }

  function cmd(text) {
    panel = null;
    if (sendMessage && text) sendMessage(text);
  }

  function doAttack() {
    const target = enemies.find((e) => e.id === targetId) || enemies.find((e) => (e.hp ?? 0) > 0);
    if (target?.name) cmd(`attack ${target.name}`);
    else cmd('attack');
  }

  function doDefend() { cmd('defend'); }
  function doFlee() { cmd('flee'); }

  function castSkill(skill) {
    if (!skill?.name) return;
    cmd(`cast ${skill.name}`);
  }

  function useItem(item) {
    if (!item?.name) return;
    cmd(`use ${item.name}`);
  }

  function togglePanel(name) {
    panel = panel === name ? null : name;
  }

  function outcomeLabel(o) {
    switch (o) {
      case 'victory': return 'Victory';
      case 'defeat': return 'Defeat';
      case 'fled': return 'Fled';
      case 'timeout': return 'Combat Ended';
      default: return o || 'Combat Ended';
    }
  }

  function onImgError(ev, key) {
    const img = ev && ev.currentTarget;
    if (!img || img.dataset.fallback === '1') return;
    img.dataset.fallback = '1';
    img.src = hashedAvatar(key || 'npc');
  }
</script>

{#if visible}
<div class="battle-stage" class:ending={phase === 'ending'} role="dialog" aria-label="Combat">
  <div class="battle-backdrop" aria-hidden="true"></div>

  <header class="battle-header">
    <i class="material-icons header-icon">explore</i>
    <span class="header-label">COMBAT</span>
    {#if turn?.round}
      <span class="round-chip">Round {turn.round}</span>
    {/if}
    {#if turn?.actorName}
      <span class="turn-chip">{isMyTurn ? 'Your turn' : `${turn.actorName}'s turn`}</span>
    {/if}
    <div class="header-rule"></div>
  </header>

  {#if isMyTurn && deadlineMs > 0}
    <div class="decision-timer" title="Decision window">
      <div class="decision-timer-fill" style="width: {timerPct}%"></div>
    </div>
  {/if}

  <!-- Enemies upper-right -->
  <section class="enemy-strip" aria-label="Enemies">
    {#each enemies as enemy (enemy.id)}
      {@const pct = hpPct(enemy.hp, enemy.maxHp)}
      {@const dead = (enemy.hp ?? 0) <= 0}
      <button
        type="button"
        class="enemy-card"
        class:targeted={enemy.id === targetId}
        class:dead
        class:hit={fx && fx.targetId === enemy.id && fxKey}
        disabled={dead || phase !== 'active'}
        on:click={() => selectEnemy(enemy)}
      >
        <div class="nameplate">{enemy.name}</div>
        <div class="hp-row">
          <span class="hp-label">HP</span>
          <div class="hp-track">
            <div class="hp-fill" style="width: {pct}%; background: {hpColor(pct)}"></div>
          </div>
          <span class="hp-nums">{enemy.hp ?? 0} / {enemy.maxHp ?? 0}</span>
        </div>
        <div class="enemy-sprite-wrap">
          <img
            class="enemy-sprite"
            src={combatantPortrait(enemy, enemy.id || enemy.name)}
            alt=""
            on:error={(e) => onImgError(e, enemy.name)}
          />
          {#if enemy.id === targetId && !dead}
            <div class="target-ring" aria-hidden="true"></div>
          {/if}
        </div>
      </button>
    {/each}
  </section>

  <!-- Center FX -->
  <div class="fx-layer" aria-hidden="true">
    {#if fx && fxKey}
      <div class="fx-flash fx-{fx.fxId || 'slash'}" data-key={fxKey}>
        {#if fx.damage > 0}
          <span class="fx-dmg">-{fx.damage}</span>
        {:else if fx.result === 'miss'}
          <span class="fx-miss">Miss</span>
        {:else if fx.fxId === 'defend'}
          <span class="fx-miss">Defend</span>
        {/if}
      </div>
    {/if}
  </div>

  <!-- Player lower-left -->
  <section class="player-panel" aria-label="Player">
    <div class="player-bust">
      <img
        src={combatantPortrait(selfCombatant, selfId || selfName)}
        alt=""
        on:error={(e) => onImgError(e, selfName)}
      />
    </div>
    <div class="player-meta">
      <div class="player-name">{selfName}</div>
      <div class="hp-row player-hp">
        <span class="hp-label">HP</span>
        <div class="hp-track">
          <div
            class="hp-fill"
            style="width: {hpPct(selfHp, selfMaxHp)}%; background: {playerHpColor(hpPct(selfHp, selfMaxHp))}"
          ></div>
        </div>
        <span class="hp-nums">{selfHp} / {selfMaxHp}</span>
      </div>
      {#if selfMaxMana > 0}
        <div class="hp-row player-mp">
          <span class="hp-label">MP</span>
          <div class="hp-track">
            <div class="hp-fill mp" style="width: {hpPct(selfMana, selfMaxMana)}%"></div>
          </div>
          <span class="hp-nums">{selfMana} / {selfMaxMana}</span>
        </div>
      {/if}
      <div class="status-chips">
        {#if selfClass}
          <span class="chip class-chip"><i class="material-icons">swords</i> {selfClass}</span>
        {/if}
        {#if isMyTurn}
          <span class="chip focus-chip"><i class="material-icons">flare</i> Focused</span>
        {/if}
      </div>
    </div>
  </section>

  <!-- Bottom dock -->
  {#if phase === 'active'}
    <nav class="battle-dock" aria-label="Combat actions">
      <button type="button" class="dock-btn" class:active={!panel} on:click={doAttack}>
        <i class="material-icons">swords</i>
        <span>Attack</span>
      </button>
      <button type="button" class="dock-btn" class:active={panel === 'skills'} on:click={() => togglePanel('skills')}>
        <i class="material-icons">auto_awesome</i>
        <span>Skills</span>
      </button>
      <button type="button" class="dock-btn" on:click={doDefend}>
        <i class="material-icons">security</i>
        <span>Defend</span>
      </button>
      <button type="button" class="dock-btn" class:active={panel === 'items'} on:click={() => togglePanel('items')}>
        <i class="material-icons">shopping_bag</i>
        <span>Items</span>
      </button>
      <button type="button" class="dock-btn flee" on:click={doFlee}>
        <i class="material-icons">directions_run</i>
        <span>Flee</span>
      </button>
    </nav>

    {#if panel === 'skills'}
      <div class="dock-panel" role="menu">
        {#if skillEntries.length === 0}
          <div class="dock-empty">No skills equipped</div>
        {:else}
          {#each skillEntries as skill (skill.id || skill.name)}
            <button type="button" class="dock-panel-btn" on:click={() => castSkill(skill)}>
              <i class="material-icons">auto_awesome</i> {skill.name}
            </button>
          {/each}
        {/if}
      </div>
    {/if}

    {#if panel === 'items'}
      <div class="dock-panel" role="menu">
        {#if consumables.length === 0}
          <div class="dock-empty">No consumables</div>
        {:else}
          {#each consumables as item (item.id || item.name)}
            <button type="button" class="dock-panel-btn" on:click={() => useItem(item)}>
              <i class="material-icons">science</i> {item.name}
            </button>
          {/each}
        {/if}
      </div>
    {/if}
  {/if}

  <!-- Thin combat log -->
  {#if log.length}
    <aside class="combat-log" aria-label="Combat log">
      <div class="combat-log-title">♦ COMBAT LOG</div>
      {#each log.slice(-4) as line (line.id)}
        <div class="combat-log-line">&gt; {line.text}</div>
      {/each}
    </aside>
  {/if}

  {#if phase === 'ending'}
    <div class="outcome-panel" class:victory={outcome === 'victory'} class:defeat={outcome === 'defeat'} class:fled={outcome === 'fled'}>
      <div class="outcome-title">{outcomeLabel(outcome)}</div>
      {#if endMessage}
        <div class="outcome-msg">{endMessage}</div>
      {/if}
    </div>
  {/if}
</div>
{/if}

<style>
  .battle-stage {
    position: fixed;
    inset: 0;
    z-index: 9000;
    display: grid;
    grid-template-rows: auto auto 1fr auto auto;
    color: #f3f4f6;
    font-family: 'Cinzel', Georgia, serif;
    pointer-events: auto;
    animation: stageIn 0.28s ease-out;
  }

  @keyframes stageIn {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  .battle-backdrop {
    position: absolute;
    inset: 0;
    background:
      radial-gradient(ellipse at 70% 30%, rgba(80, 40, 20, 0.35), transparent 55%),
      radial-gradient(ellipse at 20% 80%, rgba(30, 50, 80, 0.25), transparent 50%),
      rgba(4, 5, 8, 0.82);
    backdrop-filter: blur(6px) saturate(70%) brightness(55%);
    -webkit-backdrop-filter: blur(6px) saturate(70%) brightness(55%);
    z-index: 0;
  }

  .battle-header,
  .decision-timer,
  .enemy-strip,
  .fx-layer,
  .player-panel,
  .battle-dock,
  .dock-panel,
  .combat-log,
  .outcome-panel {
    position: relative;
    z-index: 1;
  }

  .battle-header {
    display: flex;
    align-items: center;
    gap: 0.55rem;
    padding: 1rem 1.25rem 0.35rem;
    letter-spacing: 0.12em;
  }

  .header-icon {
    color: #d4a44a;
    font-size: 1.35rem;
  }

  .header-label {
    color: #e8c878;
    font-size: 1.05rem;
    font-weight: 600;
  }

  .round-chip,
  .turn-chip {
    font-family: system-ui, sans-serif;
    font-size: 0.72rem;
    letter-spacing: 0.04em;
    padding: 0.15rem 0.55rem;
    border-radius: 999px;
    border: 1px solid rgba(212, 164, 74, 0.45);
    color: #f5e6c0;
    background: rgba(20, 14, 8, 0.65);
  }

  .header-rule {
    flex: 1;
    height: 1px;
    margin-left: 0.5rem;
    background: linear-gradient(90deg, rgba(212, 164, 74, 0.55), transparent);
  }

  .decision-timer {
    margin: 0.25rem 1.25rem 0;
    height: 4px;
    border-radius: 2px;
    background: rgba(255, 255, 255, 0.08);
    overflow: hidden;
    border: 1px solid rgba(212, 164, 74, 0.25);
  }

  .decision-timer-fill {
    height: 100%;
    background: linear-gradient(90deg, #d4a44a, #f59e0b);
    transition: width 0.2s linear;
  }

  .enemy-strip {
    justify-self: end;
    align-self: start;
    display: flex;
    flex-wrap: wrap;
    justify-content: flex-end;
    gap: 0.85rem;
    padding: 0.75rem 1.25rem 0;
    max-width: min(72vw, 640px);
  }

  .enemy-card {
    appearance: none;
    background: transparent;
    border: none;
    color: inherit;
    padding: 0;
    width: clamp(140px, 18vw, 200px);
    cursor: pointer;
    text-align: center;
    font: inherit;
  }

  .enemy-card:disabled {
    cursor: default;
    opacity: 0.45;
    filter: grayscale(0.6);
  }

  .nameplate {
    display: inline-block;
    padding: 0.2rem 0.7rem;
    margin-bottom: 0.3rem;
    border: 1.5px solid rgba(212, 164, 74, 0.65);
    border-radius: 4px;
    background: rgba(10, 8, 6, 0.85);
    color: #f5e6c0;
    font-size: 0.85rem;
    letter-spacing: 0.03em;
  }

  .hp-row {
    display: grid;
    grid-template-columns: auto 1fr auto;
    align-items: center;
    gap: 0.35rem;
    margin: 0 auto 0.45rem;
    max-width: 100%;
    font-family: system-ui, sans-serif;
  }

  .hp-label {
    color: #d4a44a;
    font-size: 0.68rem;
    font-weight: 700;
  }

  .hp-track {
    height: 8px;
    border-radius: 3px;
    background: rgba(0, 0, 0, 0.55);
    border: 1px solid rgba(212, 164, 74, 0.35);
    overflow: hidden;
  }

  .hp-fill {
    height: 100%;
    border-radius: 2px;
    transition: width 0.35s ease-out;
  }

  .hp-fill.mp {
    background: #3b82f6;
  }

  .hp-nums {
    color: #e5e7eb;
    font-size: 0.68rem;
    white-space: nowrap;
  }

  .enemy-sprite-wrap {
    position: relative;
    width: 100%;
    aspect-ratio: 1;
    display: grid;
    place-items: center;
  }

  .enemy-sprite {
    width: 78%;
    height: 78%;
    object-fit: contain;
    image-rendering: pixelated;
    filter: drop-shadow(0 8px 16px rgba(0, 0, 0, 0.55));
  }

  .target-ring {
    position: absolute;
    bottom: 6%;
    left: 50%;
    width: 70%;
    height: 18%;
    transform: translateX(-50%);
    border: 2px dashed rgba(250, 204, 21, 0.9);
    border-radius: 50%;
    box-shadow: 0 0 12px rgba(250, 204, 21, 0.35);
    pointer-events: none;
  }

  .enemy-card.targeted .nameplate {
    box-shadow: 0 0 0 1px rgba(250, 204, 21, 0.5);
  }

  .enemy-card.hit .enemy-sprite {
    animation: hitFlash 0.35s ease-out;
  }

  @keyframes hitFlash {
    0% { filter: brightness(2.2) drop-shadow(0 0 8px #ef4444); }
    100% { filter: drop-shadow(0 8px 16px rgba(0, 0, 0, 0.55)); }
  }

  .fx-layer {
    position: absolute;
    inset: 28% 20% 35% 20%;
    display: grid;
    place-items: center;
    pointer-events: none;
    z-index: 2;
  }

  .fx-flash {
    animation: fxPop 0.55s ease-out;
    font-family: system-ui, sans-serif;
    font-weight: 800;
    font-size: clamp(1.4rem, 3vw, 2.2rem);
    text-shadow: 0 2px 10px rgba(0, 0, 0, 0.8);
  }

  .fx-dmg { color: #fca5a5; }
  .fx-miss { color: #e5e7eb; letter-spacing: 0.08em; }

  @keyframes fxPop {
    0% { opacity: 0; transform: translateY(8px) scale(0.85); }
    30% { opacity: 1; transform: translateY(0) scale(1.05); }
    100% { opacity: 0; transform: translateY(-18px) scale(1); }
  }

  .player-panel {
    position: absolute;
    left: 1.1rem;
    bottom: clamp(7.5rem, 16vh, 10rem);
    display: flex;
    align-items: flex-end;
    gap: 0.85rem;
    max-width: min(420px, 92vw);
  }

  .player-bust {
    width: clamp(84px, 12vw, 120px);
    aspect-ratio: 1;
    border: 2px solid rgba(212, 164, 74, 0.75);
    border-radius: 6px;
    overflow: hidden;
    background: rgba(8, 8, 10, 0.9);
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.45), inset 0 0 0 1px rgba(255, 220, 150, 0.12);
    flex-shrink: 0;
  }

  .player-bust img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    object-position: top center;
  }

  .player-meta {
    min-width: 0;
    flex: 1;
  }

  .player-name {
    font-size: clamp(1rem, 2vw, 1.25rem);
    color: #f8fafc;
    margin-bottom: 0.35rem;
    text-shadow: 0 2px 8px rgba(0, 0, 0, 0.65);
  }

  .player-hp,
  .player-mp {
    margin-bottom: 0.3rem;
    min-width: 180px;
  }

  .status-chips {
    display: flex;
    flex-wrap: wrap;
    gap: 0.35rem;
    margin-top: 0.35rem;
  }

  .chip {
    display: inline-flex;
    align-items: center;
    gap: 0.2rem;
    font-family: system-ui, sans-serif;
    font-size: 0.68rem;
    padding: 0.15rem 0.45rem;
    border-radius: 4px;
    border: 1px solid rgba(212, 164, 74, 0.5);
    background: rgba(12, 10, 8, 0.75);
    color: #f5e6c0;
  }

  .chip i { font-size: 0.85rem; }
  .focus-chip { border-color: rgba(168, 85, 247, 0.55); color: #e9d5ff; }

  .battle-dock {
    position: absolute;
    left: 50%;
    bottom: clamp(3.6rem, 9vh, 5rem);
    transform: translateX(-50%);
    display: flex;
    gap: 0.45rem;
    padding: 0.35rem;
    z-index: 3;
  }

  .dock-btn {
    min-width: clamp(72px, 10vw, 96px);
    padding: 0.55rem 0.65rem 0.45rem;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.15rem;
    border-radius: 6px;
    border: 1.5px solid rgba(212, 164, 74, 0.55);
    background: rgba(12, 10, 8, 0.88);
    color: #f5e6c0;
    cursor: pointer;
    font-family: 'Cinzel', Georgia, serif;
    font-size: 0.72rem;
    letter-spacing: 0.04em;
    box-shadow: 0 6px 16px rgba(0, 0, 0, 0.35);
  }

  .dock-btn i { font-size: 1.25rem; color: #d4a44a; }
  .dock-btn:hover,
  .dock-btn.active {
    border-color: #e8c878;
    box-shadow: 0 0 0 1px rgba(232, 200, 120, 0.35), 0 6px 16px rgba(0, 0, 0, 0.35);
    background: rgba(30, 22, 12, 0.95);
  }
  .dock-btn.flee { border-color: rgba(239, 68, 68, 0.45); }
  .dock-btn.flee i { color: #f87171; }

  .dock-panel {
    position: absolute;
    left: 50%;
    bottom: clamp(7.2rem, 15vh, 9rem);
    transform: translateX(-50%);
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    gap: 0.4rem;
    max-width: min(90vw, 520px);
    padding: 0.55rem;
    border: 1.5px solid rgba(212, 164, 74, 0.45);
    border-radius: 8px;
    background: rgba(8, 8, 10, 0.92);
    z-index: 4;
  }

  .dock-panel-btn {
    appearance: none;
    border: 1px solid rgba(212, 164, 74, 0.4);
    background: rgba(20, 16, 10, 0.9);
    color: #f5e6c0;
    border-radius: 5px;
    padding: 0.4rem 0.65rem;
    font-family: system-ui, sans-serif;
    font-size: 0.8rem;
    cursor: pointer;
    display: inline-flex;
    align-items: center;
    gap: 0.3rem;
  }
  .dock-panel-btn i { font-size: 1rem; color: #d4a44a; }
  .dock-panel-btn:hover { border-color: #e8c878; }
  .dock-empty {
    font-family: system-ui, sans-serif;
    font-size: 0.8rem;
    color: #9ca3af;
    padding: 0.35rem 0.6rem;
  }

  .combat-log {
    position: absolute;
    left: 50%;
    bottom: 0.55rem;
    transform: translateX(-50%);
    width: min(92vw, 720px);
    max-height: 3.6rem;
    overflow: hidden;
    padding: 0.35rem 0.65rem 0.45rem;
    border: 1px solid rgba(212, 164, 74, 0.4);
    border-radius: 6px;
    background: rgba(6, 6, 8, 0.88);
    font-family: system-ui, sans-serif;
    font-size: 0.7rem;
    line-height: 1.25;
    color: #d1d5db;
  }

  .combat-log-title {
    color: #d4a44a;
    font-size: 0.62rem;
    letter-spacing: 0.08em;
    margin-bottom: 0.15rem;
  }

  .combat-log-line {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .outcome-panel {
    position: absolute;
    inset: 0;
    display: grid;
    place-content: center;
    gap: 0.6rem;
    background: rgba(0, 0, 0, 0.55);
    z-index: 5;
    text-align: center;
    padding: 1.5rem;
    animation: stageIn 0.25s ease-out;
  }

  .outcome-title {
    font-size: clamp(1.8rem, 5vw, 2.8rem);
    letter-spacing: 0.14em;
    color: #e8c878;
    text-transform: uppercase;
  }

  .outcome-panel.victory .outcome-title { color: #fbbf24; }
  .outcome-panel.defeat .outcome-title { color: #f87171; }
  .outcome-panel.fled .outcome-title { color: #93c5fd; }

  .outcome-msg {
    font-family: system-ui, sans-serif;
    max-width: 28rem;
    color: #e5e7eb;
    font-size: 0.95rem;
  }

  /* Mobile stack (graceful — full polish is C3) */
  @media (max-width: 720px) {
    .enemy-strip {
      justify-self: center;
      max-width: 100%;
      padding-top: 0.35rem;
    }
    .enemy-card { width: clamp(110px, 40vw, 160px); }
    .player-panel {
      left: 0.6rem;
      right: 0.6rem;
      bottom: clamp(8.5rem, 22vh, 11rem);
      max-width: none;
    }
    .battle-dock {
      bottom: clamp(4.2rem, 12vh, 5.5rem);
      gap: 0.3rem;
      width: min(96vw, 420px);
      justify-content: space-between;
    }
    .dock-btn {
      min-width: 0;
      flex: 1;
      padding: 0.45rem 0.2rem 0.35rem;
      font-size: 0.62rem;
    }
    .dock-btn i { font-size: 1.1rem; }
    .combat-log { max-height: 2.8rem; font-size: 0.64rem; }
  }
</style>
