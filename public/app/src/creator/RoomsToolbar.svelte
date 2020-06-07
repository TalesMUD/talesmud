<style>
  ul {
    display: flex;
    justify-content: flex-end;
    flex-direction: row;
    align-items: center;
    /* width: 100%; */

    padding: 15px;
  }

  ul li {
    display: inline-block;
    list-style: none;
    color: #eee;
  }

  .btn-flat {
    color: #eee;
    margin: 5px;
  }

  .header {
    font-weight: 500;
    font-size: 200%;
  }

  ul li:first-child {
    margin-right: auto;
  }
</style>

<script>
  export let createRoom;
  export let addExit;
  export let addAction;
</script>

<ul>
  <li class="header">Manage Rooms</li>
  <li>
    <button class="btn-flat green" on:click="{createRoom()}">
      Create New Room
    </button>
  </li>
  <li>
    <button class="btn-flat blue" on:click="{addExit()}">Create Exit</button>
  </li>
  <li>
    <button class="btn-flat blue" on:click="{addAction()}">
      Create Action
    </button>
  </li>

</ul>
