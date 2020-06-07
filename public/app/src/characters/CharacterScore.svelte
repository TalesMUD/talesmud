<style>
  .img {
    width: 64px;
  }
  .characterCard {
    width: 250px;
    height: 220px;
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
    return "https://robohash.org/" + name;
  }
</script>

<div class="card cyan darken-3 hoverable characterCard center-align">
  <div class="card-content white-text">

    <img src="{avatar()}" alt="" class="circle img" />

    <span class="card-title">{xp} XP</span>
    <span class="card-title">LVL {level} - {cclass.name}</span>
    <span class="card-content">{name}</span>
  </div>
</div>
