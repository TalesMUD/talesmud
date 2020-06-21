<style>
  .sidelist {
    width: 20em;
  }
</style>

<script>
	import RoomsEditor from './RoomsEditor.svelte';
	import ItemsEditor from './ItemsEditor.svelte';
  import ItemTemplatesEditor from "./ItemTemplatesEditor.svelte";
  import WorldEditor from "./WorldEditor.svelte";


  import { Router, Route } from "svelte-routing";
  import { writable } from "svelte/store";
  import { onMount, onDestroy } from "svelte";
  import { subMenu } from "../stores.js";

  import axios from "axios";

  onMount(async () => {
    var elems = document.querySelectorAll(".tabs");
    let instance = M.Tabs.init(elems);

    document.body.style.backgroundImage = "";

    subMenu.setItems([
      {
        name: "ROOMS",
        nav: "creator/rooms",
      },
      {
        name: "ITEMS",
        nav: "creator/items",
      },
      {
        name: "ITEM TEMPLATES",
        nav: "creator/item-templates",
      },
      {
        name: "WORLD",
        nav: "creator/world",
      },
    ]);
    subMenu.show();
  });

  onDestroy(async () => {
    subMenu.hide();
  });
</script>

<Router>
  <row>
    <Route path="rooms" component="{RoomsEditor}" />
    <Route path="items" component="{ItemsEditor}" />
    <Route path="world" component="{WorldEditor}" />
    <Route path="item-templates" component="{ItemTemplatesEditor}" />
  </row>
</Router>
