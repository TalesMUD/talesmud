<style>
  .mudx {
    padding: 1em;
    margin-top: 150px;
  }

  .inventory {
    padding: 1em;
    float: right;
    justify-content: flex-end;
  }

  .inventory ul {
    border: none;
    margin: 1em;
  }

  .inventory ul li:last-child {
    border: 1px #ffffff55 solid;
  }

  .item {
    background: #00000055;
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
    border-radius: 15px;
  }
</style>

<script>
  import Sprites from "./Sprites.svelte";
  import { writable } from "svelte/store";
  import { getClient } from "./Client";

  const showInventory = writable(false);

  const toggleInventory = () => {
    showInventory.set(!$showInventory);

    console.log("Showinventory is " + $showInventory);
  };

  export let store;
  export let term;
  export let sendMessage;

  const takeExit = (exit) => {
    console.log("Take exit " + exit);

    sendMessage(exit);
  };
</script>

<div style="clear:both;"></div>

<div class="mudx">

  <div class="inventory right-align">
    <button
      class="btn waves-effect waves-light btncolor darken-1"
      on:click="{() => toggleInventory()}"
    >
      inventory
    </button>

    {#if $showInventory === true}
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
            The person this belonged to had probably a very tiny head, not sure
            how he ever fitted into it
          </p>
        </li>
      </ul>
    {/if}

  </div>

  <ul class="ul2">
    {#each $store.exits as exit}
      {#if !exit.hidden}
        <li style="margin-right: 5px;">
          <button
            class="btn waves-effect waves-light btncolor darken-1"
            on:click="{() => takeExit(exit.name)}"
          >
            {exit.name}
          </button>
        </li>
      {/if}
    {/each}
  </ul>

  <ul class="ul2">
    {#each $store.actions as action}
      <li style="margin-right: 5px;">
        <button
          class="btn waves-effect waves-light btncolor darken-1"
          on:click="{() => takeExit(action.name)}"
        >
          {action.name}
        </button>
      </li>
    {/each}

  </ul>

  <ul class="ul2">
    <li style="margin-right: 5px;">
      <button
        class="btn waves-effect waves-light btncolor darken-1"
        on:click="{() => takeExit('look')}"
      >
        look
      </button>
    </li>

  </ul>
</div>
