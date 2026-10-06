<script>
  export let store = null;
  export let sendMessage = null;
  export let onQuestClick = null;

  const ATTR_ORDER = ['STR', 'DEX', 'INT', 'WIS', 'STA'];

  let notifications = [];
  let dismissingIds = new Set();
  let pendingSpend = {}; // attr -> local optimistic spends this card session

  $: if (store) {
    notifications = $store.questNotifications || [];
  }

  $: characterStats = store ? ($store.characterStats || {}) : {};
  $: liveUnspent = Number(characterStats.unspentAttributePoints || 0);
  $: attributes = Array.isArray(characterStats.attributes) ? characterStats.attributes : [];

  // One celebration at a time. Level-up outranks quest complete so XP gains
  // and unspent points are never buried under a reward toast.
  $: momentCards = prioritizeMoments(
    notifications.filter((n) => n.type === 'accepted' || n.type === 'completed' || n.type === 'levelup')
  );
  $: activeMoment = momentCards.length ? [momentCards[0]] : [];
  $: corner = notifications.filter((n) => n.type !== 'accepted' && n.type !== 'completed' && n.type !== 'levelup');

  function prioritizeMoments(list) {
    const rank = { levelup: 0, completed: 1, accepted: 2 };
    return [...list].sort((a, b) => {
      const ra = rank[a.type] ?? 9;
      const rb = rank[b.type] ?? 9;
      if (ra !== rb) return ra - rb;
      return 0;
    });
  }

  function dismissNotification(notification) {
    if (dismissingIds.has(notification.id)) return;

    dismissingIds.add(notification.id);
    dismissingIds = dismissingIds;
    if (notification.type === 'levelup') {
      pendingSpend = {};
    }

    setTimeout(() => {
      if (store) {
        store.update(state => {
          state.questNotifications = state.questNotifications.filter(n => n.id !== notification.id);
          return state;
        });
      }
      dismissingIds.delete(notification.id);
    }, 300);
  }

  function handleNotificationClick(notification) {
    if (notification.type === 'levelup') {
      dismissNotification(notification);
      return;
    }
    if (onQuestClick) {
      onQuestClick(notification.questId);
    }
    dismissNotification(notification);
  }

  function isDismissing(notification) {
    return dismissingIds.has(notification.id);
  }

  function kickerFor(notification) {
    if (notification.type === 'levelup') return 'Level up';
    return notification.type === 'completed' ? 'Quest complete' : 'Quest accepted';
  }

  function titleFor(notification) {
    if (notification.type === 'levelup') {
      const lvl = notification.newLevel || 0;
      return lvl ? `Level ${lvl}` : 'Level Up';
    }
    const name = (notification.questName || '').trim();
    if (name && name.toLowerCase() !== 'quest') return name;
    return kickerFor(notification);
  }

  function stripBoxLines(text) {
    return String(text || '')
      .replace(/[\u2500-\u257F╔╗╚╝║═╠╣╦╩╬]/g, '')
      .split('\n')
      .map((line) => line.replace(/^\s*[-*•]\s*/, '').trim())
      .filter((line) => {
        if (!line) return false;
        if (/^QUEST\s+(ACCEPTED|COMPLETED)\b/i.test(line)) return false;
        if (/^(OBJECTIVES|REWARDS)\s*:?$/i.test(line)) return false;
        if (/^items\s*:?$/i.test(line)) return false;
        return true;
      });
  }

  function detailLines(notification) {
    const fromObjectives = (notification.objectives || [])
      .map((o) => (o && (o.description || o.Description)) || '')
      .map((s) => s.trim())
      .filter(Boolean);
    if (notification.type !== 'completed' && fromObjectives.length) {
      return fromObjectives;
    }
    const fromMessage = stripBoxLines(notification.message);
    const title = titleFor(notification).toLowerCase();
    return fromMessage.filter((line) => line.toLowerCase() !== title);
  }

  function listLabel(notification) {
    if (notification.type === 'completed') return 'Rewards';
    return detailLines(notification).length ? 'Objectives' : '';
  }

  function levelGainRows(notification) {
    const rows = [];
    const hp = Number(notification.hpGained || 0);
    if (hp > 0) {
      const now = notification.maxHitPoints
        ? ` (now ${notification.maxHitPoints} HP)`
        : '';
      rows.push({ key: 'hp', label: `+${hp} Max HP${now}` });
    }
    const mana = Number(notification.manaGained || 0);
    if (mana > 0) {
      const now = notification.maxMana
        ? ` (now ${notification.maxMana} Mana)`
        : '';
      rows.push({ key: 'mana', label: `+${mana} Max Mana${now}` });
    }
    const gains = notification.attributeGains || {};
    for (const attr of ATTR_ORDER) {
      const amount = Number(gains[attr] || 0);
      if (amount > 0) {
        rows.push({ key: `gain-${attr}`, label: `+${amount} ${attr}` });
      }
    }
    // Any non-standard keys last
    for (const [attr, amount] of Object.entries(gains)) {
      if (ATTR_ORDER.includes(attr)) continue;
      const n = Number(amount || 0);
      if (n > 0) rows.push({ key: `gain-${attr}`, label: `+${n} ${attr}` });
    }
    const pts = Number(notification.attributePointsGained || 0);
    if (pts > 0) {
      rows.push({
        key: 'points',
        label: `+${pts} Attribute Point${pts === 1 ? '' : 's'}`,
      });
    }
    return rows;
  }

  function unspentFor(notification) {
    const payload = Math.max(0, Number(notification.unspentAttributePoints || 0));
    // After a local spend, trust the HUD (characterUpdate).
    if (Object.keys(pendingSpend).length) {
      return Math.max(0, liveUnspent);
    }
    // Prefer live when characterUpdate already arrived; else payload.
    if (liveUnspent > 0) return liveUnspent;
    return payload;
  }

  function attrCurrent(short) {
    const found = attributes.find(
      (a) => String(a.short || a.Short || '').toUpperCase() === short
    );
    return found ? (found.value ?? found.Value ?? '—') : '—';
  }

  function spendAttr(short, notification) {
    if (!sendMessage || unspentFor(notification) <= 0) return;
    pendingSpend = { ...pendingSpend, [short]: (pendingSpend[short] || 0) + 1 };
    sendMessage(`spend ${short}`);
  }

  function primaryActionLabel(notification) {
    if (notification.type === 'levelup') {
      return unspentFor(notification) > 0 ? 'Done allocating' : 'Continue';
    }
    return 'Open quest log';
  }
