<style>
  .img {
    width: 96px;
    height: 96px;
    margin: 1em;
    image-rendering: pixelated;
  }
  .characterCard {
    width: 250px;
    height: 300px;
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

<div class="card cyan darken-3 hoverable characterCard center-align">
  <div class="card-content white-text">

    <img src="{avatar()}" alt="" class="circle img z-depth-2" />

    <span class="card-title">{xp} XP</span>
    <span class="card-title">LVL {level} - {cclass.name}</span>
    <span class="card-content">{name}</span>
  </div>
</div>
