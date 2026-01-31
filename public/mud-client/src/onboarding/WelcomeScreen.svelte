<style>
  .welcome-screen {
    position: fixed;
    inset: 0;
    background: #0a0e14;
    display: flex;
    align-items: center;
    justify-content: center;
    overflow: hidden;
  }

  .bg-image {
    position: absolute;
    inset: 0;
    background-image: url('img/bg/oldtown-griphon.png');
    background-size: cover;
    background-position: center;
    image-rendering: pixelated;
    opacity: 0.08;
    filter: blur(4px) saturate(0.3) brightness(0.7);
  }

  .bg-gradient {
    position: absolute;
    inset: 0;
    background: radial-gradient(ellipse 70% 60% at 50% 45%, transparent 0%, #0a0e14 100%);
  }

  .card {
    position: relative;
    z-index: 3;
    background: rgba(0, 0, 0, 0.75);
    backdrop-filter: blur(12px);
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 12px;
    padding: 3rem 3.5rem;
    max-width: 440px;
    width: 90vw;
    text-align: center;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 1.2rem;
    opacity: 0;
    animation: fadeSlideIn 0.5s ease forwards;
    animation-delay: 0.15s;
  }

  @keyframes fadeSlideIn {
    from {
      opacity: 0;
      transform: translateY(12px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .icon {
    color: #f59e0b;
    font-size: 2.4rem;
    margin-bottom: 0.25rem;
  }

  .title {
    font-family: 'Cinzel', serif;
    font-size: 1.6rem;
    font-weight: 600;
    color: #e5e7eb;
    letter-spacing: 0.06em;
    margin: 0;
    line-height: 1.3;
  }

  .subtitle {
    font-size: 0.9rem;
    color: #9ca3af;
    line-height: 1.5;
    max-width: 340px;
  }

  .divider {
    width: 60px;
    height: 1px;
    background: rgba(255, 255, 255, 0.1);
  }

  .buttons {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    width: 100%;
    max-width: 280px;
    margin-top: 0.5rem;
  }

  .btn-welcome {
    font-size: 0.85rem;
    font-weight: 500;
    letter-spacing: 0.03em;
    padding: 0.75rem 1.5rem;
    border-radius: 6px;
    cursor: pointer;
    transition: all 0.2s ease;
    text-decoration: none;
    width: 100%;
    box-sizing: border-box;
  }

  .btn-welcome.primary {
    border: none;
    color: #fff;
    background: #16a34a;
  }

  .btn-welcome.primary:hover {
    background: #15803d;
    box-shadow: 0 4px 12px rgba(22, 163, 74, 0.3);
    transform: translateY(-1px);
  }

  .btn-welcome.secondary {
    border: 1px solid rgba(255, 255, 255, 0.12);
    color: #d1d5db;
    background: transparent;
  }

  .btn-welcome.secondary:hover {
    border-color: rgba(255, 255, 255, 0.25);
    background: rgba(255, 255, 255, 0.04);
    color: #e5e7eb;
  }

  .error-banner {
    font-size: 0.8rem;
    color: #f87171;
    background: rgba(248, 113, 113, 0.1);
    border: 1px solid rgba(248, 113, 113, 0.2);
    padding: 0.5rem 1rem;
    border-radius: 6px;
    width: 100%;
    max-width: 280px;
    text-align: center;
  }
</style>

<script>
  export let login;
  export let serverName = "Tales";
  export let authError = null;

  function handleSignup() {
    login(null, { screen_hint: "signup" });
  }

  function handleLogin() {
    login();
  }
</script>

<div class="welcome-screen">
  <div class="bg-image"></div>
  <div class="bg-gradient"></div>

  <div class="card">
    <i class="material-icons icon">auto_stories</i>

    <h1 class="title">{serverName}</h1>

    <p class="subtitle">
      A multiplayer text adventure. Create a character, explore the world, and forge your own legend.
    </p>

    <div class="divider"></div>

    {#if authError}
      <div class="error-banner">
        Authentication failed. Please try again.
      </div>
    {/if}

    <div class="buttons">
      <button class="btn-welcome primary" on:click={handleSignup}>
        Sign Up
      </button>
      <button class="btn-welcome secondary" on:click={handleLogin}>
        Log In
      </button>
    </div>
  </div>
</div>
