<script>
  export let store = null;
  export let sendMessage = null;

  let quests = [];
  let expandedQuest = null;
  let pinnedQuests = [];

  // Filtering and sorting
  let filterCategory = 'all'; // all, main, side, daily
  let showCompleted = true;
  let showAbandoned = false;
  let sortBy = 'status'; // status, name, level, category
  let searchQuery = '';
  let showHistory = false;

  // Subscribe to store for reactive updates
  $: if (store) {
    quests = $store.quests || [];
    // Load pinned quests from localStorage
    if (typeof localStorage !== 'undefined') {
      const saved = localStorage.getItem('pinnedQuests');
      pinnedQuests = saved ? JSON.parse(saved) : [];
    }
  }

  // Filter and sort quests
  $: filteredQuests = quests.filter((q) => {
    // Status filter
    if (!showCompleted && q.status === 'completed') return false;
    if (!showAbandoned && q.status === 'abandoned') return false;

    // Category filter
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
        const statusOrder = { active: 0, completed: 1, abandoned: 2, failed: 3 };
        return (statusOrder[a.status] || 99) - (statusOrder[b.status] || 99);
    }
  });

  // Separate by pinned status
  $: pinnedQuestList = sortedQuests.filter((q) => pinnedQuests.includes(q.questId) && q.status === 'active');
  $: unpinnedActiveQuests = sortedQuests.filter((q) => !pinnedQuests.includes(q.questId) && q.status === 'active');
  $: completedQuests = sortedQuests.filter((q) => q.status === "completed");
  $: abandonedQuests = sortedQuests.filter((q) => q.status === "abandoned");
  $: failedQuests = sortedQuests.filter((q) => q.status === "failed");

  function toggleQuest(questId) {
    expandedQuest = expandedQuest === questId ? null : questId;
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
      // Limit to 5 pinned quests
      if (pinnedQuests.length >= 5) {
        alert('You can only pin up to 5 quests at once.');
        return;
      }
      pinnedQuests = [...pinnedQuests, questId];
    }

    // Save to localStorage
    if (typeof localStorage !== 'undefined') {
      localStorage.setItem('pinnedQuests', JSON.stringify(pinnedQuests));
    }
  }

  function isPinned(questId) {
    return pinnedQuests.includes(questId);
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
</script>

<div class="questlog-widget game-panel">
  <div class="questlog-header">
    <div class="header-title-row">
      <h2>Quest Log</h2>
      <button
        class="history-btn"
        on:click={() => showHistory = !showHistory}
        title="View Quest History"
      >
        📊
      </button>
    </div>
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

  <div class="questlog-content">
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
          <div class="quest-entry pinned" class:expanded={expandedQuest === quest.questId}>
            <button
              class="quest-name"
              on:click={() => toggleQuest(quest.questId)}
            >
              <span class="quest-indicator active pinned"></span>
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
                  <div class="objectives-header">Objectives:</div>
                  {#each quest.objectives || [] as obj}
                    <div class="objective" class:completed={obj.completed}>
                      <span class="check">{obj.completed ? "[x]" : "[ ]"}</span>
                      <span class="obj-text">{obj.description}</span>
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
                    class="pin-btn pinned"
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
          <div class="quest-entry" class:expanded={expandedQuest === quest.questId}>
            <button
              class="quest-name"
              on:click={() => toggleQuest(quest.questId)}
            >
              <span class="quest-indicator active"></span>
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
                  <div class="objectives-header">Objectives:</div>
                  {#each quest.objectives || [] as obj}
                    <div class="objective" class:completed={obj.completed}>
                      <span class="check">{obj.completed ? "[x]" : "[ ]"}</span>
                      <span class="obj-text">{obj.description}</span>
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
                      <span class="obj-text">{obj.description}</span>
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

    {#if showAbandoned && abandonedQuests.length > 0}
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

<style>
  /* Base panel styling from global .game-panel in themes.css.
     Override padding/overflow so the flex layout works correctly. */
  .questlog-widget {
    display: flex;
    flex-direction: column;
    overflow: hidden;
    padding: 0;
  }

  .questlog-header {
    padding: 0.8em 1em;
    background: var(--panel-header-bg);
    border-bottom: 1px solid var(--panel-header-border);
    display: flex;
    flex-direction: column;
    gap: 0.5em;
    flex-shrink: 0;
  }

  .header-title-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .questlog-header h2 {
    margin: 0;
    font-family: var(--font-display);
    font-size: var(--text-header);
    text-transform: uppercase;
    letter-spacing: 0.1em;
    color: var(--panel-header-color);
    text-shadow: var(--panel-header-shadow);
  }

  .history-btn {
    background: var(--btn-bg);
    border: 1px solid var(--btn-border);
    color: var(--accent-primary);
    padding: 6px 10px;
    border-radius: 4px;
    cursor: pointer;
    font-size: 16px;
    transition: all 0.2s;
    line-height: 1;
  }

  .history-btn:hover {
    background: var(--btn-hover-bg);
    border-color: var(--btn-hover-border);
  }

  .search-bar {
    position: relative;
    display: flex;
    align-items: center;
  }

  .search-input {
    flex: 1;
    background: var(--input-bg);
    border: 1px solid var(--input-border);
    color: var(--text-primary);
    padding: 7px 32px 7px 10px;
    border-radius: 4px;
    font-size: var(--text-sm);
    font-family: inherit;
  }

  .search-input:focus {
    outline: none;
    background: var(--panel-inner-hover);
    border-color: var(--input-focus-border);
  }

  .search-input::placeholder {
    color: var(--text-dim);
  }

  .clear-search-btn {
    position: absolute;
    right: 6px;
    background: none;
    border: none;
    color: var(--text-dim);
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
    background: var(--panel-inner-hover);
    color: var(--text-primary);
  }

  .filter-controls {
    display: flex;
    gap: 6px;
  }

  .filter-select {
    flex: 1;
    background: var(--input-bg);
    border: 1px solid var(--input-border);
    color: var(--text-primary);
    padding: 5px 8px;
    border-radius: 4px;
    font-size: var(--text-xs);
    font-family: inherit;
    cursor: pointer;
  }

  .filter-select:hover {
    background: var(--panel-inner-hover);
  }

  .filter-select option {
    background: #1f2937;
    color: var(--text-primary);
  }

  .toggle-controls {
    display: flex;
    gap: 12px;
    font-size: var(--text-xs);
  }

  .toggle-label {
    display: flex;
    align-items: center;
    gap: 4px;
    cursor: pointer;
    color: var(--text-secondary);
  }

  .toggle-label:hover {
    color: var(--text-primary);
  }

  .toggle-label input[type="checkbox"] {
    cursor: pointer;
  }

  .toggle-label span {
    user-select: none;
  }

  .questlog-content {
    flex: 1;
    min-height: 0;
    overflow-y: auto;
    padding: 0.75em;
    display: flex;
    flex-direction: column;
    scrollbar-width: thin;
    scrollbar-color: var(--scrollbar-thumb) var(--scrollbar-track);
  }

  .questlog-content::-webkit-scrollbar {
    width: 6px;
  }
  .questlog-content::-webkit-scrollbar-track {
    background: var(--scrollbar-track);
  }
  .questlog-content::-webkit-scrollbar-thumb {
    background: var(--scrollbar-thumb);
    border-radius: 3px;
  }

  .empty-state {
    color: var(--text-dim);
    text-align: center;
    padding: 2em 1em;
    font-style: italic;
    font-size: var(--text-sm);
  }

  .quest-section {
    margin-bottom: 1em;
    flex-shrink: 0;
  }

  .section-title {
    font-family: var(--font-display);
    font-size: var(--text-xs);
    text-transform: uppercase;
    letter-spacing: 0.15em;
    color: var(--text-secondary);
    margin: 0 0 0.5em 0;
    padding-bottom: 0.3em;
    border-bottom: 1px solid var(--divider-color);
  }

  .quest-entry {
    display: flex;
    flex-direction: column;
    margin-bottom: 4px;
    border-radius: 6px;
  }

  .quest-entry.expanded {
    background: var(--panel-inner-bg);
  }

  .quest-name {
    display: flex;
    align-items: center;
    gap: 8px;
    width: 100%;
    padding: 0.5em 0.6em;
    background: none;
    border: none;
    color: var(--text-primary);
    cursor: pointer;
    font-size: var(--text-sm);
    text-align: left;
    border-radius: 6px;
    font-family: inherit;
  }

  .quest-name:hover {
    background: var(--panel-inner-hover);
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
    padding: 2px 7px;
    border-radius: 3px;
    font-size: var(--text-xs);
    font-weight: bold;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .level-badge {
    background: var(--panel-inner-bg);
    color: var(--text-primary);
    border: 1px solid var(--panel-inner-border);
  }

  .category-badge {
    color: #ffffff;
  }

  .quest-indicator {
    width: 9px;
    height: 9px;
    border-radius: 50%;
    flex-shrink: 0;
  }

  .quest-indicator.active {
    background: var(--accent-primary);
    box-shadow: 0 0 6px var(--accent-glow);
  }

  .quest-indicator.active.pinned {
    background: linear-gradient(135deg, #f59e0b 0%, #fbbf24 100%);
    box-shadow: 0 0 8px rgba(245, 158, 11, 0.8);
    animation: pulse 2s ease-in-out infinite;
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
    background: var(--color-success);
  }

  .quest-indicator.abandoned {
    background: var(--text-dim);
  }

  .quest-indicator.failed {
    background: var(--color-danger);
  }

  .quest-entry.pinned {
    background: var(--accent-subtle);
    border: 1px solid var(--btn-border);
    border-radius: 6px;
    margin-bottom: 8px;
  }

  .expand-icon {
    margin-left: auto;
    color: var(--text-dim);
    font-size: var(--text-xs);
  }

  .quest-details {
    padding: 4px 10px 12px 26px;
  }

  .quest-description {
    color: var(--text-secondary);
    font-size: var(--text-sm);
    margin: 0 0 8px 0;
    line-height: 1.5;
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
    font-size: var(--text-sm);
    color: var(--text-primary);
  }

  .objective.completed {
    color: var(--color-success);
  }

  .check {
    font-family: monospace;
    flex-shrink: 0;
    color: var(--text-dim);
  }

  .objective.completed .check {
    color: var(--color-success);
  }

  .quest-progress {
    margin-left: auto;
    color: var(--text-dim);
    font-size: var(--text-xs);
  }

  .quest-entry.completed .quest-name .quest-title {
    color: var(--text-secondary);
  }

  .quest-entry.abandoned .quest-name,
  .quest-entry.failed .quest-name {
    color: var(--text-dim);
  }

  .objectives-header,
  .rewards-header {
    font-size: var(--text-xs);
    font-weight: bold;
    color: var(--text-secondary);
    margin-bottom: 6px;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .rewards {
    margin-top: 12px;
    padding-top: 8px;
    border-top: 1px solid var(--divider-color);
  }

  .reward-item {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: var(--text-sm);
    color: var(--text-primary);
    padding: 2px 0;
  }

  .reward-item.earned {
    color: var(--color-success);
  }

  .reward-icon {
    font-size: 1.1em;
    flex-shrink: 0;
  }

  .reward-text {
    flex: 1;
  }

  .quest-actions {
    margin-top: 12px;
    padding-top: 8px;
    border-top: 1px solid var(--divider-color);
    display: flex;
    gap: 8px;
  }

  .abandon-btn {
    padding: 6px 14px;
    background: rgba(239, 68, 68, 0.1);
    border: 1px solid rgba(239, 68, 68, 0.3);
    color: #ef4444;
    border-radius: 4px;
    font-size: var(--text-xs);
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
    padding: 6px 14px;
    background: var(--btn-bg);
    border: 1px solid var(--btn-border);
    color: var(--accent-primary);
    border-radius: 4px;
    font-size: var(--text-xs);
    cursor: pointer;
    font-family: inherit;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    font-weight: bold;
    transition: all 0.2s;
  }

  .pin-btn:hover {
    background: var(--btn-hover-bg);
    border-color: var(--btn-hover-border);
  }

  .pin-btn.pinned {
    background: var(--btn-hover-bg);
    border-color: var(--btn-hover-border);
  }

  .pin-btn.pinned:hover {
    background: rgba(239, 68, 68, 0.1);
    border-color: rgba(239, 68, 68, 0.3);
    color: #ef4444;
  }

  .quest-timestamp {
    margin-top: 8px;
    font-size: var(--text-xs);
    color: var(--text-dim);
    font-style: italic;
    text-align: right;
  }

  /* History Panel Styles */
  .history-panel {
    padding: 0.75em;
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
    margin-bottom: 1em;
  }

  .history-header h3 {
    margin: 0;
    font-family: var(--font-display);
    font-size: var(--text-sm);
    text-transform: uppercase;
    letter-spacing: 0.1em;
    color: var(--panel-header-color);
  }

  .close-history-btn {
    background: none;
    border: none;
    color: var(--text-dim);
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
    background: var(--panel-inner-hover);
    color: var(--text-primary);
  }

  .stats-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 8px;
    margin-bottom: 1em;
  }

  .stat-card {
    background: var(--panel-inner-bg);
    border: 1px solid var(--panel-inner-border);
    border-radius: 6px;
    padding: 0.75em;
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
    font-size: var(--text-xl);
    font-weight: bold;
    color: var(--text-primary);
    margin-bottom: 4px;
  }

  .stat-label {
    font-size: var(--text-xs);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--text-secondary);
  }

  .rewards-summary,
  .category-breakdown,
  .achievements-section {
    margin-bottom: 1em;
    padding-bottom: 1em;
    border-bottom: 1px solid var(--divider-color);
  }

  .rewards-summary h4,
  .category-breakdown h4,
  .achievements-section h4 {
    margin: 0 0 0.5em 0;
    font-size: var(--text-xs);
    text-transform: uppercase;
    letter-spacing: 0.1em;
    color: var(--text-secondary);
  }

  .reward-summary-row {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 4px 0;
    font-size: var(--text-sm);
  }

  .reward-amount {
    color: var(--color-success);
    font-weight: bold;
  }

  .category-stat {
    display: flex;
    justify-content: space-between;
    padding: 4px 0;
    font-size: var(--text-sm);
  }

  .category-name {
    font-weight: bold;
  }

  .category-count {
    color: var(--text-secondary);
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
    padding: 0.5em;
    border-radius: 6px;
    background: var(--panel-inner-bg);
    border: 1px solid var(--panel-inner-border);
  }

  .achievement.unlocked {
    border-color: var(--btn-border);
    background: var(--accent-subtle);
  }

  .achievement.locked {
    opacity: 0.5;
  }

  .achievement-icon {
    font-size: 1.2em;
    flex-shrink: 0;
  }

  .achievement-details {
    flex: 1;
    min-width: 0;
  }

  .achievement-name {
    font-size: var(--text-sm);
    font-weight: bold;
    color: var(--text-primary);
    margin-bottom: 2px;
  }

  .achievement-desc {
    font-size: var(--text-xs);
    color: var(--text-secondary);
    line-height: 1.3;
  }

  .locked-achievements {
    margin-top: 8px;
  }

  .locked-achievements summary {
    cursor: pointer;
    font-size: var(--text-xs);
    color: var(--text-secondary);
    padding: 6px 0;
    user-select: none;
  }

  .locked-achievements summary:hover {
    color: var(--text-primary);
  }

  .locked-achievements .achievements-list {
    margin-top: 6px;
  }

  .back-to-quests-btn {
    width: 100%;
    padding: 0.6em;
    background: var(--btn-bg);
    border: 1px solid var(--btn-border);
    color: var(--accent-primary);
    border-radius: 6px;
    font-size: var(--text-sm);
    font-weight: bold;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    cursor: pointer;
    font-family: inherit;
    transition: all 0.2s;
  }

  .back-to-quests-btn:hover {
    background: var(--btn-hover-bg);
    border-color: var(--btn-hover-border);
  }
</style>
