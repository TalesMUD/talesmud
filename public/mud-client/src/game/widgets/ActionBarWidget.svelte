<script>
  import MUDXPlus from '../MUDXPlus.svelte';

  export let store;
  export let sendMessage;
  export let term = null;
</script>

<style>
  .actionbar-widget {
    height: 100%;
    display: flex;
    flex-direction: column;
  }
</style>

<div class="actionbar-widget">
  <MUDXPlus {store} {sendMessage} {term} />
</div>
