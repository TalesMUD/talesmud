<style>

</style>

<script>
	import Credits from './Credits.svelte';
  //import { Router, Link, Route, navigate } from "svelte-routing";
  import { Router, Route, Link } from 'yrv';

  import Game from "./game/Game.svelte";
  import Welcome from "./Welcome.svelte";
  import Creator from "./creator/Creator.svelte";
  import Characters from "./characters/Characters.svelte";
  import UserForm from "./UserForm.svelte";
  
</script>


<Route exact path="/account" component="{UserForm}" />
<Route path="/creator" component="{Creator}" />
<Route exact path="/list" component="{Characters}" />
<Route exact path="/credits" component="{Credits}" />
<Route exact path="/play" component="{Game}"/>
  
<Route exact path="/">
  <Welcome />
</Route>
