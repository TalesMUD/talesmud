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
  $: timerSec = Math.ceil(timerLeftMs / 1000);

  $: if (fx?.at) fxKey = fx.at;

  $: fxActive = !!(fx && fxKey);
  $: fxId = (fx && fx.fxId) || '';
  $: fxResult = (fx && fx.result) || '';
  $: fxDamage = Number(fx && fx.damage) || 0;
  $: fxHeal = Number(fx && fx.heal) || 0;
  $: fxIsMiss = fxActive && (fxId === 'miss' || fxResult === 'miss' || fxResult === 'dodged');
  $: fxIsDeath = fxActive && fxId === 'death';
  $: fxIsCast = fxActive && (fxId === 'cast' || fxResult === 'cast');
  $: fxIsDefend = fxActive && (fxId === 'defend' || fxResult === 'defended' || fxResult === 'block');
  $: fxIsFlee = fxActive && (fxId === 'flee' || fxResult === 'fled');
  $: fxIsHit = fxActive && !fxIsMiss && !fxIsDefend && !fxIsFlee && (
    fxId === 'slash' || fxIsDeath || fxResult === 'hit' || fxResult === 'crit' || fxDamage > 0
  );
  $: fxIsCrit = fxActive && fxResult === 'crit';

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
    // C3: tap portrait/sprite retargets only; Attack dock queues the hit
    if (!enemy || phase !== 'active') return;
    if ((enemy.hp ?? 0) <= 0) return;
    if (store.setCombatTarget) store.setCombatTarget(enemy.id);
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

  function isFxTarget(id) {
    return fxActive && id && fx && fx.targetId === id;
  }
  function isFxActor(id) {
    return fxActive && id && fx && fx.actorId === id;
  }
  function showFloatOn(id) {
    return isFxTarget(id) && (fxDamage > 0 || fxHeal > 0 || fxIsMiss);
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

  <div
    class="decision-timer"
    class:idle={!isMyTurn || deadlineMs <= 0}
    title="Decision window"
    aria-live="polite"
    aria-hidden={!(isMyTurn && deadlineMs > 0)}
  >
    {#if isMyTurn && deadlineMs > 0}
      <div class="decision-timer-fill" style="width: {timerPct}%"></div>
      <span class="decision-timer-label">{timerSec}s</span>
    {/if}
  </div>

  <!-- Enemies upper-right -->
  <section class="enemy-strip" aria-label="Enemies">
    {#each enemies as enemy (enemy.id)}
      {@const pct = hpPct(enemy.hp, enemy.maxHp)}
      {@const dead = (enemy.hp ?? 0) <= 0}
      {@const tgt = isFxTarget(enemy.id)}
      {@const act = isFxActor(enemy.id)}
      <button
        type="button"
        class="enemy-card"
        class:targeted={enemy.id === targetId}
        class:dead
        class:fx-hit={tgt && fxIsHit}
        class:fx-crit={tgt && fxIsCrit}
        class:fx-miss={tgt && fxIsMiss}
        class:fx-death={tgt && fxIsDeath}
        class:fx-cast={act && fxIsCast}
        disabled={dead || phase !== 'active'}
        aria-pressed={enemy.id === targetId}
        aria-label={`Target ${enemy.name || 'enemy'}`}
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
        <div class="enemy-sprite-wrap" class:shake={tgt && fxIsHit}>
          <img
            class="enemy-sprite"
            src={combatantPortrait(enemy, enemy.id || enemy.name)}
            alt=""
            on:error={(e) => onImgError(e, enemy.name)}
          />
          {#if enemy.id === targetId && !dead}
            <div class="target-ring" aria-hidden="true"></div>
          {/if}
          {#if tgt && fxIsMiss}
            <div class="fx-puff" data-key={fxKey} aria-hidden="true"></div>
          {/if}
          {#if tgt && fxIsHit && !fxIsMiss}
            <div class="fx-slash" class:crit={fxIsCrit} data-key={fxKey} aria-hidden="true"></div>
          {/if}
          {#if showFloatOn(enemy.id)}
            <div class="fx-float" data-key={fxKey}>
              {#if fxDamage > 0}
                <span class="fx-dmg" class:crit={fxIsCrit}>-{fxDamage}</span>
              {:else if fxHeal > 0}
                <span class="fx-heal">+{fxHeal}</span>
              {:else if fxIsMiss}
                <span class="fx-miss-label">Miss</span>
              {/if}
            </div>
          {/if}
        </div>
      </button>
    {/each}
  </section>

  <!-- Center FX burst (cast / stage-level accent) -->
  <div class="fx-layer" aria-hidden="true">
    {#if fxActive && (fxIsCast || fxIsDefend || fxIsFlee)}
      <div class="fx-burst fx-{fxId || 'cast'}" data-key={fxKey}>
        {#if fxIsDefend}
          <span class="fx-burst-label">Defend</span>
        {:else if fxIsFlee}
          <span class="fx-burst-label">Flee</span>
        {:else if fxIsCast && fxHeal <= 0 && fxDamage <= 0}
          <span class="fx-burst-label">Cast</span>
        {/if}
      </div>
    {/if}
  </div>

  <!-- Player lower-left -->
  <section
    class="player-panel"
    class:fx-hit={isFxTarget(selfId) && fxIsHit}
    class:fx-miss={isFxTarget(selfId) && fxIsMiss}
    class:fx-death={isFxTarget(selfId) && fxIsDeath}
    class:fx-cast={isFxActor(selfId) && (fxIsCast || fxIsDefend)}
    class:fx-defend={isFxActor(selfId) && fxIsDefend}
    aria-label="Player"
  >
    <div class="player-bust" class:shake={isFxTarget(selfId) && fxIsHit}>
      <img
        src={combatantPortrait(selfCombatant, selfId || selfName)}
        alt=""
        on:error={(e) => onImgError(e, selfName)}
      />
      {#if isFxTarget(selfId) && fxIsMiss}
        <div class="fx-puff" data-key={fxKey} aria-hidden="true"></div>
      {/if}
      {#if isFxActor(selfId) && fxIsCast}
        <div class="fx-cast-glow" data-key={fxKey} aria-hidden="true"></div>
      {/if}
      {#if isFxActor(selfId) && fxIsDefend}
        <div class="fx-shield" data-key={fxKey} aria-hidden="true"></div>
      {/if}
      {#if showFloatOn(selfId)}
        <div class="fx-float" data-key={fxKey}>
          {#if fxDamage > 0}
            <span class="fx-dmg" class:crit={fxIsCrit}>-{fxDamage}</span>
          {:else if fxHeal > 0}
            <span class="fx-heal">+{fxHeal}</span>
          {:else if fxIsMiss}
            <span class="fx-miss-label">Miss</span>
          {/if}
        </div>
      {/if}
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

  <!-- Bottom dock (desktop Classic + mobile thumb dock) -->
  {#if phase === 'active'}
    <nav class="battle-dock" aria-label="Combat actions">
      <button type="button" class="dock-btn primary" class:active={!panel} on:click={doAttack}>
        <i class="material-icons">swords</i>
        <span>Attack</span>
      </button>
      <button type="button" class="dock-btn" class:active={panel === 'skills'} on:click={() => togglePanel('skills')}>
        <i class="material-icons">auto_awesome</i>
        <span>Skills</span>
      </button>
      <button type="button" class="dock-btn secondary" on:click={doDefend}>
        <i class="material-icons">security</i>
        <span>Defend</span>
      </button>
      <button type="button" class="dock-btn items-btn" class:active={panel === 'items'} on:click={() => togglePanel('items')}>
        <i class="material-icons">shopping_bag</i>
        <span>Items</span>
      </button>
      <button type="button" class="dock-btn flee secondary" on:click={doFlee}>
        <i class="material-icons">directions_run</i>
        <span>Flee</span>
      </button>
    </nav>

    {#if panel === 'skills' || panel === 'items'}
      <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
      <div class="dock-sheet-backdrop" on:click={() => panel = null}></div>
      <div class="dock-panel" class:sheet={true} role="menu" aria-label={panel === 'skills' ? 'Skills' : 'Items'}>
        <div class="dock-sheet-head">
          <div class="dock-sheet-tabs">
            <button type="button" class="dock-tab" class:active={panel === 'skills'} on:click={() => panel = 'skills'}>
              <i class="material-icons">auto_awesome</i> Skills
            </button>
            <button type="button" class="dock-tab" class:active={panel === 'items'} on:click={() => panel = 'items'}>
              <i class="material-icons">shopping_bag</i> Items
            </button>
          </div>
          <button type="button" class="dock-sheet-close" aria-label="Close" on:click={() => panel = null}>
            <i class="material-icons">close</i>
          </button>
        </div>
        {#if panel === 'skills'}
          {#if skillEntries.length === 0}
            <div class="dock-empty">No skills equipped</div>
          {:else}
            {#each skillEntries as skill (skill.id || skill.name)}
              <button type="button" class="dock-panel-btn" on:click={() => castSkill(skill)}>
                <i class="material-icons">auto_awesome</i> {skill.name}
              </button>
            {/each}
          {/if}
        {:else}
          {#if consumables.length === 0}
            <div class="dock-empty">No consumables</div>
          {:else}
            {#each consumables as item (item.id || item.name)}
              <button type="button" class="dock-panel-btn" on:click={() => useItem(item)}>
                <i class="material-icons">science</i> {item.name}
              </button>
            {/each}
          {/if}
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
    position: relative;
  }
  .decision-timer.idle {
    height: 0;
    margin: 0;
    border: none;
    background: transparent;
    overflow: hidden;
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

  /* ===== C5 Combat FX pack (CSS/transform only) ===== */

  .enemy-card.fx-hit .enemy-sprite,
  .player-panel.fx-hit .player-bust img {
    animation: hitFlash 0.4s ease-out;
  }

  .enemy-card.fx-crit .enemy-sprite {
    animation: hitFlashCrit 0.45s ease-out;
  }

  .enemy-sprite-wrap.shake,
  .player-bust.shake {
    animation: hitShake 0.4s ease-out;
  }

  .enemy-card.fx-death .enemy-sprite {
    animation: deathDissolve 0.9s ease-out forwards;
  }

  .player-panel.fx-death .player-bust img {
    animation: deathDissolve 0.9s ease-out forwards;
  }

  .enemy-card.fx-cast .enemy-sprite,
  .player-panel.fx-cast .player-bust {
    animation: castGlow 0.7s ease-out;
  }

  .enemy-card.fx-miss .enemy-sprite {
    animation: missDim 0.45s ease-out;
  }

  @keyframes hitFlash {
    0% { filter: brightness(2.4) saturate(1.4) drop-shadow(0 0 10px #ef4444); }
    40% { filter: brightness(1.6) drop-shadow(0 0 6px #f87171); }
    100% { filter: drop-shadow(0 8px 16px rgba(0, 0, 0, 0.55)); }
  }

  @keyframes hitFlashCrit {
    0% { filter: brightness(2.8) saturate(1.6) drop-shadow(0 0 14px #fbbf24); }
    50% { filter: brightness(1.8) drop-shadow(0 0 10px #f59e0b); }
    100% { filter: drop-shadow(0 8px 16px rgba(0, 0, 0, 0.55)); }
  }

  @keyframes hitShake {
    0%, 100% { transform: translate3d(0, 0, 0); }
    15% { transform: translate3d(-7px, 1px, 0) rotate(-1.5deg); }
    30% { transform: translate3d(7px, -1px, 0) rotate(1.5deg); }
    45% { transform: translate3d(-5px, 0, 0); }
    60% { transform: translate3d(5px, 1px, 0); }
    75% { transform: translate3d(-2px, 0, 0); }
  }

  @keyframes deathDissolve {
    0% { opacity: 1; filter: brightness(1.8) drop-shadow(0 0 12px #f87171); transform: scale(1); }
    35% { opacity: 0.85; filter: brightness(1.2) grayscale(0.3); }
    100% {
      opacity: 0.2;
      filter: grayscale(1) brightness(0.55);
      transform: scale(0.92) translateY(6px);
    }
  }

  @keyframes castGlow {
    0% { box-shadow: 0 0 0 0 rgba(167, 139, 250, 0); filter: brightness(1); }
    30% {
      box-shadow: 0 0 22px 6px rgba(167, 139, 250, 0.55), 0 0 40px 2px rgba(232, 200, 120, 0.25);
      filter: brightness(1.35) saturate(1.25);
    }
    100% { box-shadow: 0 0 0 0 rgba(167, 139, 250, 0); filter: brightness(1); }
  }

  @keyframes missDim {
    0% { filter: brightness(1.15) drop-shadow(0 0 6px #94a3b8); opacity: 1; }
    40% { filter: brightness(0.85); opacity: 0.85; }
    100% { filter: drop-shadow(0 8px 16px rgba(0, 0, 0, 0.55)); opacity: 1; }
  }

  .fx-float {
    position: absolute;
    left: 50%;
    top: 28%;
    transform: translateX(-50%);
    z-index: 5;
    pointer-events: none;
    animation: floatNum 0.85s ease-out forwards;
    font-family: system-ui, sans-serif;
    font-weight: 800;
    font-size: clamp(1.15rem, 2.6vw, 1.85rem);
    text-shadow: 0 2px 8px rgba(0, 0, 0, 0.9), 0 0 4px rgba(0, 0, 0, 0.6);
    white-space: nowrap;
  }

  .fx-dmg { color: #fca5a5; }
  .fx-dmg.crit { color: #fde68a; font-size: 1.15em; }
  .fx-heal { color: #86efac; }
  .fx-miss-label {
    color: #e5e7eb;
    letter-spacing: 0.08em;
    font-size: 0.95em;
    font-weight: 700;
  }

  @keyframes floatNum {
    0% { opacity: 0; transform: translate(-50%, 10px) scale(0.8); }
    18% { opacity: 1; transform: translate(-50%, 0) scale(1.08); }
    100% { opacity: 0; transform: translate(-50%, -36px) scale(1); }
  }

  .fx-puff {
    position: absolute;
    left: 50%;
    top: 42%;
    width: 28%;
    aspect-ratio: 1;
    transform: translate(-50%, -50%);
    border-radius: 50%;
    pointer-events: none;
    z-index: 4;
    background: radial-gradient(circle, rgba(226, 232, 240, 0.85) 0%, rgba(148, 163, 184, 0.35) 45%, transparent 70%);
    box-shadow: 0 0 0 0 rgba(226, 232, 240, 0.4);
    animation: missPuff 0.55s ease-out forwards;
  }

  .fx-puff::after {
    content: '';
    position: absolute;
    inset: -35%;
    border-radius: 50%;
    border: 2px solid rgba(226, 232, 240, 0.45);
    animation: missPuffRing 0.55s ease-out forwards;
  }

  @keyframes missPuff {
    0% { opacity: 0; transform: translate(-50%, -50%) scale(0.35); }
    25% { opacity: 1; transform: translate(-50%, -50%) scale(1); }
    100% { opacity: 0; transform: translate(-50%, -58%) scale(1.55); }
  }

  @keyframes missPuffRing {
    0% { opacity: 0.8; transform: scale(0.6); }
    100% { opacity: 0; transform: scale(1.4); }
  }

  .fx-slash {
    position: absolute;
    left: 18%;
    top: 22%;
    width: 64%;
    height: 10%;
    pointer-events: none;
    z-index: 4;
    background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.95), #fca5a5, transparent);
    border-radius: 2px;
    transform: rotate(-28deg);
    box-shadow: 0 0 12px rgba(248, 113, 113, 0.7);
    animation: slashStreak 0.35s ease-out forwards;
  }

  .fx-slash.crit {
    background: linear-gradient(90deg, transparent, #fef3c7, #fbbf24, transparent);
    box-shadow: 0 0 16px rgba(251, 191, 36, 0.85);
    height: 12%;
  }

  @keyframes slashStreak {
    0% { opacity: 0; transform: rotate(-28deg) scaleX(0.2); }
    30% { opacity: 1; transform: rotate(-28deg) scaleX(1); }
    100% { opacity: 0; transform: rotate(-28deg) translateX(12%) scaleX(1.05); }
  }

  .fx-cast-glow,
  .fx-shield {
    position: absolute;
    inset: -6%;
    border-radius: inherit;
    pointer-events: none;
    z-index: 3;
  }

  .fx-cast-glow {
    background: radial-gradient(circle at 50% 40%, rgba(167, 139, 250, 0.45), transparent 65%);
    animation: castPulse 0.7s ease-out forwards;
  }

  .fx-shield {
    border: 2px solid rgba(96, 165, 250, 0.85);
    box-shadow: inset 0 0 18px rgba(59, 130, 246, 0.35), 0 0 16px rgba(59, 130, 246, 0.45);
    animation: shieldPulse 0.65s ease-out forwards;
  }

  @keyframes castPulse {
    0% { opacity: 0; transform: scale(0.85); }
    35% { opacity: 1; transform: scale(1.05); }
    100% { opacity: 0; transform: scale(1.15); }
  }

  @keyframes shieldPulse {
    0% { opacity: 0; transform: scale(0.9); }
    40% { opacity: 1; transform: scale(1.02); }
    100% { opacity: 0; transform: scale(1.08); }
  }

  .player-bust {
    position: relative;
  }

  .fx-layer {
    position: absolute;
    inset: 28% 20% 35% 20%;
    display: grid;
    place-items: center;
    pointer-events: none;
    z-index: 2;
  }

  .fx-burst {
    animation: fxBurstPop 0.65s ease-out forwards;
    font-family: system-ui, sans-serif;
    font-weight: 800;
    font-size: clamp(1.1rem, 2.4vw, 1.7rem);
    text-shadow: 0 2px 10px rgba(0, 0, 0, 0.8);
    letter-spacing: 0.06em;
  }

  .fx-burst.fx-cast {
    color: #ddd6fe;
    text-shadow: 0 0 16px rgba(167, 139, 250, 0.8), 0 2px 8px rgba(0, 0, 0, 0.85);
  }
  .fx-burst.fx-defend { color: #93c5fd; }
  .fx-burst.fx-flee { color: #fdba74; }

  .fx-burst-label { display: inline-block; }

  @keyframes fxBurstPop {
    0% { opacity: 0; transform: scale(0.75); }
    30% { opacity: 1; transform: scale(1.08); }
    100% { opacity: 0; transform: scale(1.2) translateY(-8px); }
  }

  /* Prefer GPU compositing for FX transforms */
  .enemy-sprite-wrap,
  .player-bust,
  .fx-float,
  .fx-puff,
  .fx-slash {
    will-change: transform, opacity;
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
    overflow: visible; /* C5: allow float numbers / puff outside bust */
    background: rgba(8, 8, 10, 0.9);
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.45), inset 0 0 0 1px rgba(255, 220, 150, 0.12);
    flex-shrink: 0;
  }

  .player-bust img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    object-position: top center;
    border-radius: 4px;
    display: block;
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

  .dock-btn.primary {
    border-color: #e8c878;
    border-width: 2px;
    background: rgba(36, 26, 12, 0.95);
    box-shadow: 0 0 0 1px rgba(232, 200, 120, 0.25), 0 6px 16px rgba(0, 0, 0, 0.35);
  }
  .dock-btn.primary i { color: #f5d78c; }

  .decision-timer-label {
    display: none;
  }

  .dock-sheet-backdrop {
    display: none;
  }
  .dock-sheet-head {
    display: none;
  }
  .dock-sheet-tabs,
  .dock-tab,
  .dock-sheet-close {
    display: none;
  }

  /* C3: mobile stacked stage — enemies ~35%, FX mid, player bar, thumb dock */
  @media (max-width: 768px) {
    .battle-stage {
      display: grid;
      grid-template-rows:
        auto
        auto
        minmax(0, 0.35fr)
        minmax(48px, 0.14fr)
        auto
        auto
        auto;
      grid-template-areas:
        "header"
        "timer"
        "enemies"
        "fx"
        "player"
        "dock"
        "log";
      padding-bottom: env(safe-area-inset-bottom, 0px);
      height: 100dvh;
      max-height: 100dvh;
      overflow: hidden;
    }

    .battle-header {
      grid-area: header;
      padding: 0.55rem 0.75rem 0.2rem;
      gap: 0.4rem;
    }
    .header-label { font-size: 0.92rem; }
    .header-icon { font-size: 1.15rem; }
    .round-chip,
    .turn-chip {
      font-size: 0.65rem;
      padding: 0.12rem 0.45rem;
    }
    .header-rule { display: none; }

    .decision-timer {
      grid-area: timer;
      margin: 0.2rem 0.75rem 0.15rem;
      height: 10px;
      border-radius: 5px;
      border-width: 1.5px;
    }
    .decision-timer.idle {
      height: 0;
      margin: 0;
      border: none;
    }
    .decision-timer-label {
      display: block;
      position: absolute;
      right: 0.45rem;
      top: 50%;
      transform: translateY(-50%);
      font-family: system-ui, sans-serif;
      font-size: 0.7rem;
      font-weight: 700;
      color: #f8fafc;
      text-shadow: 0 1px 2px rgba(0, 0, 0, 0.85);
      letter-spacing: 0.02em;
      pointer-events: none;
      z-index: 1;
    }

    .enemy-strip {
      grid-area: enemies;
      position: relative;
      justify-self: stretch;
      align-self: stretch;
      justify-content: center;
      align-content: center;
      max-width: none;
      width: 100%;
      padding: 0.25rem 0.6rem 0;
      gap: 0.55rem;
      overflow: hidden;
    }
    .enemy-card {
      width: clamp(120px, 42vw, 180px);
      touch-action: manipulation;
      -webkit-tap-highlight-color: transparent;
    }
    .enemy-card:active:not(:disabled) {
      transform: scale(0.97);
    }
    .enemy-sprite-wrap {
      aspect-ratio: 1;
      max-height: min(28vh, 200px);
    }
    .nameplate { font-size: 0.78rem; }
    .hp-track { height: 10px; }
    .hp-nums { font-size: 0.72rem; }

    .fx-layer {
      grid-area: fx;
      position: relative;
      inset: auto;
      min-height: 48px;
      z-index: 2;
    }

    .player-panel {
      grid-area: player;
      position: relative;
      left: auto;
      right: auto;
      bottom: auto;
      max-width: none;
      width: calc(100% - 1.2rem);
      margin: 0.15rem 0.6rem 0.25rem;
      padding: 0.45rem 0.55rem;
      border: 1.5px solid rgba(212, 164, 74, 0.55);
      border-radius: 8px;
      background: rgba(8, 8, 10, 0.88);
      box-shadow: 0 4px 16px rgba(0, 0, 0, 0.35);
      gap: 0.65rem;
      align-items: center;
    }
    .player-bust {
      width: clamp(64px, 16vw, 84px);
      border-radius: 50%;
    }
    .player-hp,
    .player-mp {
      min-width: 0;
    }
    .player-name {
      font-size: 0.95rem;
      margin-bottom: 0.2rem;
    }

    .battle-dock {
      grid-area: dock;
      position: relative;
      left: auto;
      bottom: auto;
      transform: none;
      width: calc(100% - 0.8rem);
      max-width: none;
      margin: 0.15rem 0.4rem 0.2rem;
      padding: 0.4rem;
      gap: 0.35rem;
      justify-content: stretch;
      border: 1.5px solid rgba(212, 164, 74, 0.5);
      border-radius: 10px;
      background: rgba(8, 8, 10, 0.92);
      box-shadow: 0 -4px 20px rgba(0, 0, 0, 0.35);
      z-index: 3;
    }
    .dock-btn {
      min-width: 0;
      flex: 1;
      min-height: 56px;
      padding: 0.55rem 0.2rem 0.4rem;
      font-size: 0.68rem;
      touch-action: manipulation;
      -webkit-tap-highlight-color: transparent;
    }
    .dock-btn i { font-size: 1.35rem; }
    .dock-btn.primary {
      flex: 1.55;
      min-height: 64px;
      font-size: 0.78rem;
      font-weight: 700;
    }
    .dock-btn.primary i { font-size: 1.55rem; }
    .dock-btn.secondary {
      opacity: 0.95;
    }
    /* Items moves into Skills/Items sheet on narrow; keep reachable via sheet tabs */
    .dock-btn.items-btn {
      display: none;
    }
    .dock-btn:active {
      transform: scale(0.97);
    }

    .dock-sheet-backdrop {
      display: block;
      position: absolute;
      inset: 0;
      background: rgba(0, 0, 0, 0.45);
      z-index: 4;
    }
    .dock-panel.sheet {
      position: absolute;
      left: 0;
      right: 0;
      bottom: 0;
      transform: none;
      max-width: none;
      width: 100%;
      max-height: min(52vh, 420px);
      padding: 0.55rem 0.65rem calc(0.65rem + env(safe-area-inset-bottom, 0px));
      border-radius: 14px 14px 0 0;
      border: 1.5px solid rgba(212, 164, 74, 0.5);
      border-bottom: none;
      background: rgba(8, 8, 10, 0.97);
      z-index: 5;
      flex-direction: column;
      flex-wrap: nowrap;
      align-items: stretch;
      gap: 0.35rem;
      overflow-y: auto;
      -webkit-overflow-scrolling: touch;
      animation: sheetUp 0.22s ease-out;
    }
    @keyframes sheetUp {
      from { transform: translateY(100%); }
      to { transform: translateY(0); }
    }
    .dock-sheet-head {
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 0.5rem;
      margin-bottom: 0.25rem;
      padding-bottom: 0.35rem;
      border-bottom: 1px solid rgba(212, 164, 74, 0.25);
    }
    .dock-sheet-tabs {
      display: flex;
      gap: 0.35rem;
    }
    .dock-tab {
      display: inline-flex;
      align-items: center;
      gap: 0.25rem;
      appearance: none;
      border: 1px solid rgba(212, 164, 74, 0.35);
      background: rgba(20, 16, 10, 0.85);
      color: #c4b5a0;
      border-radius: 999px;
      padding: 0.35rem 0.7rem;
      font-family: system-ui, sans-serif;
      font-size: 0.75rem;
      cursor: pointer;
      touch-action: manipulation;
    }
    .dock-tab i { font-size: 0.95rem; color: #d4a44a; }
    .dock-tab.active {
      border-color: #e8c878;
      color: #f5e6c0;
      background: rgba(40, 28, 12, 0.95);
    }
    .dock-sheet-close {
      display: inline-flex;
      appearance: none;
      border: 1px solid rgba(255, 255, 255, 0.15);
      background: rgba(255, 255, 255, 0.06);
      color: #9ca3af;
      border-radius: 8px;
      padding: 0.35rem;
      cursor: pointer;
      touch-action: manipulation;
    }
    .dock-sheet-close i { font-size: 1.15rem; }
    .dock-panel-btn {
      min-height: 48px;
      justify-content: flex-start;
      font-size: 0.9rem;
      padding: 0.55rem 0.75rem;
      touch-action: manipulation;
    }

    .combat-log {
      grid-area: log;
      position: relative;
      left: auto;
      bottom: auto;
      transform: none;
      width: calc(100% - 1.2rem);
      margin: 0 0.6rem 0.35rem;
      max-height: 2.4rem;
      font-size: 0.62rem;
    }

    .outcome-panel {
      padding: 1rem;
    }
  }
</style>
