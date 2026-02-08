<script>
  export let store = null;
  export let sendMessage = null;

  let quests = [];
  let expandedQuest = null;

  // Subscribe to store for reactive updates
  $: if (store) {
    quests = $store.quests || [];
  }

  $: activeQuests = quests.filter((q) => q.status === "active");
  $: completedQuests = quests.filter((q) => q.status === "completed");

  function toggleQuest(questId) {
    expandedQuest = expandedQuest === questId ? null : questId;
  }
</script>

<div class="questlog-widget">
  <div class="questlog-header">
    <h2>Quest Log</h2>
  </div>

  <div class="questlog-content">
    {#if activeQuests.length === 0 && completedQuests.length === 0}
      <div class="empty-state">No quests in your log.</div>
    {/if}

    {#if activeQuests.length > 0}
      <div class="quest-section">
        <h3 class="section-title">Active ({activeQuests.length})</h3>
        {#each activeQuests as quest}
          <div class="quest-entry" class:expanded={expandedQuest === quest.questId}>
            <button
              class="quest-name"
              on:click={() => toggleQuest(quest.questId)}
            >
              <span class="quest-indicator active"></span>
              <span>{quest.questName || 'Unnamed Quest'}</span>
              <span class="expand-icon">{expandedQuest === quest.questId ? "▾" : "▸"}</span>
            </button>

            {#if expandedQuest === quest.questId}
              <div class="quest-details">
                {#if quest.description}
                  <p class="quest-description">{quest.description}</p>
                {/if}
                <div class="objectives">
                  {#each quest.objectives || [] as obj}
                    <div class="objective" class:completed={obj.completed}>
                      <span class="check">{obj.completed ? "[x]" : "[ ]"}</span>
                      <span class="obj-text">{obj.description}</span>
                      <span class="progress">({obj.current}/{obj.required})</span>
                    </div>
                  {/each}
                </div>
              </div>
            {/if}
          </div>
        {/each}
      </div>
    {/if}

    {#if completedQuests.length > 0}
      <div class="quest-section">
        <h3 class="section-title">Completed ({completedQuests.length})</h3>
        {#each completedQuests as quest}
          <div class="quest-entry completed">
            <button
              class="quest-name"
              on:click={() => toggleQuest(quest.questId)}
            >
              <span class="quest-indicator completed"></span>
              <span>{quest.questName || 'Unnamed Quest'}</span>
            </button>
          </div>
        {/each}
      </div>
    {/if}
  </div>
</div>

<style>
  .questlog-widget {
    height: 100%;
    background: rgba(0, 0, 0, 0.85);
    backdrop-filter: blur(12px);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 8px;
    display: flex;
    flex-direction: column;
    color: #e5e7eb;
    font-family: "Fira Code", "Cascadia Code", monospace;
    font-size: 13px;
    overflow: hidden;
  }

  .questlog-header {
    padding: 12px 16px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  }

  .questlog-header h2 {
    margin: 0;
    font-size: 14px;
    text-transform: uppercase;
    letter-spacing: 0.1em;
    color: #f59e0b;
  }

  .questlog-content {
    flex: 1;
    overflow-y: auto;
    padding: 12px;
  }

  .empty-state {
    color: #6b7280;
    text-align: center;
    padding: 32px 16px;
    font-style: italic;
  }

  .quest-section {
    margin-bottom: 16px;
  }

  .section-title {
    font-size: 10px;
    text-transform: uppercase;
    letter-spacing: 0.15em;
    color: #6b7280;
    margin: 0 0 8px 0;
    padding-bottom: 4px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  }

  .quest-entry {
    margin-bottom: 4px;
    border-radius: 6px;
    overflow: hidden;
  }

  .quest-entry.expanded {
    background: rgba(255, 255, 255, 0.03);
  }

  .quest-name {
    display: flex;
    align-items: center;
    gap: 8px;
    width: 100%;
    padding: 8px 10px;
    background: none;
    border: none;
    color: #e5e7eb;
    cursor: pointer;
    font-size: 13px;
    text-align: left;
    border-radius: 6px;
    font-family: inherit;
  }

  .quest-name:hover {
    background: rgba(255, 255, 255, 0.05);
  }

  .quest-indicator {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    flex-shrink: 0;
  }

  .quest-indicator.active {
    background: #f59e0b;
    box-shadow: 0 0 6px rgba(245, 158, 11, 0.5);
  }

  .quest-indicator.completed {
    background: #22c55e;
  }

  .expand-icon {
    margin-left: auto;
    color: #6b7280;
    font-size: 11px;
  }

  .quest-details {
    padding: 4px 10px 12px 26px;
  }

  .quest-description {
    color: #9ca3af;
    font-size: 12px;
    margin: 0 0 8px 0;
    line-height: 1.4;
  }

  .objectives {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .objective {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 12px;
    color: #d1d5db;
  }

  .objective.completed {
    color: #22c55e;
  }

  .check {
    font-family: monospace;
    flex-shrink: 0;
    color: #6b7280;
  }

  .objective.completed .check {
    color: #22c55e;
  }

  .progress {
    margin-left: auto;
    color: #6b7280;
    font-size: 11px;
  }

  .quest-entry.completed .quest-name {
    color: #6b7280;
    text-decoration: line-through;
  }
</style>
