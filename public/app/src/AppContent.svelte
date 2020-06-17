<style>

</style>

<script>
  import { Router, Link, Route, navigate } from "svelte-routing";

  import Game from "./game/Game.svelte";
  import Welcome from "./Welcome.svelte";
  import Creator from "./creator/Creator.svelte";
  import Characters from "./characters/Characters.svelte";
  import UserForm from "./UserForm.svelte";
</script>

<Route path="account" component="{UserForm}" />
<Route path="creator/*" component="{Creator}" />
<Route path="list" component="{Characters}" />
<Route path="play" component="{Game}">
  <Game />
</Route>
<Route path="/">
  <Welcome />
</Route>
