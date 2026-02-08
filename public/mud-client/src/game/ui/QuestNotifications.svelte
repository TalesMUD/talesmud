<script>
  export let store = null;

  let notifications = [];

  // Subscribe to store for reactive updates
  $: if (store) {
    notifications = $store.questNotifications || [];
  }
</script>

<div class="quest-notifications">
  {#each notifications as notification (notification)}
    <div class="notification {notification.type}" class:slide-in={true}>
      <div class="notification-title">{notification.questName}</div>
      <div class="notification-message">{notification.message}</div>
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
  }

  .notification.slide-in {
    animation: slideIn 0.3s ease-out;
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
</style>
