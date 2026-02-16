<script>
  export let store = null;
  export let onQuestClick = null;

  let notifications = [];
  let dismissingIds = new Set();

  // Subscribe to store for reactive updates
  $: if (store) {
    notifications = $store.questNotifications || [];
  }

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
</script>

<div class="quest-notifications">
  {#each notifications as notification (notification.id || notification)}
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
      >
        ×
      </button>
    </div>
  {/each}
</div>

<style>
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

  .notification.accepted {
    border-left: 4px solid #f59e0b;
  }

  .notification.progress {
    border-left: 4px solid #4a9eff;
  }

  .notification.completed {
    border-left: 4px solid #22c55e;
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
