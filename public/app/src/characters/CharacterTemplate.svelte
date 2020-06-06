<style>
  .img {
    width: 64px;
  }
  .characterCard {
    width: 250px;
  }

  td {
    padding: 2px;
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
  export let templateId;
  export let attributes;
  export let callback;

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
    <span class="card-title">{name}</span>
    <span>{description}</span>

  </div>
  <div class="card-content">
    <table>
      <tbody>
        {#each attributes as attribute}
          <tr>
            <td>{attribute.name}</td>
            <td class="right">{attribute.value}</td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
  <div class="card-content">
    <button class="btn green" on:click="{() => {
      callback(templateId)
    }}">
      Choose
    </button>
  </div>

</div>
