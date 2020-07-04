<style>
  h2 {
    overflow: hidden;
    position: relative;
    width: 16px;
    height: 16px;
    
  }
  h2 img {
    position: relative;
  }
  h2.weapon img {
    top: 0px;
    left: -32px;
  }
  h2.shield img {
    top: -64;
    left: -128px;
  }
</style>

<script>
  export let item = "weapon";
</script>

<h2 class="{item}">
  <img src="img/sword.png" alt="" />
</h2>
