<style>
  :global(body) {
    background-color: #263238;
    transition: background-color 0.3s;
    margin: 0 auto;
    padding: 0px;
    color: #d8dee9;

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
  .iconspacing {
    margin-right: 0.5em;
  }

  :global(a:visited) {
    text-decoration: none;
    text-decoration-line: none;
    color: #aaa;
  }

  .back-link {
    position: absolute;
    left: 15px;
    top: 15px;
    z-index: 1000;
    padding: 1em;
  }
</style>

<script>
  import { BookOpenIcon } from "svelte-feather-icons";
  import Game from "./game/Game.svelte";
  import { onMount } from "svelte";
  import UserMenu from "./UserMenu.svelte";
  import SettingsModal from "./game/ui/SettingsModal.svelte";
  import { createAuth } from "./auth.js";
  import { getServerInfo } from "./api/server-info.js";

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

  let serverName = "Tales";

  String.prototype.capitalize = function () {
    return this.charAt(0).toUpperCase() + this.slice(1);
  };

  onMount(async () => {
    getServerInfo(
      (data) => {
        if (data.serverName) {
          serverName = data.serverName;
        }
      },
      (err) => console.warn("Could not load server info:", err)
    );
  });
</script>

<svelte:head>
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
</svelte:head>

<div class="root default">
  <!-- Back to main site link -->
  <a href="/" class="back-link">
    <span class="valign-wrapper italic">
      <span class="iconspacing">
        <BookOpenIcon size="24" />
      </span>
      {serverName}
    </span>
  </a>

  <!-- User menu in top right -->
  <div style="position: absolute; right: 15px; top: 15px; z-index: 1000;">
    <ul style="display: flex; list-style: none; margin: 0; padding: 0;">
      <UserMenu />
    </ul>
  </div>

  <!-- Game component -->
  <Game />

  <!-- Settings Modal (global) -->
  <SettingsModal />
</div>
