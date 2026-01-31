<style>
  :global(body) {
    background-color: #06080c;
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

  .user-menu-wrapper {
    position: absolute;
    right: 15px;
    top: 15px;
    z-index: 1000;
  }

  .user-menu-wrapper ul {
    display: flex;
    list-style: none;
    margin: 0;
    padding: 0;
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
  import { getUser } from "./api/user.js";
  import { getMyCharacters } from "./api/characters.js";

  // Onboarding components
  import LoadingScreen from "./onboarding/LoadingScreen.svelte";
  import WelcomeScreen from "./onboarding/WelcomeScreen.svelte";
  import NicknameSetup from "./onboarding/NicknameSetup.svelte";
  import CharacterCreationWizard from "./onboarding/CharacterCreationWizard.svelte";

  // Auth0 config
  const config = {
    domain: "owndnd.eu.auth0.com",
    client_id: "mxcEqTuAUOzrL798mbVTpqFxpGGVp3gI",
    audience: "http://talesofapirate.com/dnd/api",
  };

  const { isLoading, isAuthenticated, authToken, authError, login, logout, userInfo } = createAuth(config);

  // Onboarding phase: loading | welcome | nickname | character | ready
  let phase = "loading";
  let serverName = "Tales";
  let currentUser = null;
  let loadingUser = false;

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

  // Phase detection: single reactive block to avoid race conditions
  // between auth state changes and onboarding data loading
  $: if (!$isLoading) {
    if (!$isAuthenticated) {
      phase = "welcome";
    } else if ($authToken && !loadingUser && phase === "loading") {
      // Only trigger once (when phase is still "loading")
      loadOnboardingData();
    }
  }

  function loadOnboardingData() {
    loadingUser = true;

    getUser(
      $authToken,
      (user) => {
        currentUser = user;
        if (user.isNewUser || !user.nickname) {
          phase = "nickname";
          loadingUser = false;
        } else {
          checkCharacters();
        }
      },
      (err) => {
        console.error("Failed to load user:", err);
        // If user fetch fails, show nickname setup as fallback
        phase = "nickname";
        loadingUser = false;
      }
    );
  }

  function checkCharacters() {
    getMyCharacters(
      $authToken,
      (chars) => {
        const charList = chars || [];
        phase = charList.length > 0 ? "ready" : "character";
        loadingUser = false;
      },
      (err) => {
        console.error("Failed to load characters:", err);
        phase = "character";
        loadingUser = false;
      }
    );
  }

  function onNicknameComplete(updatedUser) {
    currentUser = updatedUser;
    checkCharacters();
  }

  function onCharacterCreated() {
    phase = "ready";
  }
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
  <link rel="preconnect" href="https://fonts.googleapis.com" />
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin />
  <link
    href="https://fonts.googleapis.com/css2?family=Cinzel+Decorative:wght@400;700;900&family=Cinzel:wght@400;500;600;700&family=Fira+Code:wght@300;400;500&family=Cormorant+Garamond:ital,wght@0,300;0,400;0,600;1,300;1,400&display=swap"
    rel="stylesheet"
  />
</svelte:head>

{#if phase === "loading"}
  <LoadingScreen />

{:else if phase === "welcome"}
  <WelcomeScreen {login} {serverName} authError={$authError} />

{:else if phase === "nickname"}
  <div class="user-menu-wrapper">
    <ul>
      <UserMenu />
    </ul>
  </div>
  <NicknameSetup
    authToken={$authToken}
    userInfo={$userInfo}
    {currentUser}
    onComplete={onNicknameComplete}
  />

{:else if phase === "character"}
  <div class="user-menu-wrapper">
    <ul>
      <UserMenu />
    </ul>
  </div>
  <CharacterCreationWizard
    authToken={$authToken}
    onComplete={onCharacterCreated}
  />

{:else}
  <!-- phase === "ready" — normal game view -->
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
    <div class="user-menu-wrapper">
      <ul>
        <UserMenu />
      </ul>
    </div>

    <!-- Game component -->
    <Game />

    <!-- Settings Modal (global) -->
    <SettingsModal />
  </div>
{/if}
