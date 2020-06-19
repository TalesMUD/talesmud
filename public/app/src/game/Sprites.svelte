<style>
  .container {
    width: 64px !important;
    height: 64px !important;
  }
  .weapon {
    width: 16px;
    height: 16px;
    transform: scale(2) translate(4px, 4px);
    background: url("img/items.png") -32px 0px;
    image-rendering: pixelated;
  }
  .shield {
    width: 16px;
    height: 16px;
    

    background: url("img/items.png") -64px -128px;
    image-rendering: pixelated;
  }
</style>

<script>
  export let item = "weapon";
</script>

<div class="container">
  <div class="{item}"></div>
</div>
