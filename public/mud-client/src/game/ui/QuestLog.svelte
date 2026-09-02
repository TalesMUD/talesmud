<script>
  import { writable } from "svelte/store";

  export let quests = [];
  export let visible = false;
  export let sendMessage = null;

  let expandedQuest = null;
  let pinnedQuests = [];

  // Filtering and sorting
  let filterCategory = 'all';
  let showCompleted = true;
  let showAbandoned = false;
  let sortBy = 'status';
  let searchQuery = '';
  let showHistory = false;
  let showToolsMenu = false;

  $: toolsActive = !!(searchQuery || filterCategory !== 'all' || sortBy !== 'status' || !showCompleted || showAbandoned);

  function toggleToolsMenu() {
    showToolsMenu = !showToolsMenu;
  }

  function closeToolsMenu() {
    showToolsMenu = false;
  }

  // Load pinned quests
  $: if (visible && typeof localStorage !== 'undefined') {
    const saved = localStorage.getItem('pinnedQuests');
    pinnedQuests = saved ? JSON.parse(saved) : [];
  }

  // Filter and sort
  $: filteredQuests = quests.filter((q) => {
    if (!showCompleted && q.status === 'completed') return false;
    if (!showAbandoned && q.status === 'abandoned') return false;
    if (filterCategory !== 'all' && q.category?.toLowerCase() !== filterCategory) return false;

    // Search filter
    if (searchQuery.trim()) {
      const query = searchQuery.toLowerCase();
      const nameMatch = (q.questName || '').toLowerCase().includes(query);
      const descMatch = (q.description || '').toLowerCase().includes(query);
      const objMatch = (q.objectives || []).some(obj =>
        (obj.description || '').toLowerCase().includes(query)
      );
      if (!nameMatch && !descMatch && !objMatch) return false;
    }

    return true;
  });

  $: sortedQuests = [...filteredQuests].sort((a, b) => {
    switch (sortBy) {
      case 'name':
        return (a.questName || '').localeCompare(b.questName || '');
      case 'level':
        return (a.level || 0) - (b.level || 0);
      case 'category':
        return (a.category || '').localeCompare(b.category || '');
      case 'status':
      default:
        const statusOrder = { ready: 0, active: 1, completed: 2, abandoned: 3, failed: 4 };
        return (questStatusRank(a, statusOrder) ?? 99) - (questStatusRank(b, statusOrder) ?? 99);
    }
  });

  $: pinnedQuestList = sortedQuests.filter((q) => pinnedQuests.includes(q.questId) && q.status === 'active');
  $: unpinnedActiveQuests = sortedQuests.filter((q) => !pinnedQuests.includes(q.questId) && q.status === 'active');
  $: completedQuests = sortedQuests.filter((q) => q.status === "completed");
  $: abandonedQuests = sortedQuests.filter((q) => q.status === "abandoned");
  $: failedQuests = sortedQuests.filter((q) => q.status === "failed");

  function toggleQuest(questId) {
    expandedQuest = expandedQuest === questId ? null : questId;
  }

  function close() {
    visible = false;
  }

  function abandonQuest(questName) {
    if (sendMessage && confirm(`Abandon quest: ${questName}?`)) {
      sendMessage(`abandon ${questName}`);
      expandedQuest = null;
    }
  }

  function togglePin(questId) {
    if (pinnedQuests.includes(questId)) {
      pinnedQuests = pinnedQuests.filter(id => id !== questId);
    } else {
      if (pinnedQuests.length >= 5) {
        alert('You can only pin up to 5 quests at once.');
        return;
      }
      pinnedQuests = [...pinnedQuests, questId];
    }

    if (typeof localStorage !== 'undefined') {
      localStorage.setItem('pinnedQuests', JSON.stringify(pinnedQuests));
    }
  }

  function getCategoryColor(category) {
    switch (category?.toLowerCase()) {
      case 'main': return '#f59e0b';
      case 'side': return '#3b82f6';
      case 'daily': return '#8b5cf6';
      default: return '#6b7280';
    }
  }

  function getCategoryLabel(category) {
    if (!category) return '';
    return category.charAt(0).toUpperCase() + category.slice(1);
  }

  // Quest statistics
  $: questStats = {
    total: quests.length,
    active: quests.filter(q => q.status === 'active').length,
    completed: quests.filter(q => q.status === 'completed').length,
    abandoned: quests.filter(q => q.status === 'abandoned').length,
    failed: quests.filter(q => q.status === 'failed').length,
    mainCompleted: quests.filter(q => q.status === 'completed' && q.category === 'main').length,
    sideCompleted: quests.filter(q => q.status === 'completed' && q.category === 'side').length,
    dailyCompleted: quests.filter(q => q.status === 'completed' && q.category === 'daily').length,
    totalXP: quests
      .filter(q => q.status === 'completed')
      .reduce((sum, q) => sum + (q.rewards?.xp || 0), 0),
    totalGold: quests
      .filter(q => q.status === 'completed')
      .reduce((sum, q) => sum + (q.rewards?.gold || 0), 0),
    completionRate: quests.length > 0
      ? Math.round((quests.filter(q => q.status === 'completed').length / quests.length) * 100)
      : 0,
  };

  // Quest achievements
  $: achievements = [
    { name: 'First Steps', desc: 'Complete your first quest', unlocked: questStats.completed >= 1 },
    { name: 'Quest Novice', desc: 'Complete 5 quests', unlocked: questStats.completed >= 5 },
    { name: 'Quest Veteran', desc: 'Complete 10 quests', unlocked: questStats.completed >= 10 },
    { name: 'Quest Master', desc: 'Complete 25 quests', unlocked: questStats.completed >= 25 },
    { name: 'Story Seeker', desc: 'Complete 5 main quests', unlocked: questStats.mainCompleted >= 5 },
    { name: 'Side Quest Hero', desc: 'Complete 10 side quests', unlocked: questStats.sideCompleted >= 10 },
    { name: 'Daily Devotee', desc: 'Complete 5 daily quests', unlocked: questStats.dailyCompleted >= 5 },
    { name: 'Completionist', desc: '100% quest completion rate', unlocked: questStats.completionRate === 100 && questStats.total >= 5 },
  ];

  $: unlockedAchievements = achievements.filter(a => a.unlocked);
  $: lockedAchievements = achievements.filter(a => !a.unlocked);

  function isQuestReady(quest) {
    return quest.status === 'active'
      && (quest.objectives || []).length > 0
      && (quest.objectives || []).every((obj) => obj.completed);
  }

  function questStatusRank(quest, statusOrder) {
    return isQuestReady(quest) ? statusOrder.ready : statusOrder[quest.status];
  }

  function objectiveText(objective) {
    return objective.description || 'Objective';
  }
