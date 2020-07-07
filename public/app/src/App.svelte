<style>
  :global(body) {
    background-color: #263238;
    transition: background-color 0.3s;
    margin: 0 auto;
    padding: 0px;
    color: #d8dee9;

    /* The image used */
    /*background-image: url("bg.jpg");*/
    /* Full height */
    height: 100%;
    /* Center and scale the image nicely */
    background-position: center;
    background-repeat: no-repeat;
    background-size: cover;
    text-decoration: none;
    text-decoration-line: none;
  }
  :global(a:href) {
    text-decoration: none;
  }
  :global(label) {
    color: #00796b;
  }
  :global(a) {
    color: #dedede;

    text-decoration: none;
  }
  :global(a:hover) {
    text-decoration: none;
  }
  .brand-logo {
    margin-left: 0em;
    margin-right: 1em;
  }
  .iconspacing {
    margin-right: 0.5em;
  }

  :global(a:visited) {
    text-decoration: none;
    text-decoration-line: none;
    color: #aaa;
  }
  nav {
    background: #00000055;
  }
  nav {
    margin-bottom: 2em;
  }
  .mobile {
    font-size: 10px;
  }
  img {
    height: 40px;
  }
</style>

<script>
  import Sprites from "./game/Sprites.svelte";
  import AppContent from "./AppContent.svelte";
  import MediaQuery from "./MediaQuery.svelte";
  import {
    HashIcon,
    ShieldIcon,
    BookOpenIcon,
    EditIcon,
    PlayIcon,
    UsersIcon,
  } from "svelte-feather-icons";
  import { Router, Link, Route, navigate } from "svelte-routing";

  import Game from "./game/Game.svelte";
  import Welcome from "./Welcome.svelte";
  import Creator from "./creator/Creator.svelte";
  import Characters from "./characters/Characters.svelte";
  import UserForm from "./UserForm.svelte";
  import { afterUpdate, onMount } from "svelte";
  import NavLink from "./components/NavLink.svelte";

  import { user, subMenu } from "./stores.js";
  import { fade } from "svelte/transition";

  import UserMenu from "./UserMenu.svelte";
  import { createAuth } from "./auth.js";

  // Auth0 config
  const config = {
    domain: "owndnd.eu.auth0.com",
    client_id: "mxcEqTuAUOzrL798mbVTpqFxpGGVp3gI",
    audience: "http://talesofapirate.com/dnd/api",
  };

  const { isLoading, isAuthenticated, authToken } = createAuth(config);
  $: state = {
    isAuthenticated: $isAuthenticated,
    authToken: $authToken.slice(0, 20),
  };

  String.prototype.capitalize = function () {
    return this.charAt(0).toUpperCase() + this.slice(1);
  };

  onMount(async () => {
    //var elems = document.querySelectorAll(".tabs");
    //let instance = M.Tabs.init(elems);
  });
  export let url = "";

  const onHLJSLoaded = () => {
    hljs.initHighlightingOnLoad();
  };
</script>

<svelte:head>
  <script src="https://cdn.auth0.com/js/auth0/9.0/auth0.min.js">

  </script>
  <link
    rel="stylesheet"
    href="https://cdnjs.cloudflare.com/ajax/libs/materialize/1.0.0/css/materialize.min.css"
  />
  <script
    src="https://cdnjs.cloudflare.com/ajax/libs/materialize/1.0.0-beta/js/materialize.min.js">

  </script>
  <link
    rel="stylesheet"
    href="https://fonts.googleapis.com/icon?family=Material+Icons"
  />
  <link
    rel="stylesheet"
    href="https://cdn.jsdelivr.net/gh/highlightjs/cdn-release@10.1.1/build/styles/atom-one-dark.min.css"
  />
  <script
    src="https://cdn.jsdelivr.net/gh/highlightjs/cdn-release@10.1.1/build/highlight.min.js"
    on:load="{onHLJSLoaded}">

  </script>
  <script
    src="https://cdn.jsdelivr.net/gh/highlightjs/cdn-release@10.1.1/build/languages/javascript.min.js">
  </script>
</svelte:head>

<div class="root default">

  <Router url="{url}">
    <nav class="nav-extended">
      <div class="nav-wrapper container">
        <a href="#" class="brand-logo">
          <span class="valign-wrapper italic">
            <span class="iconspacing">

              <BookOpenIcon size="24" />
            </span>
            <NavLink to="/">Tales</NavLink>
          </span>
        </a>

        <ul class="right hide-on-small-only">

          <li>
            <NavLink to="/play">
              <span class="valign-wrapper">
                <span class="iconspacing valign-wrapper">
                  <PlayIcon size="18" />
                </span>
                Play
              </span>
            </NavLink>

          </li>
          {#if $isAuthenticated}
            <li>
              <NavLink to="/list">
                <span class="valign-wrapper">
                  <span class="iconspacing valign-wrapper">
                    <UsersIcon size="18" />
                  </span>
                  Top Characters
                </span>
              </NavLink>
            </li>
            <li>
              <NavLink to="/creator/rooms">
                <span class="valign-wrapper">
                  <span class="iconspacing valign-wrapper">
                    <EditIcon size="18" />
                  </span>
                  Creator
                </span>
              </NavLink>
            </li>
          {/if}
          <li>
            <NavLink to="signup">News</NavLink>
          </li>
          <UserMenu />
        </ul>
      </div>

      {#if $subMenu.active}
        <MediaQuery query="(max-width: 1280px)" let:matches>
          {#if matches}
            <div class="nav-content">
              <ul class="tabs tabs-transparent">
                {#each $subMenu.entries as entry}
                  <li class="tab">
                    <NavLink to="{entry.nav}">{entry.name}</NavLink>
                  </li>
                {/each}
              </ul>
            </div>
          {:else}
            <div class="nav-content container">
              <ul class="tabs tabs-transparent">
                {#each $subMenu.entries as entry}
                  <li class="tab">
                    <NavLink to="{entry.nav}">{entry.name}</NavLink>
                  </li>
                {/each}
              </ul>
            </div>
          {/if}
        </MediaQuery>
      {/if}

    </nav>

    <MediaQuery query="(min-width: 1281px)" let:matches>
      {#if matches}
        <main class="container">
          <AppContent />

        </main>
      {:else}
        <main>
          <AppContent />
        </main>
      {/if}
    </MediaQuery>

  </Router>
</div>
