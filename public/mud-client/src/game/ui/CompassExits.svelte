<script>
  export let cardinalExits = []; // [{name: 'north', available: true}, ...]
  export let sendMessage;

  // Debug: log when cardinalExits changes
  $: console.log("CompassExits received:", cardinalExits);

  // Reactive availability for each direction
  $: northAvailable = cardinalExits.find(e => e.name === 'north')?.available || false;
  $: southAvailable = cardinalExits.find(e => e.name === 'south')?.available || false;
  $: eastAvailable = cardinalExits.find(e => e.name === 'east')?.available || false;
  $: westAvailable = cardinalExits.find(e => e.name === 'west')?.available || false;

  function handleExit(direction) {
    sendMessage(direction);
  }
</script>

<style>
  .compass-wrapper {
    position: relative;
    width: 88px;
    height: 88px;
  }

  /* Circular background behind buttons */
  .compass-circle {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    width: 72px;
    height: 72px;
    border-radius: 50%;
    background: radial-gradient(
      circle,
      rgba(249, 115, 22, 0.15) 0%,
      rgba(249, 115, 22, 0.08) 60%,
      rgba(249, 115, 22, 0.02) 100%
    );
    border: 1px solid rgba(249, 115, 22, 0.25);
    box-shadow: 0 0 20px rgba(249, 115, 22, 0.1);
  }

  .compass {
    position: relative;
    display: grid;
    grid-template-columns: repeat(3, 28px);
    grid-template-rows: repeat(3, 28px);
    gap: 2px;
    z-index: 1;
  }

  .compass-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(249, 115, 22, 0.3);
    border: 1px solid rgba(249, 115, 22, 0.5);
    border-radius: 4px;
    color: #fdba74;
    cursor: pointer;
    transition: all 0.15s ease;
    padding: 0;
  }

  .compass-btn:hover:not(:disabled) {
    background: rgba(249, 115, 22, 0.5);
    border-color: rgba(249, 115, 22, 0.8);
    transform: scale(1.1);
  }

  .compass-btn:active:not(:disabled) {
    transform: scale(0.95);
  }

  .compass-btn:disabled {
    opacity: 0.3;
    cursor: not-allowed;
    background: rgba(100, 100, 100, 0.15);
    border-color: rgba(100, 100, 100, 0.25);
    color: #666;
  }

  .compass-btn i {
    font-size: 18px;
  }

  /* Grid positioning */
  .compass-north { grid-area: 1 / 2; }
  .compass-west { grid-area: 2 / 1; }
  .compass-center { grid-area: 2 / 2; }
  .compass-east { grid-area: 2 / 3; }
  .compass-south { grid-area: 3 / 2; }

  /* Center - transparent to show circle behind */
  .compass-center {
    background: transparent;
    border: none;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: default;
  }

  .compass-center i {
    font-size: 14px;
    color: rgba(249, 115, 22, 0.5);
  }
</style>

<div class="compass-wrapper">
  <div class="compass-circle"></div>
  <div class="compass">
  <button
    class="compass-btn compass-north"
    disabled={!northAvailable}
    on:click={() => handleExit('north')}
    title="North"
  >
    <i class="material-icons">north</i>
  </button>

  <button
    class="compass-btn compass-west"
    disabled={!westAvailable}
    on:click={() => handleExit('west')}
    title="West"
  >
    <i class="material-icons">west</i>
  </button>

  <div class="compass-center">
    <i class="material-icons">my_location</i>
  </div>

  <button
    class="compass-btn compass-east"
    disabled={!eastAvailable}
    on:click={() => handleExit('east')}
    title="East"
  >
    <i class="material-icons">east</i>
  </button>

  <button
    class="compass-btn compass-south"
    disabled={!southAvailable}
    on:click={() => handleExit('south')}
    title="South"
  >
    <i class="material-icons">south</i>
  </button>
  </div>
</div>