</script>

{#if visible}
  <div class="quest-log-overlay">
    <div class="quest-log-panel">
      <div class="quest-log-header">
        <div class="header-row">
          <h2>Quest Log</h2>
          <div class="header-actions">
            <button
              class="history-btn"
              on:click={() => { showHistory = !showHistory; showToolsMenu = false; }}
              title="View Quest History"
            >
              📊
            </button>
            <button
              class="tools-btn"
              class:active={showToolsMenu || toolsActive}
              on:click={toggleToolsMenu}
              title="Search and filters"
              aria-label="Search and filters"
              aria-expanded={showToolsMenu}
            >
              <i class="material-icons">more_horiz</i>
            </button>
            <button class="close-btn" on:click={close}>&times;</button>
          </div>
        </div>
        {#if showToolsMenu}
          <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
          <div class="tools-backdrop" on:click={closeToolsMenu}></div>
          <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
          <div class="tools-menu" on:click|stopPropagation>
            <div class="search-bar">
              <input
                type="text"
                bind:value={searchQuery}
                placeholder="Search quests..."
                class="search-input"
              />
              {#if searchQuery}
                <button class="clear-search-btn" on:click={() => searchQuery = ''}>×</button>
              {/if}
            </div>
            <div class="filter-controls">
              <select bind:value={filterCategory} class="filter-select">
                <option value="all">All Types</option>
                <option value="main">Main</option>
                <option value="side">Side</option>
                <option value="daily">Daily</option>
              </select>
              <select bind:value={sortBy} class="filter-select">
                <option value="status">Sort: Status</option>
                <option value="name">Sort: Name</option>
                <option value="level">Sort: Level</option>
                <option value="category">Sort: Type</option>
              </select>
            </div>
            <div class="toggle-controls">
              <label class="toggle-label">
                <input type="checkbox" bind:checked={showCompleted} />
                <span>Completed</span>
              </label>
              <label class="toggle-label">
                <input type="checkbox" bind:checked={showAbandoned} />
                <span>Abandoned</span>
              </label>
            </div>
          </div>
        {/if}
      </div>

      <div class="quest-log-content">
        {#if showHistory}
          <div class="history-panel">
            <div class="history-header">
              <h3>Quest Statistics</h3>
              <button class="close-history-btn" on:click={() => showHistory = false}>×</button>
            </div>

            <div class="stats-grid">
              <div class="stat-card">
                <div class="stat-value">{questStats.total}</div>
                <div class="stat-label">Total Quests</div>
              </div>
              <div class="stat-card active">
                <div class="stat-value">{questStats.active}</div>
                <div class="stat-label">Active</div>
              </div>
              <div class="stat-card completed">
                <div class="stat-value">{questStats.completed}</div>
                <div class="stat-label">Completed</div>
              </div>
              <div class="stat-card completion-rate">
                <div class="stat-value">{questStats.completionRate}%</div>
                <div class="stat-label">Completion</div>
              </div>
            </div>

            <div class="rewards-summary">
              <h4>Total Rewards Earned</h4>
              <div class="reward-summary-row">
                <span class="reward-icon">💫</span>
                <span class="reward-amount">{questStats.totalXP.toLocaleString()} XP</span>
              </div>
              <div class="reward-summary-row">
                <span class="reward-icon">💰</span>
                <span class="reward-amount">{questStats.totalGold.toLocaleString()} Gold</span>
              </div>
            </div>

            <div class="category-breakdown">
              <h4>By Category</h4>
              <div class="category-stat">
                <span class="category-name" style="color: #f59e0b">Main</span>
                <span class="category-count">{questStats.mainCompleted} completed</span>
              </div>
              <div class="category-stat">
                <span class="category-name" style="color: #3b82f6">Side</span>
                <span class="category-count">{questStats.sideCompleted} completed</span>
              </div>
              <div class="category-stat">
                <span class="category-name" style="color: #8b5cf6">Daily</span>
                <span class="category-count">{questStats.dailyCompleted} completed</span>
              </div>
            </div>

            <div class="achievements-section">
              <h4>Achievements ({unlockedAchievements.length}/{achievements.length})</h4>

              {#if unlockedAchievements.length > 0}
                <div class="achievements-list">
                  {#each unlockedAchievements as achievement}
                    <div class="achievement unlocked">
                      <span class="achievement-icon">🏆</span>
                      <div class="achievement-details">
                        <div class="achievement-name">{achievement.name}</div>
                        <div class="achievement-desc">{achievement.desc}</div>
                      </div>
                    </div>
                  {/each}
                </div>
              {/if}

              {#if lockedAchievements.length > 0}
                <details class="locked-achievements">
                  <summary>Locked Achievements ({lockedAchievements.length})</summary>
                  <div class="achievements-list">
                    {#each lockedAchievements as achievement}
                      <div class="achievement locked">
                        <span class="achievement-icon">🔒</span>
                        <div class="achievement-details">
                          <div class="achievement-name">{achievement.name}</div>
                          <div class="achievement-desc">{achievement.desc}</div>
                        </div>
                      </div>
                    {/each}
                  </div>
                </details>
              {/if}
            </div>

            <button class="back-to-quests-btn" on:click={() => showHistory = false}>
              Back to Quest Log
            </button>
          </div>
        {:else}
          {#if sortedQuests.length === 0}
            <div class="empty-state">No quests match your filters.</div>
          {/if}

          {#if pinnedQuestList.length > 0}
          <div class="quest-section">
            <h3 class="section-title">📌 Pinned ({pinnedQuestList.length})</h3>
            {#each pinnedQuestList as quest}
              <div class="quest-entry is-pinned" class:ready={isQuestReady(quest)} class:expanded={expandedQuest === quest.questId}>
                <button
                  class="quest-name"
                  on:click={() => toggleQuest(quest.questId)}
                >
                  <span class="quest-indicator active is-pinned" class:ready={isQuestReady(quest)}></span>
                  <div class="quest-title-row">
                    <span class="quest-title">{quest.questName || 'Unnamed Quest'}</span>
                    <div class="quest-badges">
                      {#if isQuestReady(quest)}
                        <span class="quest-badge ready-badge">Ready</span>
                      {/if}
                      {#if quest.level}
                        <span class="quest-badge level-badge">L{quest.level}</span>
                      {/if}
                      {#if quest.category}
                        <span class="quest-badge category-badge" style="background: {getCategoryColor(quest.category)}">
                          {getCategoryLabel(quest.category)}
                        </span>
                      {/if}
                    </div>
                  </div>
                  <span class="expand-icon">{expandedQuest === quest.questId ? "▾" : "▸"}</span>
                </button>

                {#if expandedQuest === quest.questId}
                  <div class="quest-details">
                    {#if quest.description}
                      <p class="quest-description">{quest.description}</p>
                    {/if}

                    <div class="objectives">
                      <div class="objectives-header">Objectives:</div>
                      {#each quest.objectives || [] as obj}
                        <div class="objective" class:completed={obj.completed}>
                          <span class="check">{obj.completed ? "[x]" : "[ ]"}</span>
                          <span class="obj-text">{objectiveText(obj)}</span>
                          <span class="quest-progress">({obj.current}/{obj.required})</span>
                        </div>
                      {/each}
                    </div>

                    {#if quest.rewards && (quest.rewards.xp > 0 || quest.rewards.gold > 0 || (quest.rewards.itemTemplateIds && quest.rewards.itemTemplateIds.length > 0))}
                      <div class="rewards">
                        <div class="rewards-header">Rewards:</div>
                        {#if quest.rewards.xp > 0}
                          <div class="reward-item">
                            <span class="reward-icon">💫</span>
                            <span class="reward-text">{quest.rewards.xp} XP</span>
                          </div>
                        {/if}
                        {#if quest.rewards.gold > 0}
                          <div class="reward-item">
                            <span class="reward-icon">💰</span>
                            <span class="reward-text">{quest.rewards.gold} Gold</span>
                          </div>
                        {/if}
                        {#if quest.rewards.itemTemplateIds && quest.rewards.itemTemplateIds.length > 0}
                          <div class="reward-item">
                            <span class="reward-icon">🎁</span>
                            <span class="reward-text">{quest.rewards.itemTemplateIds.length} item(s)</span>
                          </div>
                        {/if}
                      </div>
                    {/if}

                    <div class="quest-actions">
                      <button
                        class="pin-btn is-pinned"
                        on:click|stopPropagation={() => togglePin(quest.questId)}
                      >
                        Unpin Quest
                      </button>
                      <button
                        class="abandon-btn"
                        on:click|stopPropagation={() => abandonQuest(quest.questName)}
                      >
                        Abandon Quest
                      </button>
                    </div>
                  </div>
                {/if}
              </div>
            {/each}
          </div>
        {/if}

        {#if unpinnedActiveQuests.length > 0}
          <div class="quest-section">
            <h3 class="section-title">Active ({unpinnedActiveQuests.length})</h3>
            {#each unpinnedActiveQuests as quest}
              <div class="quest-entry" class:ready={isQuestReady(quest)} class:expanded={expandedQuest === quest.questId}>
                <button
                  class="quest-name"
                  on:click={() => toggleQuest(quest.questId)}
                >
                  <span class="quest-indicator active" class:ready={isQuestReady(quest)}></span>
                  <div class="quest-title-row">
                    <span class="quest-title">{quest.questName || 'Unnamed Quest'}</span>
                    <div class="quest-badges">
                      {#if isQuestReady(quest)}
                        <span class="quest-badge ready-badge">Ready</span>
                      {/if}
                      {#if quest.level}
                        <span class="quest-badge level-badge">L{quest.level}</span>
                      {/if}
                      {#if quest.category}
                        <span class="quest-badge category-badge" style="background: {getCategoryColor(quest.category)}">
                          {getCategoryLabel(quest.category)}
                        </span>
                      {/if}
                    </div>
                  </div>
                  <span class="expand-icon">{expandedQuest === quest.questId ? "▾" : "▸"}</span>
                </button>

                {#if expandedQuest === quest.questId}
                  <div class="quest-details">
                    {#if quest.description}
                      <p class="quest-description">{quest.description}</p>
                    {/if}

                    <div class="objectives">
                      <div class="objectives-header">Objectives:</div>
                      {#each quest.objectives || [] as obj}
                        <div class="objective" class:completed={obj.completed}>
                          <span class="check">{obj.completed ? "[x]" : "[ ]"}</span>
                          <span class="obj-text">{objectiveText(obj)}</span>
                          <span class="quest-progress">({obj.current}/{obj.required})</span>
                        </div>
                      {/each}
                    </div>

                    {#if quest.rewards && (quest.rewards.xp > 0 || quest.rewards.gold > 0 || (quest.rewards.itemTemplateIds && quest.rewards.itemTemplateIds.length > 0))}
                      <div class="rewards">
                        <div class="rewards-header">Rewards:</div>
                        {#if quest.rewards.xp > 0}
                          <div class="reward-item">
                            <span class="reward-icon">💫</span>
                            <span class="reward-text">{quest.rewards.xp} XP</span>
                          </div>
                        {/if}
                        {#if quest.rewards.gold > 0}
                          <div class="reward-item">
                            <span class="reward-icon">💰</span>
                            <span class="reward-text">{quest.rewards.gold} Gold</span>
                          </div>
                        {/if}
                        {#if quest.rewards.itemTemplateIds && quest.rewards.itemTemplateIds.length > 0}
                          <div class="reward-item">
                            <span class="reward-icon">🎁</span>
                            <span class="reward-text">{quest.rewards.itemTemplateIds.length} item(s)</span>
                          </div>
                        {/if}
                      </div>
                    {/if}

                    <div class="quest-actions">
                      <button
                        class="pin-btn"
                        on:click|stopPropagation={() => togglePin(quest.questId)}
                      >
                        Pin Quest
                      </button>
                      <button
                        class="abandon-btn"
                        on:click|stopPropagation={() => abandonQuest(quest.questName)}
                      >
                        Abandon Quest
                      </button>
                    </div>
                  </div>
                {/if}
              </div>
            {/each}
          </div>
        {/if}

        {#if showCompleted && completedQuests.length > 0}
          <div class="quest-section">
            <h3 class="section-title">Completed ({completedQuests.length})</h3>
            {#each completedQuests as quest}
              <div class="quest-entry completed" class:expanded={expandedQuest === quest.questId}>
                <button
                  class="quest-name"
                  on:click={() => toggleQuest(quest.questId)}
                >
                  <span class="quest-indicator completed"></span>
                  <div class="quest-title-row">
                    <span class="quest-title">{quest.questName || 'Unnamed Quest'}</span>
                    <div class="quest-badges">
                      {#if quest.level}
                        <span class="quest-badge level-badge">L{quest.level}</span>
                      {/if}
                      {#if quest.category}
                        <span class="quest-badge category-badge" style="background: {getCategoryColor(quest.category)}">
                          {getCategoryLabel(quest.category)}
                        </span>
                      {/if}
                    </div>
                  </div>
                  <span class="expand-icon">{expandedQuest === quest.questId ? "▾" : "▸"}</span>
                </button>

                {#if expandedQuest === quest.questId}
                  <div class="quest-details">
                    {#if quest.description}
                      <p class="quest-description">{quest.description}</p>
                    {/if}

                    <div class="objectives">
                      <div class="objectives-header">Objectives Completed:</div>
                      {#each quest.objectives || [] as obj}
                        <div class="objective completed">
                          <span class="check">[x]</span>
                          <span class="obj-text">{objectiveText(obj)}</span>
                          <span class="quest-progress">({obj.required}/{obj.required})</span>
                        </div>
                      {/each}
                    </div>

                    {#if quest.rewards && (quest.rewards.xp > 0 || quest.rewards.gold > 0 || (quest.rewards.itemTemplateIds && quest.rewards.itemTemplateIds.length > 0))}
                      <div class="rewards">
                        <div class="rewards-header">Rewards Earned:</div>
                        {#if quest.rewards.xp > 0}
                          <div class="reward-item earned">
                            <span class="reward-icon">💫</span>
                            <span class="reward-text">{quest.rewards.xp} XP</span>
                          </div>
                        {/if}
                        {#if quest.rewards.gold > 0}
                          <div class="reward-item earned">
                            <span class="reward-icon">💰</span>
                            <span class="reward-text">{quest.rewards.gold} Gold</span>
                          </div>
                        {/if}
                        {#if quest.rewards.itemTemplateIds && quest.rewards.itemTemplateIds.length > 0}
                          <div class="reward-item earned">
                            <span class="reward-icon">🎁</span>
                            <span class="reward-text">{quest.rewards.itemTemplateIds.length} item(s)</span>
                          </div>
                        {/if}
                      </div>
                    {/if}

                    {#if quest.completedAt}
                      <div class="quest-timestamp">
                        Completed: {new Date(quest.completedAt).toLocaleDateString()}
                      </div>
                    {/if}
                  </div>
                {/if}
              </div>
            {/each}
          </div>
        {/if}

        {#if abandonedQuests.length > 0}
          <div class="quest-section">
            <h3 class="section-title">Abandoned ({abandonedQuests.length})</h3>
            {#each abandonedQuests as quest}
              <div class="quest-entry abandoned">
                <button
                  class="quest-name"
                  on:click={() => toggleQuest(quest.questId)}
                >
                  <span class="quest-indicator abandoned"></span>
                  <div class="quest-title-row">
                    <span class="quest-title">{quest.questName || 'Unnamed Quest'}</span>
                  </div>
                  <span class="expand-icon">{expandedQuest === quest.questId ? "▾" : "▸"}</span>
                </button>
              </div>
            {/each}
          </div>
        {/if}

        {#if failedQuests.length > 0}
          <div class="quest-section">
            <h3 class="section-title">Failed ({failedQuests.length})</h3>
            {#each failedQuests as quest}
              <div class="quest-entry failed">
                <button
                  class="quest-name"
                  on:click={() => toggleQuest(quest.questId)}
                >
                  <span class="quest-indicator failed"></span>
                  <div class="quest-title-row">
                    <span class="quest-title">{quest.questName || 'Unnamed Quest'}</span>
                  </div>
                  <span class="expand-icon">{expandedQuest === quest.questId ? "▾" : "▸"}</span>
                </button>
              </div>
            {/each}
          </div>
        {/if}
        {/if}
      </div>
    </div>
  </div>
{/if}

<style>
  .quest-log-overlay {
    position: fixed;
    top: 0;
    right: 0;
    bottom: 0;
    width: 340px;
    z-index: 100;
    pointer-events: auto;
  }

  .quest-log-panel {
    height: 100%;
    background: rgba(0, 0, 0, 0.85);
    backdrop-filter: blur(12px);
    border-left: 1px solid rgba(255, 255, 255, 0.1);
    display: flex;
    flex-direction: column;
    color: #e5e7eb;
    font-family: "Fira Code", "Cascadia Code", monospace;
    font-size: 13px;
  }

  .quest-log-header {
    position: relative;
    padding: 8px 12px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
    display: flex;
    flex-direction: column;
    gap: 0;
  }

  .header-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.5em;
    min-height: 28px;
  }

  .header-actions {
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .quest-log-header h2 {
    margin: 0;
    font-size: 0.95rem;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: #f59e0b;
    line-height: 1.2;
  }

  .history-btn,
  .tools-btn {
    background: rgba(245, 158, 11, 0.1);
    border: 1px solid rgba(245, 158, 11, 0.3);
    color: #f59e0b;
    padding: 4px 8px;
    border-radius: 4px;
    cursor: pointer;
    font-size: 14px;
    transition: all 0.2s;
    line-height: 1;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 28px;
    min-height: 28px;
  }

  .tools-btn i {
    font-size: 18px;
  }

  .history-btn:hover,
  .tools-btn:hover,
  .tools-btn.active {
    background: rgba(245, 158, 11, 0.2);
    border-color: rgba(245, 158, 11, 0.5);
  }

  .tools-backdrop {
    position: fixed;
    inset: 0;
    z-index: 40;
    background: transparent;
  }

  .tools-menu {
    position: absolute;
    top: calc(100% - 1px);
    right: 8px;
    left: 8px;
    z-index: 45;
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding: 10px;
    background: rgba(12, 18, 28, 0.98);
    border: 1px solid rgba(255, 255, 255, 0.12);
    border-radius: 8px;
    box-shadow: 0 12px 28px rgba(0, 0, 0, 0.45);
  }

  .search-bar {
    position: relative;
    display: flex;
    align-items: center;
  }

  .search-input {
    flex: 1;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.1);
    color: #e5e7eb;
    padding: 6px 32px 6px 10px;
    border-radius: 4px;
    font-size: 12px;
    font-family: inherit;
  }

  .search-input:focus {
    outline: none;
    background: rgba(255, 255, 255, 0.08);
    border-color: rgba(245, 158, 11, 0.5);
  }

  .search-input::placeholder {
    color: #6b7280;
  }

  .clear-search-btn {
    position: absolute;
    right: 6px;
    background: none;
    border: none;
    color: #6b7280;
    font-size: 18px;
    cursor: pointer;
    padding: 0;
    width: 20px;
    height: 20px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 3px;
  }

  .clear-search-btn:hover {
    background: rgba(255, 255, 255, 0.1);
    color: #e5e7eb;
  }

  .filter-controls {
    display: flex;
    gap: 6px;
  }

  .filter-select {
    flex: 1;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.1);
    color: #e5e7eb;
    padding: 4px 8px;
    border-radius: 4px;
    font-size: 11px;
    font-family: inherit;
    cursor: pointer;
  }

  .filter-select:hover {
    background: rgba(255, 255, 255, 0.08);
  }

  .filter-select option {
    background: #1f2937;
    color: #e5e7eb;
  }

  .toggle-controls {
    display: flex;
    gap: 12px;
    font-size: 11px;
  }

  .toggle-label {
    display: flex;
    align-items: center;
    gap: 4px;
    cursor: pointer;
    color: #9ca3af;
  }

  .toggle-label:hover {
    color: #e5e7eb;
  }

  .toggle-label input[type="checkbox"] {
    cursor: pointer;
  }

  .toggle-label span {
    user-select: none;
  }

  .close-btn {
    background: none;
    border: none;
    color: #9ca3af;
    font-size: 20px;
    cursor: pointer;
    padding: 4px 8px;
    border-radius: 4px;
  }

  .close-btn:hover {
    background: rgba(255, 255, 255, 0.1);
    color: #e5e7eb;
  }

  .quest-log-content {
    flex: 1;
    min-height: 0;
    overflow-y: auto;
    padding: 12px;
    display: flex;
    flex-direction: column;
  }

  .empty-state {
    color: #6b7280;
    text-align: center;
    padding: 32px 16px;
    font-style: italic;
  }

  .quest-section {
    margin-bottom: 16px;
    flex-shrink: 0;
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
    display: flex;
    flex-direction: column;
    margin-bottom: 4px;
    border-radius: 6px;
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

  .quest-title-row {
    display: flex;
    flex-direction: column;
    gap: 4px;
    flex: 1;
    min-width: 0;
  }

  .quest-title {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .quest-badges {
    display: flex;
    gap: 4px;
    flex-wrap: wrap;
  }

  .quest-badge {
    display: inline-block;
    padding: 2px 6px;
    border-radius: 3px;
    font-size: 9px;
    font-weight: bold;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .level-badge {
    background: #4b5563;
    color: #e5e7eb;
  }

  .category-badge {
    color: #ffffff;
  }

  .ready-badge {
    background: rgba(250, 204, 21, 0.18);
    border: 1px solid rgba(250, 204, 21, 0.45);
    color: #fde68a;
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

  .quest-indicator.active.is-pinned {
    background: linear-gradient(135deg, #f59e0b 0%, #fbbf24 100%);
    box-shadow: 0 0 8px rgba(245, 158, 11, 0.8);
    animation: pulse 2s ease-in-out infinite;
  }

  .quest-indicator.active.ready {
    background: #facc15;
    box-shadow: 0 0 8px rgba(250, 204, 21, 0.8);
  }

  @keyframes pulse {
    0%, 100% {
      box-shadow: 0 0 8px rgba(245, 158, 11, 0.8);
    }
    50% {
      box-shadow: 0 0 12px rgba(245, 158, 11, 1);
    }
  }

  .quest-indicator.completed {
    background: #22c55e;
  }

  .quest-indicator.abandoned {
    background: #6b7280;
  }

  .quest-indicator.failed {
    background: #ef4444;
  }

  .quest-entry.is-pinned {
    background: rgba(245, 158, 11, 0.05);
    border: 1px solid rgba(245, 158, 11, 0.2);
    border-radius: 6px;
    margin-bottom: 8px;
  }

  .quest-entry.ready {
    background: rgba(250, 204, 21, 0.06);
    border: 1px solid rgba(250, 204, 21, 0.28);
    border-radius: 6px;
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

  .quest-progress {
    margin-left: auto;
    color: #6b7280;
    font-size: 11px;
  }

  .quest-entry.completed .quest-name .quest-title {
    color: #9ca3af;
  }

  .quest-entry.abandoned .quest-name,
  .quest-entry.failed .quest-name {
    color: #6b7280;
  }

  .objectives-header,
  .rewards-header {
    font-size: 11px;
    font-weight: bold;
    color: #9ca3af;
    margin-bottom: 6px;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .rewards {
    margin-top: 12px;
    padding-top: 8px;
    border-top: 1px solid rgba(255, 255, 255, 0.05);
  }

  .reward-item {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 12px;
    color: #d1d5db;
    padding: 2px 0;
  }

  .reward-item.earned {
    color: #22c55e;
  }

  .reward-icon {
    font-size: 14px;
    flex-shrink: 0;
  }

  .reward-text {
    flex: 1;
  }

  .quest-actions {
    margin-top: 12px;
    padding-top: 8px;
    border-top: 1px solid rgba(255, 255, 255, 0.05);
    display: flex;
    gap: 8px;
  }

  .abandon-btn {
    padding: 6px 12px;
    background: rgba(239, 68, 68, 0.1);
    border: 1px solid rgba(239, 68, 68, 0.3);
    color: #ef4444;
    border-radius: 4px;
    font-size: 11px;
    cursor: pointer;
    font-family: inherit;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    font-weight: bold;
    transition: all 0.2s;
  }

  .abandon-btn:hover {
    background: rgba(239, 68, 68, 0.2);
    border-color: rgba(239, 68, 68, 0.5);
  }

  .pin-btn {
    padding: 6px 12px;
    background: rgba(245, 158, 11, 0.1);
    border: 1px solid rgba(245, 158, 11, 0.3);
    color: #f59e0b;
    border-radius: 4px;
    font-size: 11px;
    cursor: pointer;
    font-family: inherit;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    font-weight: bold;
    transition: all 0.2s;
  }

  .pin-btn:hover {
    background: rgba(245, 158, 11, 0.2);
    border-color: rgba(245, 158, 11, 0.5);
  }

  .pin-btn.is-pinned {
    background: rgba(245, 158, 11, 0.2);
    border-color: rgba(245, 158, 11, 0.4);
  }

  .pin-btn.is-pinned:hover {
    background: rgba(239, 68, 68, 0.1);
    border-color: rgba(239, 68, 68, 0.3);
    color: #ef4444;
  }

  .quest-timestamp {
    margin-top: 8px;
    font-size: 10px;
    color: #6b7280;
    font-style: italic;
    text-align: right;
  }

  /* History Panel Styles */
  .history-panel {
    padding: 12px;
    animation: fadeIn 0.3s ease-out;
  }

  @keyframes fadeIn {
    from { opacity: 0; transform: translateY(-10px); }
    to { opacity: 1; transform: translateY(0); }
  }

  .history-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 16px;
  }

  .history-header h3 {
    margin: 0;
    font-size: 13px;
    text-transform: uppercase;
    letter-spacing: 0.1em;
    color: #f59e0b;
  }

  .close-history-btn {
    background: none;
    border: none;
    color: #6b7280;
    font-size: 20px;
    cursor: pointer;
    padding: 0;
    width: 24px;
    height: 24px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 4px;
  }

  .close-history-btn:hover {
    background: rgba(255, 255, 255, 0.1);
    color: #e5e7eb;
  }

  .stats-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 8px;
    margin-bottom: 16px;
  }

  .stat-card {
    background: rgba(255, 255, 255, 0.03);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 6px;
    padding: 12px;
    text-align: center;
  }

  .stat-card.active {
    border-color: rgba(245, 158, 11, 0.3);
    background: rgba(245, 158, 11, 0.05);
  }

  .stat-card.completed {
    border-color: rgba(34, 197, 94, 0.3);
    background: rgba(34, 197, 94, 0.05);
  }

  .stat-card.completion-rate {
    border-color: rgba(59, 130, 246, 0.3);
    background: rgba(59, 130, 246, 0.05);
  }

  .stat-value {
    font-size: 20px;
    font-weight: bold;
    color: #e5e7eb;
    margin-bottom: 4px;
  }

  .stat-label {
    font-size: 10px;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: #9ca3af;
  }

  .rewards-summary,
  .category-breakdown,
  .achievements-section {
    margin-bottom: 16px;
    padding-bottom: 16px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  }

  .rewards-summary h4,
  .category-breakdown h4,
  .achievements-section h4 {
    margin: 0 0 8px 0;
    font-size: 11px;
    text-transform: uppercase;
    letter-spacing: 0.1em;
    color: #9ca3af;
  }

  .reward-summary-row {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 4px 0;
    font-size: 13px;
  }

  .reward-amount {
    color: #22c55e;
    font-weight: bold;
  }

  .category-stat {
    display: flex;
    justify-content: space-between;
    padding: 4px 0;
    font-size: 12px;
  }

  .category-name {
    font-weight: bold;
  }

  .category-count {
    color: #9ca3af;
  }

  .achievements-list {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .achievement {
    display: flex;
    align-items: flex-start;
    gap: 8px;
    padding: 8px;
    border-radius: 6px;
    background: rgba(255, 255, 255, 0.03);
    border: 1px solid rgba(255, 255, 255, 0.1);
  }

  .achievement.unlocked {
    border-color: rgba(245, 158, 11, 0.3);
    background: rgba(245, 158, 11, 0.05);
  }

  .achievement.locked {
    opacity: 0.5;
  }

  .achievement-icon {
    font-size: 18px;
    flex-shrink: 0;
  }

  .achievement-details {
    flex: 1;
    min-width: 0;
  }

  .achievement-name {
    font-size: 12px;
    font-weight: bold;
    color: #e5e7eb;
    margin-bottom: 2px;
  }

  .achievement-desc {
    font-size: 10px;
    color: #9ca3af;
    line-height: 1.3;
  }

  .locked-achievements {
    margin-top: 8px;
  }

  .locked-achievements summary {
    cursor: pointer;
    font-size: 11px;
    color: #9ca3af;
    padding: 6px 0;
    user-select: none;
  }

  .locked-achievements summary:hover {
    color: #e5e7eb;
  }

  .locked-achievements .achievements-list {
    margin-top: 6px;
  }

  .back-to-quests-btn {
    width: 100%;
    padding: 10px;
    background: rgba(245, 158, 11, 0.1);
    border: 1px solid rgba(245, 158, 11, 0.3);
    color: #f59e0b;
    border-radius: 6px;
    font-size: 12px;
    font-weight: bold;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    cursor: pointer;
    font-family: inherit;
    transition: all 0.2s;
  }

  .back-to-quests-btn:hover {
    background: rgba(245, 158, 11, 0.2);
    border-color: rgba(245, 158, 11, 0.5);
  }
</style>
