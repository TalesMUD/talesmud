<script>
  import { createEventDispatcher } from "svelte";

  export let open = false;
  export let item = null;

  const dispatch = createEventDispatcher();

  const close = () => {
    dispatch("close");
  };

  const handleKey = (event) => {
    if (!open) return;
    if (event.key === "Escape") {
      close();
    }
  };

  // Quality color mapping
  const qualityColors = {
    normal: "#9ca3af",
    magic: "#3b82f6",
    rare: "#eab308",
    legendary: "#f97316",
    mythic: "#a855f7",
  };

  $: qualityColor = item?.quality ? qualityColors[item.quality] || "#9ca3af" : "#9ca3af";
</script>

<svelte:window on:keydown={handleKey} />

{#if open && item}
  <div class="modal-backdrop" on:click|self={close}>
    <div class="modal-container">
      <div class="modal-header">
        <div class="header-text">
          <h2 style="color: {qualityColor}">{item.name || "Unnamed Item"}</h2>
          <p class="header-subtitle">
            {#if item.isTemplate}
              <span class="template-badge">Template</span>
            {:else if item.instanceSuffix}
              <span class="instance-badge">Instance</span>
              <span class="instance-id">#{item.instanceSuffix}</span>
            {/if}
            {#if item.type}
              <span class="type-badge">{item.type}</span>
            {/if}
            {#if item.subType}
              <span class="subtype-badge">{item.subType}</span>
            {/if}
          </p>
        </div>
        <button class="close-btn" type="button" on:click={close}>
          <span class="material-symbols-outlined">close</span>
        </button>
      </div>

      <div class="modal-content">
        {#if item.description}
          <div class="description-section">
            <p class="description">{item.description}</p>
          </div>
        {/if}

        <div class="stats-grid">
          {#if item.quality}
            <div class="stat-item">
              <span class="stat-label">Quality</span>
              <span class="stat-value" style="color: {qualityColor}">{item.quality}</span>
            </div>
          {/if}
          {#if item.level}
            <div class="stat-item">
              <span class="stat-label">Level</span>
              <span class="stat-value">{item.level}</span>
            </div>
          {/if}
          {#if item.slot && item.slot !== "inventory"}
            <div class="stat-item">
              <span class="stat-label">Slot</span>
              <span class="stat-value">{item.slot}</span>
            </div>
          {/if}
          {#if item.basePrice}
            <div class="stat-item">
              <span class="stat-label">Base Price</span>
              <span class="stat-value">{item.basePrice} gold</span>
            </div>
          {/if}
          {#if item.stackable}
            <div class="stat-item">
              <span class="stat-label">Stack</span>
              <span class="stat-value">{item.quantity || 1} / {item.maxStack || "unlimited"}</span>
            </div>
          {/if}
        </div>

        {#if item.attributes && Object.keys(item.attributes).length > 0}
          <div class="section">
            <h3 class="section-title">Attributes</h3>
            <div class="attributes-list">
              {#each Object.entries(item.attributes) as [key, value]}
                <div class="attribute-item">
                  <span class="attribute-key">{key}</span>
                  <span class="attribute-value">{value}</span>
                </div>
              {/each}
            </div>
          </div>
        {/if}

        {#if item.properties && Object.keys(item.properties).length > 0}
          <div class="section">
            <h3 class="section-title">Properties</h3>
            <div class="attributes-list">
              {#each Object.entries(item.properties) as [key, value]}
                <div class="attribute-item">
                  <span class="attribute-key">{key}</span>
                  <span class="attribute-value">{typeof value === 'object' ? JSON.stringify(value) : value}</span>
                </div>
              {/each}
            </div>
          </div>
        {/if}

        {#if item.tags && item.tags.length > 0}
          <div class="section">
            <h3 class="section-title">Tags</h3>
            <div class="tags-list">
              {#each item.tags as tag}
                <span class="tag">{tag}</span>
              {/each}
            </div>
          </div>
        {/if}

        <div class="meta-section">
          <div class="meta-item">
            <span class="meta-label">ID</span>
            <span class="meta-value mono">{item.id}</span>
          </div>
          {#if item.templateId}
            <div class="meta-item">
              <span class="meta-label">Template ID</span>
              <span class="meta-value mono">{item.templateId}</span>
            </div>
          {/if}
        </div>
      </div>

      <div class="modal-footer">
        <button class="btn-close" type="button" on:click={close}>
          Close
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .modal-backdrop {
    position: fixed;
    inset: 0;
    z-index: 100;
    display: flex;
    align-items: flex-start;
    justify-content: center;
    background: rgba(0, 0, 0, 0.7);
    backdrop-filter: blur(4px);
    padding: 40px 20px;
    overflow-y: auto;
  }

  .modal-container {
    width: 100%;
    max-width: 500px;
    background: #1e1e1e;
    border: 1px solid #3a3a3a;
    border-radius: 12px;
    box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
    overflow: hidden;
  }

  .modal-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    padding: 20px 24px;
    border-bottom: 1px solid #3a3a3a;
    background: #252525;
  }

  .header-text h2 {
    margin: 0;
    font-size: 18px;
    font-weight: 600;
  }

  .header-subtitle {
    display: flex;
    align-items: center;
    gap: 8px;
    margin: 6px 0 0;
    font-size: 12px;
    color: #888;
    flex-wrap: wrap;
  }

  .template-badge {
    padding: 2px 8px;
    background: rgba(168, 85, 247, 0.2);
    border: 1px solid rgba(168, 85, 247, 0.4);
    border-radius: 4px;
    color: #a855f7;
    font-size: 10px;
    font-weight: 600;
    text-transform: uppercase;
  }

  .instance-badge {
    padding: 2px 8px;
    background: rgba(0, 188, 212, 0.2);
    border: 1px solid rgba(0, 188, 212, 0.4);
    border-radius: 4px;
    color: #00bcd4;
    font-size: 10px;
    font-weight: 600;
    text-transform: uppercase;
  }

  .instance-id {
    font-family: monospace;
    color: #666;
  }

  .type-badge, .subtype-badge {
    padding: 2px 8px;
    background: rgba(100, 116, 139, 0.2);
    border-radius: 4px;
    color: #94a3b8;
    font-size: 10px;
    text-transform: capitalize;
  }

  .close-btn {
    background: transparent;
    border: none;
    color: #888;
    cursor: pointer;
    padding: 4px;
    border-radius: 4px;
    transition: all 0.2s ease;
  }

  .close-btn:hover {
    background: #333;
    color: #fff;
  }

  .modal-content {
    padding: 20px 24px;
    max-height: 60vh;
    overflow-y: auto;
  }

  .description-section {
    margin-bottom: 20px;
    padding-bottom: 16px;
    border-bottom: 1px solid #2a2a2a;
  }

  .description {
    margin: 0;
    font-size: 14px;
    line-height: 1.6;
    color: #ccc;
  }

  .stats-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 12px;
    margin-bottom: 20px;
  }

  .stat-item {
    display: flex;
    flex-direction: column;
    gap: 2px;
    padding: 10px;
    background: #252525;
    border-radius: 6px;
  }

  .stat-label {
    font-size: 10px;
    font-weight: 600;
    text-transform: uppercase;
    color: #666;
  }

  .stat-value {
    font-size: 14px;
    font-weight: 500;
    color: #fff;
    text-transform: capitalize;
  }

  .section {
    margin-bottom: 16px;
  }

  .section-title {
    font-size: 11px;
    font-weight: 600;
    text-transform: uppercase;
    color: #666;
    margin: 0 0 10px;
  }

  .attributes-list {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .attribute-item {
    display: flex;
    justify-content: space-between;
    padding: 8px 12px;
    background: #252525;
    border-radius: 4px;
  }

  .attribute-key {
    font-size: 12px;
    color: #888;
    text-transform: capitalize;
  }

  .attribute-value {
    font-size: 12px;
    font-weight: 500;
    color: #00bcd4;
  }

  .tags-list {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }

  .tag {
    padding: 4px 10px;
    background: rgba(100, 116, 139, 0.2);
    border-radius: 12px;
    font-size: 11px;
    color: #94a3b8;
  }

  .meta-section {
    margin-top: 20px;
    padding-top: 16px;
    border-top: 1px solid #2a2a2a;
  }

  .meta-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 6px 0;
  }

  .meta-label {
    font-size: 10px;
    font-weight: 600;
    text-transform: uppercase;
    color: #666;
  }

  .meta-value {
    font-size: 11px;
    color: #888;
  }

  .meta-value.mono {
    font-family: monospace;
    font-size: 10px;
  }

  .modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
    padding: 16px 24px;
    border-top: 1px solid #3a3a3a;
    background: #252525;
  }

  .btn-close {
    padding: 10px 24px;
    font-size: 13px;
    font-weight: 500;
    background: #00bcd4;
    border: none;
    border-radius: 6px;
    color: #fff;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .btn-close:hover {
    background: #00a5bb;
  }
</style>
