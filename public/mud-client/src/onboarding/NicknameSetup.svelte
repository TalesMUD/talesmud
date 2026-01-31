<style>
  .nickname-screen {
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
    background-image: url('img/bg/castle-1.png');
    background-size: cover;
    background-position: center;
    image-rendering: pixelated;
    opacity: 0.06;
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
    padding: 3rem;
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

  .avatar {
    width: 72px;
    height: 72px;
    border-radius: 50%;
    border: 2px solid rgba(255, 255, 255, 0.1);
    object-fit: cover;
  }

  .greeting {
    font-family: 'Cinzel', serif;
    font-size: 1.4rem;
    font-weight: 600;
    color: #e5e7eb;
    letter-spacing: 0.04em;
    margin: 0;
  }

  .description {
    font-size: 0.9rem;
    color: #9ca3af;
    line-height: 1.5;
    max-width: 340px;
  }

  .input-group {
    width: 100%;
    max-width: 300px;
    text-align: left;
  }

  .input-label {
    font-size: 0.75rem;
    font-weight: 500;
    color: #6b7280;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    margin-bottom: 0.5rem;
    display: block;
  }

  .nickname-input {
    width: 100%;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.12);
    outline: none;
    box-shadow: none;
    color: #e5e7eb;
    font-size: 1rem;
    padding: 0.7rem 0.9rem;
    border-radius: 6px;
    box-sizing: border-box;
    transition: border-color 0.2s ease;
  }

  .nickname-input:focus {
    border-color: rgba(255, 255, 255, 0.25);
    outline: none;
    box-shadow: none;
  }

  .nickname-input::placeholder {
    color: #4b5563;
  }

  .validation-msg {
    font-size: 0.75rem;
    color: #f87171;
    margin-top: 0.4rem;
    min-height: 1em;
  }

  .btn-continue {
    font-size: 0.85rem;
    font-weight: 500;
    padding: 0.75rem 2rem;
    border: none;
    color: #fff;
    background: #16a34a;
    cursor: pointer;
    transition: all 0.2s ease;
    border-radius: 6px;
    margin-top: 0.5rem;
  }

  .btn-continue:hover:not(:disabled) {
    background: #15803d;
    box-shadow: 0 4px 12px rgba(22, 163, 74, 0.3);
    transform: translateY(-1px);
  }

  .btn-continue:disabled {
    opacity: 0.3;
    cursor: not-allowed;
  }

  .error-banner {
    font-size: 0.8rem;
    color: #f87171;
    background: rgba(248, 113, 113, 0.1);
    border: 1px solid rgba(248, 113, 113, 0.2);
    padding: 0.5rem 1rem;
    border-radius: 6px;
    width: 100%;
    max-width: 300px;
    text-align: center;
  }

  .saving-text {
    font-size: 0.8rem;
    color: #6b7280;
  }
</style>

<script>
  import { updateUser } from "../api/user.js";

  export let authToken;
  export let userInfo = {};
  export let currentUser = null;
  export let onComplete;

  let nickname = "";
  let validationError = "";
  let saving = false;
  let apiError = "";

  // Pre-fill from Auth0 or backend data
  $: {
    if (!nickname && currentUser) {
      nickname = currentUser.nickname || currentUser.name || "";
    }
    if (!nickname && userInfo && userInfo.nickname) {
      nickname = userInfo.nickname || "";
    }
  }

  $: isValid = nickname.trim().length >= 2 && nickname.trim().length <= 30;

  function validate() {
    const trimmed = nickname.trim();
    if (trimmed.length < 2) {
      validationError = "Name must be at least 2 characters";
      return false;
    }
    if (trimmed.length > 30) {
      validationError = "Name must be 30 characters or less";
      return false;
    }
    validationError = "";
    return true;
  }

  async function handleSubmit() {
    if (!validate() || saving) return;

    saving = true;
    apiError = "";

    const updatedUser = {
      ...currentUser,
      nickname: nickname.trim(),
      name: nickname.trim(),
      isNewUser: false,
    };

    updateUser(
      authToken,
      updatedUser,
      (result) => {
        saving = false;
        onComplete(result);
      },
      (err) => {
        saving = false;
        apiError = "Failed to save. Please try again.";
        console.error("Failed to update user:", err);
      }
    );
  }

  function handleKeydown(e) {
    if (e.key === "Enter" && isValid && !saving) {
      handleSubmit();
    }
  }
</script>

<div class="nickname-screen">
  <div class="bg-image"></div>
  <div class="bg-gradient"></div>

  <div class="card">
    {#if userInfo && userInfo.picture}
      <img src={userInfo.picture} alt="" class="avatar" />
    {/if}

    <h2 class="greeting">Welcome</h2>

    <p class="description">
      Choose a display name for your account.
    </p>

    <div class="input-group">
      <label class="input-label" for="nickname-input">Display Name</label>
      <input
        id="nickname-input"
        class="nickname-input"
        type="text"
        bind:value={nickname}
        placeholder="Enter a name..."
        on:keydown={handleKeydown}
        on:input={() => { validationError = ""; apiError = ""; }}
        maxlength="30"
        autofocus
      />
      <div class="validation-msg">
        {validationError}
      </div>
    </div>

    {#if apiError}
      <div class="error-banner">{apiError}</div>
    {/if}

    {#if saving}
      <span class="saving-text">Saving...</span>
    {:else}
      <button
        class="btn-continue"
        on:click={handleSubmit}
        disabled={!isValid}
      >
        Continue
      </button>
    {/if}
  </div>
</div>
