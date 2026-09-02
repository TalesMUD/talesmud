<script>
  export let store = null;
  export let onQuestClick = null;

  let notifications = [];
  let dismissingIds = new Set();

  // Subscribe to store for reactive updates
  $: if (store) {
    notifications = $store.questNotifications || [];
  }

  $: accepted = notifications.filter((n) => n.type === 'accepted');
  $: corner = notifications.filter((n) => n.type !== 'accepted');

  function dismissNotification(notification) {
    if (dismissingIds.has(notification.id)) return;

    dismissingIds.add(notification.id);
    dismissingIds = dismissingIds; // Trigger reactivity

    // Wait for animation to finish before removing
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

  function acceptHeadline(notification) {
    const name = (notification.questName || '').trim();
    if (name && name.toLowerCase() !== 'quest') return name;
    return 'Quest accepted';
  }
</script>

{#if accepted.length > 0}
  <!-- Large CENTRAL accept banner — not the old top-right chip -->
  <div class="quest-accept-layer" aria-live="polite">
    {#each accepted as notification (notification.id || notification)}
      <div
        class="accept-card"
        class:slide-out={isDismissing(notification)}
        role="dialog"
        aria-label="Quest accepted"
      >
        <button
          class="dismiss-btn accept-dismiss"
          on:click={() => dismissNotification(notification)}
          aria-label="Dismiss"
          type="button"
        >
          ×
        </button>
        <div class="accept-kicker">QUEST ACCEPTED</div>
        <div class="accept-title">{acceptHeadline(notification)}</div>
        {#if notification.message}
          <pre class="accept-body">{notification.message}</pre>
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
  .quest-accept-layer {
    position: fixed;
    inset: 0;
    z-index: 1200;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 0.85rem;
    pointer-events: none;
    background: rgba(0, 0, 0, 0.35);
  }

  .accept-card {
    pointer-events: auto;
    position: relative;
    width: min(82vw, 340px);
    max-height: min(62vh, 380px);
    overflow: auto;
    background: rgba(8, 10, 16, 0.94);
    border: 1.5px solid rgba(245, 158, 11, 0.7);
    border-radius: 12px;
    padding: 0.85rem 0.95rem 0.8rem;
    box-shadow: 0 18px 48px rgba(0, 0, 0, 0.55);
    animation: acceptPop 0.28s ease-out;
    font-family: "Fira Code", "Cascadia Code", monospace;
  }

  .accept-card.slide-out {
    animation: acceptOut 0.25s ease-in forwards;
  }

  .accept-kicker {
    font-size: 0.72rem;
    font-weight: 700;
    letter-spacing: 0.14em;
    color: #fbbf24;
    margin-bottom: 0.35rem;
  }

  .accept-title {
    font-size: clamp(0.92rem, 1.6vw, 1.08rem);
    font-weight: 700;
    color: #fde68a;
    margin-bottom: 0.45rem;
    line-height: 1.22;
  }

  .accept-body {
    margin: 0 0 0.9rem;
    white-space: pre-wrap;
    word-break: break-word;
    color: #e5e7eb;
    font-size: 0.72rem;
    line-height: 1.28;
    max-height: 10rem;
    overflow-y: auto;
  }

  .accept-open {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 100%;
    padding: 0.55rem 0.75rem;
    border-radius: 8px;
    border: 1px solid rgba(245, 158, 11, 0.45);
    background: rgba(245, 158, 11, 0.12);
    color: #fde68a;
    font: inherit;
    font-size: 0.85rem;
    font-weight: 600;
    cursor: pointer;
  }

  .accept-open:hover {
    background: rgba(245, 158, 11, 0.22);
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

  .notification.completed {
    border-left: 4px solid #22c55e;
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

  .notification.completed .notification-title {
    color: #22c55e;
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
