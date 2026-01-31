<style>
  .img {
    width: 96px;
    height: 96px;
    margin: 1em;
    image-rendering: pixelated;
  }
  .characterCard {
    width: 250px;
  }

  td {
    padding: 2px;
  }
</style>

<script>
  import moment from "moment";

  export let name;
  export let description;
  // Props passed from parent but not displayed in template
  // svelte-ignore unused-export-let
  export let created;
  // svelte-ignore unused-export-let
  export let level;
  // svelte-ignore unused-export-let
  export let xp;
  export let templateId;
  export let attributes;
  export let callback;

  function formattedDate() {
    return moment(created).format("MMMM Do YYYY, h:mm:ss a");
  }

  function avatar() {
    let num = 1 + Math.abs((name || "").hashCode()%12)
    return "img/avatars/" + num + "p.png";
  }
</script>

<div class="card cyan darken-3 hoverable characterCard center-align">
  <div class="card-content white-text">
    <img src="{avatar()}" alt="" class="circle img" />
    <span class="card-title">{name}</span>
    <span>{description}</span>

  </div>
  <div class="card-content">

    {#if attributes}
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

    {/if}
  </div>
  <div class="card-content">
    <button
      class="btn green"
      on:click="{() => {
        callback(templateId);
      }}"
    >
      Choose
    </button>
  </div>

</div>
