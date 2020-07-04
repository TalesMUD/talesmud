<style>
  .img {
    width: 96px;
    height: 96px;
    margin: 1em;
    image-rendering: pixelated;
  }
  .characterCard {
    width: 250px;
    height: 200px;
    margin: 2em;
  }
  .card-title {
    font-size: 10px;
  }
</style>

<script>
  import MediaQuery from "svelte-media-query";
  import { fade } from "svelte/transition";
  import moment from "moment";

  export let name;
  export let level;
  export let cclass;
  export let xp;

  function formattedDate() {
    return moment(created).format("MMMM Do YYYY, h:mm:ss a");
  }

  function avatar() {
    let num = 1 + Math.abs(name.hashCode() % 12);
    return "img/avatars/" + num + "p.png";
  }
</script>

<div class="card blue-grey darken-2 hoverable characterCard center-align" style="margin-top:50px;">
  <div class="card-content white-text">

    <img src="{avatar()}" alt="" style="margin-top:-70px;" class="circle img z-depth-2" />

    <span class="card-title">{name}</span>
    <span class="card-content">LVL {level} - {cclass.name}</span><br>
    <span class="card-content">{xp} XP</span>
  </div>
</div>
