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
  export let create;
  export let createFromTemplate;
</script>

<div class="row">
  <ul>
    <li class="header">Manage Item Templates</li>
    <li>
      <button class="btn-flat green" on:click="{create()}">
        Create Item Template
      </button>
    </li>
    <li>
      <button class="btn-flat blue" on:click="{createFromTemplate()}">
        Create Item From template
      </button>
    </li>

  </ul>
</div>
