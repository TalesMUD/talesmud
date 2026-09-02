<script>
  export let store = null;
  export let onQuestClick = null;

  let notifications = [];
  let dismissingIds = new Set();

  $: if (store) {
    notifications = $store.questNotifications || [];
  }

  $: momentCards = notifications.filter((n) => n.type === 'accepted' || n.type === 'completed');
  $: corner = notifications.filter((n) => n.type !== 'accepted' && n.type !== 'completed');

  function dismissNotification(notification) {
    if (dismissingIds.has(notification.id)) return;

    dismissingIds.add(notification.id);
    dismissingIds = dismissingIds;

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
    if (onQuestClick) {
      onQuestClick(notification.questId);
    }
    dismissNotification(notification);
  }

  function isDismissing(notification) {
    return dismissingIds.has(notification.id);
  }

  function kickerFor(notification) {
    return notification.type === 'completed' ? 'Quest complete' : 'Quest accepted';
  }

  function titleFor(notification) {
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
</script>

<svelte:head>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Cinzel:wght@400;600;700&display=swap" rel="stylesheet">
</svelte:head>

{#if momentCards.length > 0}
  <div class="quest-moment-layer" aria-live="polite">
    {#each momentCards as notification (notification.id || notification)}
      <div
        class="veilspan-card"
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
        {#if detailLines(notification).length}
          <div class="card-list-label">{listLabel(notification)}</div>
          <ul class="card-list">
            {#each detailLines(notification) as line}
              <li>{line}</li>
            {/each}
          </ul>
        {/if}
        <button
          class="accept-open"
          type="button"
          on:click={() => handleNotificationClick(notification)}
        >
          Open quest log
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
    max-height: min(62vh, 400px);
    overflow: auto;
    text-align: center;
    background:
      linear-gradient(165deg, rgba(33, 27, 21, 0.96) 0%, rgba(11, 16, 23, 0.96) 100%);
    border: 1.4px solid rgba(211, 173, 99, 0.55);
    border-radius: 8px;
    padding: 1.05rem 1.1rem 0.9rem;
    box-shadow:
      0 18px 48px rgba(0, 0, 0, 0.55),
      inset 0 0 0 1px rgba(116, 91, 51, 0.28);
    animation: acceptPop 0.28s ease-out;
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
    margin-bottom: 0.35rem;
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
    margin-bottom: 0.55rem;
    line-height: 1.3;
  }

  .card-list-label {
    font-size: 0.68rem;
    font-weight: 700;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: rgba(211, 173, 99, 0.9);
    margin-bottom: 0.3rem;
  }

  .card-list {
    list-style: none;
    margin: 0 0 0.85rem;
    padding: 0;
    text-align: left;
    color: #e8e0d2;
    font-size: clamp(0.88rem, 1.68vw, 1.10rem);
    line-height: 1.35;
  }

  .card-list li {
    padding: 0.18rem 0 0.18rem 0.9rem;
    position: relative;
  }

  .card-list li::before {
    content: "•";
    position: absolute;
    left: 0;
    color: #d3ad63;
  }

  .accept-open {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 100%;
    padding: 0.55rem 0.75rem;
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
    top: 8px;
    right: 8px;
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
