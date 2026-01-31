<style>
  .loading-screen {
    position: fixed;
    inset: 0;
    background: #0a0e14;
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 9999;
  }

  .loading-content {
    text-align: center;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 1.5rem;
  }

  .loading-icon {
    color: #f59e0b;
    font-size: 2rem;
  }

  .loading-bar {
    width: 160px;
    height: 2px;
    background: rgba(255, 255, 255, 0.08);
    border-radius: 1px;
    overflow: hidden;
  }

  .loading-bar-fill {
    height: 100%;
    width: 40%;
    background: linear-gradient(90deg, transparent, #9ca3af, transparent);
    border-radius: 1px;
    animation: loadingSlide 1.5s ease-in-out infinite;
  }

  @keyframes loadingSlide {
    0% { transform: translateX(-100%); }
    100% { transform: translateX(350%); }
  }

  .loading-text {
    font-size: 0.8rem;
    color: #6b7280;
  }
</style>

<script>
</script>

<div class="loading-screen">
  <div class="loading-content">
    <i class="material-icons loading-icon">auto_stories</i>
    <div class="loading-bar">
      <div class="loading-bar-fill"></div>
    </div>
    <span class="loading-text">Loading...</span>
  </div>
</div>
