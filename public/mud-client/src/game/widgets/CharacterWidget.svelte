<script>
  // Placeholder widget - will be connected to server data later
  const mockCharacter = {
    name: 'Legolas Longstride',
    race: 'Elf',
    class: 'Ranger',
    level: 12,
    xp: 4250,
    xpToNext: 5000,
    gold: 1847,
    currentHp: 85,
    maxHp: 100,
    currentMp: 45,
    maxMp: 60,
    attributes: {
      strength: 14,
      dexterity: 18,
      constitution: 12,
      intelligence: 13,
      wisdom: 15,
      charisma: 11
    }
  };
</script>

<style>
  .character-widget {
    background: rgba(0, 0, 0, 0.85);
    backdrop-filter: blur(12px);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 12px;
    padding: 1em;
    height: 100%;
    overflow-y: auto;
    color: #e5e7eb;
  }

  .widget-header {
    display: flex;
    align-items: center;
    gap: 0.5em;
    margin-bottom: 1em;
    padding-bottom: 0.75em;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  }

  .widget-header i {
    color: #f59e0b;
  }

  .widget-title {
    font-size: 1em;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .char-name {
    font-size: 1.1em;
    font-weight: 600;
    margin-bottom: 0.25em;
  }

  .char-info {
    font-size: 0.85em;
    color: #9ca3af;
    margin-bottom: 1em;
  }

  .stat-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 0.5em;
    font-size: 0.9em;
  }

  .stat-label {
    color: #9ca3af;
  }

  .stat-value {
    font-weight: 600;
  }

  .stat-value.gold {
    color: #fbbf24;
  }

  .bar-container {
    margin-bottom: 0.75em;
  }

  .bar-label {
    display: flex;
    justify-content: space-between;
    font-size: 0.8em;
    margin-bottom: 0.25em;
  }

  .bar {
    height: 6px;
    background: rgba(255, 255, 255, 0.1);
    border-radius: 3px;
    overflow: hidden;
  }

  .bar-fill {
    height: 100%;
    border-radius: 3px;
    transition: width 0.3s ease;
  }

  .bar-fill.hp {
    background: linear-gradient(90deg, #22c55e, #4ade80);
  }

  .bar-fill.mp {
    background: linear-gradient(90deg, #3b82f6, #60a5fa);
  }

  .bar-fill.xp {
    background: linear-gradient(90deg, #a855f7, #c084fc);
  }

  .section-title {
    font-size: 0.75em;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    color: #6b7280;
    margin: 1em 0 0.5em 0;
  }

  .attributes-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 0.4em;
    font-size: 0.85em;
  }

  .attr-item {
    display: flex;
    justify-content: space-between;
    padding: 0.3em 0.5em;
    background: rgba(255, 255, 255, 0.05);
    border-radius: 4px;
  }

  .attr-name {
    color: #9ca3af;
    text-transform: uppercase;
    font-size: 0.75em;
  }

  .attr-value {
    font-weight: 600;
    color: #f59e0b;
  }

  .placeholder-notice {
    font-size: 0.7em;
    color: #6b7280;
    text-align: center;
    margin-top: 1em;
    padding-top: 0.75em;
    border-top: 1px solid rgba(255, 255, 255, 0.1);
  }
</style>

<div class="character-widget">
  <div class="widget-header">
    <i class="material-icons">person</i>
    <span class="widget-title">Character</span>
  </div>

  <div class="char-name">{mockCharacter.name}</div>
  <div class="char-info">
    Level {mockCharacter.level} {mockCharacter.race} {mockCharacter.class}
  </div>

  <div class="bar-container">
    <div class="bar-label">
      <span>HP</span>
      <span>{mockCharacter.currentHp}/{mockCharacter.maxHp}</span>
    </div>
    <div class="bar">
      <div
        class="bar-fill hp"
        style="width: {(mockCharacter.currentHp / mockCharacter.maxHp) * 100}%"
      ></div>
    </div>
  </div>

  <div class="bar-container">
    <div class="bar-label">
      <span>MP</span>
      <span>{mockCharacter.currentMp}/{mockCharacter.maxMp}</span>
    </div>
    <div class="bar">
      <div
        class="bar-fill mp"
        style="width: {(mockCharacter.currentMp / mockCharacter.maxMp) * 100}%"
      ></div>
    </div>
  </div>

  <div class="bar-container">
    <div class="bar-label">
      <span>XP</span>
      <span>{mockCharacter.xp}/{mockCharacter.xpToNext}</span>
    </div>
    <div class="bar">
      <div
        class="bar-fill xp"
        style="width: {(mockCharacter.xp / mockCharacter.xpToNext) * 100}%"
      ></div>
    </div>
  </div>

  <div class="stat-row">
    <span class="stat-label">Gold</span>
    <span class="stat-value gold">{mockCharacter.gold.toLocaleString()}</span>
  </div>

  <div class="section-title">Attributes</div>
  <div class="attributes-grid">
    {#each Object.entries(mockCharacter.attributes) as [name, value]}
      <div class="attr-item">
        <span class="attr-name">{name.slice(0, 3)}</span>
        <span class="attr-value">{value}</span>
      </div>
    {/each}
  </div>

  <div class="placeholder-notice">
    Placeholder data - server integration pending
  </div>
</div>