</script>

<svelte:head>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Cinzel:wght@400;600;700&display=swap" rel="stylesheet">
</svelte:head>

{#if activeMoment.length > 0}
  <div class="quest-moment-layer" aria-live="polite">
    {#each activeMoment as notification (notification.id || notification)}
      <div
        class="veilspan-card"
        class:level-up={notification.type === 'levelup'}
        class:slide-out={isDismissing(notification)}
        role="dialog"
        aria-label={kickerFor(notification)}
      >
        <button
          class="dismiss-btn accept-dismiss"
          on:click={() => dismissNotification(notification)}
          aria-label="Dismiss"
          type="button"
        >
          ×
        </button>
        <div class="card-kicker">{kickerFor(notification)}</div>
        <div class="card-title">{titleFor(notification)}</div>

        {#if notification.type === 'levelup'}
          {#if notification.levelsGained > 1}
            <div class="level-sub">+{notification.levelsGained} levels</div>
          {:else if notification.oldLevel}
            <div class="level-sub">Reached from level {notification.oldLevel}</div>
          {/if}

          {#if levelGainRows(notification).length}
            <div class="card-list-label">Stat increases</div>
            <ul class="card-list">
              {#each levelGainRows(notification) as row (row.key)}
                <li>{row.label}</li>
              {/each}
            </ul>
          {/if}

          {#if unspentFor(notification) > 0}
            <div class="card-list-label">Allocate points · {unspentFor(notification)} left</div>
            <div class="attr-spend-grid" role="group" aria-label="Spend attribute points">
              {#each ATTR_ORDER as attr}
                <button
                  type="button"
                  class="attr-spend"
                  disabled={!sendMessage || unspentFor(notification) <= 0}
                  on:click={() => spendAttr(attr, notification)}
                  title="Spend 1 point on {attr}"
                >
                  <span class="attr-spend-name">{attr}</span>
                  <span class="attr-spend-val">{attrCurrent(attr)}</span>
                  <span class="attr-spend-plus">+</span>
                </button>
              {/each}
            </div>
          {:else}
            <p class="level-flavor">You feel more powerful.</p>
          {/if}
        {:else}
          {#if detailLines(notification).length}
            <div class="card-list-label">{listLabel(notification)}</div>
            <ul class="card-list">
              {#each detailLines(notification) as line}
                <li>{line}</li>
              {/each}
            </ul>
          {/if}
        {/if}

        <button
          class="accept-open"
          type="button"
          on:click={() => handleNotificationClick(notification)}
        >
          {primaryActionLabel(notification)}
        </button>
      </div>
    {/each}
  </div>
{/if}

{#if corner.length > 0}
  <div class="quest-notifications">
    {#each corner as notification (notification.id || notification)}
      <div
        class="notification {notification.type}"
        class:slide-in={!isDismissing(notification)}
        class:slide-out={isDismissing(notification)}
        on:click={() => handleNotificationClick(notification)}
        on:keydown={(e) => e.key === 'Enter' && handleNotificationClick(notification)}
        role="button"
        tabindex="0"
      >
        <div class="notification-content">
          <div class="notification-title">{notification.questName}</div>
          <div class="notification-message">{notification.message}</div>
        </div>
        <button
          class="dismiss-btn"
          on:click|stopPropagation={() => dismissNotification(notification)}
          aria-label="Dismiss notification"
          type="button"
        >
          ×
        </button>
      </div>
    {/each}
  </div>
{/if}

<style>
  .quest-moment-layer {
    position: fixed;
    inset: 0;
    z-index: 1200;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 0.85rem;
    pointer-events: none;
    background: rgba(0, 0, 0, 0.38);
  }

  .veilspan-card {
    pointer-events: auto;
    position: relative;
    width: min(84vw, 360px);
    max-height: min(72vh, 480px);
    overflow: auto;
    text-align: center;
    background:
      linear-gradient(165deg, rgba(33, 27, 21, 0.96) 0%, rgba(11, 16, 23, 0.96) 100%);
    border: 1.4px solid rgba(211, 173, 99, 0.55);
    border-radius: 8px;
    padding: 1.4rem 1.55rem 1.25rem;
    box-shadow:
      0 18px 48px rgba(0, 0, 0, 0.55),
      inset 0 0 0 1px rgba(116, 91, 51, 0.28);
    animation: acceptPop 0.28s ease-out;
  }

  .veilspan-card.level-up {
    border-color: rgba(240, 195, 106, 0.75);
    box-shadow:
      0 18px 48px rgba(0, 0, 0, 0.55),
      0 0 28px rgba(240, 195, 106, 0.18),
      inset 0 0 0 1px rgba(116, 91, 51, 0.35);
  }

  .veilspan-card.slide-out {
    animation: acceptOut 0.25s ease-in forwards;
  }

  .card-kicker {
    font-family: "Cinzel", serif;
    font-size: 0.72rem;
    font-weight: 700;
    letter-spacing: 0.16em;
    text-transform: uppercase;
    color: #f0c36a;
    text-shadow: 0 0 10px rgba(255, 215, 140, 0.28);
    margin-bottom: 0.55rem;
  }

  .card-title {
    font-family: "Cinzel", serif;
    font-size: clamp(1.02rem, 2.1vw, 1.28rem);
    font-weight: 600;
    letter-spacing: 0.06em;
    color: #f0e6d3;
    text-shadow:
      0 0 10px rgba(255, 215, 140, 0.25),
      0 2px 4px rgba(0, 0, 0, 0.75);
    margin-bottom: 0.45rem;
    line-height: 1.3;
  }

  .level-sub {
    font-size: 0.78rem;
    color: rgba(232, 224, 210, 0.72);
    margin-bottom: 0.85rem;
  }

  .level-flavor {
    margin: 0.35rem 0 1.1rem;
    color: rgba(232, 224, 210, 0.78);
    font-size: 0.9rem;
  }

  .card-list-label {
    font-size: 0.68rem;
    font-weight: 700;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: rgba(211, 173, 99, 0.9);
    margin-bottom: 0.5rem;
  }

  .card-list {
    list-style: none;
    margin: 0 0 1.05rem;
    padding: 0 0.15rem;
    text-align: left;
    color: #e8e0d2;
    font-size: clamp(0.88rem, 1.68vw, 1.10rem);
    line-height: 1.35;
  }

  .card-list li {
    padding: 0.28rem 0.1rem 0.28rem 1.35rem;
    position: relative;
  }

  .card-list li::before {
    content: "•";
    position: absolute;
    left: 0;
    color: #d3ad63;
  }

  .attr-spend-grid {
    display: grid;
    grid-template-columns: repeat(5, minmax(0, 1fr));
    gap: 0.4rem;
    margin: 0 0 1.15rem;
  }

  .attr-spend {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.12rem;
    padding: 0.45rem 0.2rem 0.4rem;
    border-radius: 7px;
    border: 1px solid rgba(211, 173, 99, 0.4);
    background: rgba(211, 173, 99, 0.1);
    color: #f0e6d3;
    font: inherit;
    cursor: pointer;
    min-width: 0;
  }

  .attr-spend:hover:not(:disabled) {
    background: rgba(211, 173, 99, 0.22);
    border-color: rgba(240, 195, 106, 0.7);
  }

  .attr-spend:disabled {
    opacity: 0.45;
    cursor: default;
  }

  .attr-spend-name {
    font-size: 0.68rem;
    font-weight: 700;
    letter-spacing: 0.06em;
    color: #f0c36a;
  }

  .attr-spend-val {
    font-size: 0.92rem;
    font-weight: 600;
    color: #f0e6d3;
  }

  .attr-spend-plus {
    font-size: 0.85rem;
    font-weight: 700;
    color: #d3ad63;
    line-height: 1;
  }

  .accept-open {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 100%;
    padding: 0.78rem 1.05rem;
    border-radius: 8px;
    border: 1px solid rgba(211, 173, 99, 0.45);
    background: rgba(211, 173, 99, 0.12);
    color: #f0e6d3;
    font: inherit;
    font-size: 0.85rem;
    font-weight: 600;
    cursor: pointer;
  }

  .accept-open:hover {
    background: rgba(211, 173, 99, 0.22);
  }

  .accept-dismiss {
    position: absolute;
    top: 12px;
    right: 12px;
  }

  .quest-notifications {
    position: fixed;
    top: 80px;
    right: 20px;
    z-index: 1000;
    display: flex;
    flex-direction: column;
    gap: 10px;
    pointer-events: none;
  }

  .notification {
    background: rgba(0, 0, 0, 0.9);
    backdrop-filter: blur(10px);
    border: 1px solid rgba(255, 255, 255, 0.2);
    border-radius: 8px;
    padding: 12px 16px;
    min-width: 300px;
    max-width: 400px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.5);
    font-family: "Fira Code", "Cascadia Code", monospace;
    font-size: 13px;
    display: flex;
    align-items: flex-start;
    gap: 12px;
    cursor: pointer;
    transition: all 0.2s;
    pointer-events: auto;
  }

  .notification:hover {
    background: rgba(0, 0, 0, 0.95);
    border-color: rgba(255, 255, 255, 0.3);
    box-shadow: 0 6px 16px rgba(0, 0, 0, 0.6);
    transform: translateX(-4px);
  }

  .notification.slide-in {
    animation: slideIn 0.3s ease-out;
  }

  .notification.slide-out {
    animation: slideOut 0.3s ease-in forwards;
  }

  .notification-content {
    flex: 1;
    min-width: 0;
  }

  .notification.progress {
    border-left: 4px solid #4a9eff;
  }

  .notification.ready {
    border-left: 4px solid #facc15;
  }

  .notification-title {
    font-weight: bold;
    margin-bottom: 4px;
    color: #f59e0b;
  }

  .notification.progress .notification-title {
    color: #4a9eff;
  }

  .notification.ready .notification-title {
    color: #facc15;
  }

  .notification-message {
    color: #e5e7eb;
    line-height: 1.4;
    font-size: 12px;
  }

  .dismiss-btn {
    background: none;
    border: none;
    color: #6b7280;
    font-size: 20px;
    line-height: 1;
    cursor: pointer;
    padding: 0;
    width: 24px;
    height: 24px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 4px;
    transition: all 0.2s;
    flex-shrink: 0;
  }

  .dismiss-btn:hover {
    background: rgba(255, 255, 255, 0.1);
    color: #e5e7eb;
  }

  @media (max-width: 420px) {
    .attr-spend-grid {
      grid-template-columns: repeat(3, minmax(0, 1fr));
    }
  }

  @keyframes acceptPop {
    from {
      opacity: 0;
      transform: scale(0.94) translateY(10px);
    }
    to {
      opacity: 1;
      transform: scale(1) translateY(0);
    }
  }

  @keyframes acceptOut {
    from {
      opacity: 1;
      transform: scale(1);
    }
    to {
      opacity: 0;
      transform: scale(0.96) translateY(8px);
    }
  }

  @keyframes slideIn {
    from {
      transform: translateX(100%);
      opacity: 0;
    }
    to {
      transform: translateX(0);
      opacity: 1;
    }
  }

  @keyframes slideOut {
    from {
      transform: translateX(0);
      opacity: 1;
      max-height: 100px;
      margin-bottom: 10px;
    }
    to {
      transform: translateX(100%);
      opacity: 0;
      max-height: 0;
      margin-bottom: 0;
      padding-top: 0;
      padding-bottom: 0;
    }
  }
</style>
