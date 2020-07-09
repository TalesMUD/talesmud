<style>
  .modal {
    background-color: #000000aa;
    color: #eee;
       border-width: 1px;
    border-style: solid;
    border-color: #ffffff33;
    border-radius: 0.5em;
    padding: 1em;
    width: 50%;
  }
  .modal-content {
    padding: 0;
  }

  .modal-footer {
        background-color: #000000ee;

  }

  .inventory {
padding: 0;

  }

  .inventory ul {
    border: none;
    margin: 1em;
  }

  .inventory ul li:last-child {
    border: 1px #ffffff55 solid;
  }

  .item {
    background: #000000cc;
    color: #fff;
    border: 1px #ffffff55 solid;
    margin: 0.5em;
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
</style>

<script>
  import { onMount } from "svelte";
  import { writable } from "svelte/store";

  onMount(async () => {
    document.addEventListener("DOMContentLoaded", function () {
      var elems = document.querySelectorAll(".modal");
      var instances = M.Modal.init(elems, {});
    });
  });

  const templates = writable([]);
</script>

<!-- Modal Structure -->
<div id="inventoryModal" class="modal">
  <div class="modal-content">
    <h4>Inventory</h4>

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

      </ul>
    </div>

  </div>
  <div class="modal-footer">
    <button href="#!" class="modal-close waves-effect waves-red btn">
      Close
    </button>
    <button href="#!" class="modal-close waves-effect waves-green btn">
      Drop
    </button>
  </div>
</div>
