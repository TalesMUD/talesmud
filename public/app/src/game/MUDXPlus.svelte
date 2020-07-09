<style>
  .mudx {
    margin-top: 1em;
    margin-left: auto;
    margin-right: auto;
    margin-bottom: 1em;
    max-width: 640px;
  }

  .inventory {
    max-width: 600px;
    max-height: 250px;

    position: absolute;
    top: 20px;
    left: 100px;

    z-index: 1003;
    display: block;
    opacity: 1;
    overflow-y: auto;
    transform: scaleX(1) scaleY(1);
  }

  .inventory ul {
    border: none;
    margin: 0.2em;
  }

  .inventory ul li:last-child {
    border: 1px #ffffff55 solid;
  }

  .item {
    background: #000000cc;
    color: #fff;
    border: 1px #ffffff55 solid;
    margin: 0.2em;
    border-radius: 5px;
    justify-content: flex-start;
    padding: 5px;
  }

  .item-header {
    width: 200px;
    max-width: 200px;
    font-weight: bolder;
    font-size: 13px;
    word-wrap: break-word;
    display: inline-block;
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
  }
  li img {
    padding: 2px;
    margin-right: 0.5em;
    border-radius: 3px;
    background: #00000044;
    border: 1px solid #ffffff22;
    image-rendering: pixelated;
    height: 42px;
    width: 42px;
    float: left;
  }

  .actions {
    border: 1px solid blue;
    margin: 0;
    padding: 0;
  }

  .ul2 {
    display: flex;
    justify-content: flex-start;
    flex-direction: row;
    align-items: center;
    margin: 0;
    padding: 0;
    margin-bottom: 0.5em;
  }
  .ul2 li {
    display: inline-block;
    list-style: none;
    color: #eee;
  }
  .btncolor {
    background: #00000000;
    border: 1px solid #ffffff33;
  }

  .btn {
    border-radius: 0.5em;
    padding-left: 1em;
    padding-right: 1em;
    margin-left: 0.5em;
  }
</style>

<script>
  import Inventory from "./ui/Inventory.svelte";
  import Sprites from "./Sprites.svelte";
  import { writable } from "svelte/store";
  import { getClient } from "./Client";

  const showInventory = writable(false);
  const showExits = writable(false);

  const showActions = writable(false);
  const showSkills = writable(false);

  const hideAll = () => {
    showInventory.set(false);
    showExits.set(false);
    showActions.set(false);
    showSkills.set(false);
  };
  const toggleInventory = () => {
    let show = !$showInventory;
    hideAll();
    showInventory.set(show);
    /*
    var Modalelem = document.getElementById("inventoryModal");
    var instance = M.Modal.init(Modalelem);
     instance.open();
     */
  };
  const toggleExits = () => {
    let show = !$showExits;
    hideAll();
    showExits.set(show);
  };
  const toggleSkills = () => {
    let show = !$showSkills;
    hideAll();
    showSkills.set(show);
  };
  const toggleActions = () => {
    let show = !$showActions;
    hideAll();
    showActions.set(show);
  };
  export let store;
  export let term;
  export let sendMessage;

  const takeExit = (exit) => {
    console.log("Take exit " + exit);

    sendMessage(exit);
  };
</script>

<Inventory />

<div style="clear:both;"></div>

<div class="mudx">

  <button
    class="btn waves-effect waves-light btncolor darken-1 right"
    on:click="{() => toggleInventory()}"
  >
    <i class="material-icons">inbox</i>
  </button>

  <button
    class="btn waves-effect waves-light btncolor darken-1 left {$showExits === true ? "orange" : ""}"
    style="margin-left:0;"
    on:click="{() => toggleExits()}"
  >
    <i class="material-icons">exit_to_app</i>
  </button>

  {#if $showExits === true}
    <ul class="ul2 left">
      {#each $store.exits as exit}
        {#if !exit.hidden}
          <li style="margin-right: 5px;">
            <button
              class="btn waves-effect waves-light btncolor darken-1 orange"
              on:click="{() => takeExit(exit.name)}"
            >
              {exit.name}
            </button>
          </li>
        {/if}
      {/each}
    </ul>
  {/if}

  <button
    class="btn waves-effect waves-light btncolor darken-1 left {$showActions === true ? "blue" : ""}"
    on:click="{() => toggleActions()}"
  >
    <i class="material-icons">flare</i>
  </button>

  {#if $showActions === true}
    <ul class="ul2">
      {#each $store.actions as action}
        <li style="margin-right: 5px;">
          <button
            class="btn waves-effect waves-light btncolor darken-1 blue"
            on:click="{() => takeExit(action.name)}"
          >
            {action.name}
          </button>
        </li>
      {/each}

    </ul>
  {/if}
 <button
    class="btn waves-effect waves-light btncolor darken-1 left {$showSkills === true ? "green" : ""}"
    on:click="{() => toggleSkills()}"
  >
    <i class="material-icons">brightness_7</i>
  </button>
  {#if $showSkills}
    <ul class="ul2">
      <li style="margin-right: 5px;">
        <button
          class="btn waves-effect waves-light btncolor darken-1 green"
          on:click="{() => takeExit('look')}"
        >
          look
        </button>
      </li>

    </ul>
  {/if}

  <div style="clear:both;"></div>
  {#if $showInventory === true}
    <div class="inventory right-align">
      <ul class="collection">
        <li class="collection-item item left-align">
          <img src="img/sword.png" alt="" />

          <span class="item-header">
            Sword of Long John Silver that never was
          </span>
          <br />
          <span style="font-size:10px; color:orange;">legendary sword</span>
          <span style="font-size:10px; color:grey; float:right;">
            main hand
          </span>

          <p style="font-size:10px; margin: 0;" class="center-align">
            +10 Dex / +5 Str / +7 Stam
          </p>
          <p
            style="margin: 0; font-size:10px; color:grey; font-style:italic;"
            class="center-align"
          >
            He never needed it anyway so i took it
          </p>
        </li>

        {#each Array(10) as _, i}
          <li class="collection-item item left-align">
            <img src="img/sword.png" alt="" />

            <span class="item-header">Sturdy Iron Helmet</span>
            <br />
            <span style="font-size:10px; color:purple;">epic helmet</span>
            <span style="font-size:10px; color:grey; float:right;">head</span>

            <p style="font-size:10px; margin: 0;" class="center-align">
              +5 Dex / +20 Str / +3 Stam
            </p>
            <p
              style="margin: 0; font-size:10px; color:grey; font-style:italic;
              max-width: 250px; line-height:95%;"
              class="center-align"
            >
              The person this belonged to had probably a very tiny head, not
              sure how he ever fitted into it
            </p>
          </li>
        {/each}
        }
      </ul>
    </div>
  {/if}

</div>
