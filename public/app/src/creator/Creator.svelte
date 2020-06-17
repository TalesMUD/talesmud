<style>
  .sidelist {
    width: 20em;
  }
</style>

<script>
	import SampleEditor from './SampleEditor.svelte';
  import ItemTemplatesEditor from "./ItemTemplatesEditor.svelte";
  import WorldEditor from "./WorldEditor.svelte";
  import RoomsEditor from "./RoomsEditor.svelte";

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
        name: "SAMPLE",
        nav: "creator/sample",
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
    <Route path="items" component="{ItemTemplatesEditor}" />
    <Route path="world" component="{WorldEditor}" />
    <Route path="sample" component="{SampleEditor}" />
  </row>
</Router>
