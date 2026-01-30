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
</style>

<script>
  import MediaQuery from "svelte-media-query";
  import { fade } from "svelte/transition";
  import moment from "moment";

  export let name;
  export let description;
  export let created;
  export let level;
  export let xp;

 

  function formattedDate() {
    return moment(created).format("MMMM Do YYYY, h:mm:ss a");
  }

  function avatar() {
    let num = 1 + Math.abs((name || "").hashCode()%12)
    return "/img/avatars/" + num + "p.png";
  }
</script>

<div class="card cyan darken-3 hoverable characterCard center-align">
  <div class="card-content white-text">
    <img src="{avatar()}" alt="" class="circle img" />
    <span class="card-title">{name}</span>
    <span class="card-content">LVL {level} - XP {xp}</span>

  </div>
  <div class="card-content">
    <p>{description}</p>
    <p>{formattedDate()}</p>
  </div>

</div>
