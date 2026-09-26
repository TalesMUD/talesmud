// src/auth.js

import { onMount, setContext, getContext } from "svelte";
import { writable } from "svelte/store";
import createAuth0Client from "@auth0/auth0-spa-js";
import { clearGuestToken, clearLocalSession, readGuestToken, restoredSession } from "./authSession.js";

const isLoading = writable(true);
const isAuthenticated = writable(false);
const authToken = writable("");
const userInfo = writable({});
const authError = writable(null);
const AUTH_KEY = {};

// Refresh token 30 minutes before typical expiration
// Auth0 access tokens typically expire in 24 hours, but we refresh more frequently
const refreshRate = 30 * 60 * 1000; // 30 minutes

function createAuth(config) {
  let auth0 = null;
  let intervalId = undefined;

  onMount(async () => {
    try {
      auth0 = await createAuth0Client({
        domain: config.domain,
        client_id: config.client_id,
        audience: config.audience,
        cacheLocation: "localstorage",  // Persist tokens across page reloads
        useRefreshTokens: true,          // Enable refresh token rotation for long sessions
      });

      const params = new URLSearchParams(window.location.search);

      // Check if something went wrong during login redirect
      if (params.has("error")) {
        authError.set(new Error(params.get("error_description")));
      }

      // Handle redirect callback after login
      if (params.has("code")) {
        try {
          await auth0.handleRedirectCallback();
          // Clear URL parameters after successful callback
          window.history.replaceState({}, document.title, window.location.pathname);
          authError.set(null);
        } catch (callbackError) {
          console.error("Error handling redirect callback:", callbackError);
          authError.set(callbackError);
        }
      }

      const _isAuthenticated = await auth0.isAuthenticated();
      const guestToken = readGuestToken();
      // Auth0 wins. A guest token is restored only when this tab is not signed in,
      // so a reload keeps a guest in the game. Logout clears that token first.
      const sessionKind = restoredSession({
        auth0Authenticated: _isAuthenticated,
        guestToken,
      });

      if (sessionKind === "auth0") {
        // Drop a guest token left over from this tab so logout cannot fall
        // back into that guest. Leave the character-picker flag alone so a
        // reload does not open the picker again.
        clearGuestToken();
        isAuthenticated.set(true);
      } else if (sessionKind === "guest") {
        authToken.set(guestToken);
        isAuthenticated.set(true);
      } else {
        isAuthenticated.set(false);
      }

      if (_isAuthenticated) {
        // Fetch user profile from Auth0
        userInfo.set(await auth0.getUser());

        try {
          // Get the access token with proper parameter format
          const token = await auth0.getTokenSilently({ audience: config.audience });
          authToken.set(token);

          // Set up automatic token refresh
          intervalId = setInterval(async () => {
            try {
              const newToken = await auth0.getTokenSilently({ audience: config.audience });
              authToken.set(newToken);
              console.log("Token refreshed successfully");
            } catch (refreshError) {
              console.error("Token refresh failed:", refreshError);

              // Check if it's a login_required error (session truly expired)
              if (refreshError.error === "login_required" ||
                  refreshError.error === "consent_required") {
                isAuthenticated.set(false);
                authToken.set("");
                userInfo.set({});
                authError.set(refreshError);
                clearInterval(intervalId);
              }
              // For other errors, we might just have a temporary network issue
              // Don't immediately log out, let the next refresh attempt try again
            }
          }, refreshRate);
        } catch (tokenError) {
          console.error("Failed to get initial token:", tokenError);

          // If we can't get a token but were authenticated, session may be stale
          if (tokenError.error === "login_required") {
            isAuthenticated.set(false);
            authError.set(tokenError);
          }
        }
      }
    } catch (initError) {
      console.error("Failed to initialize Auth0:", initError);
      authError.set(initError);
      const guestToken = readGuestToken();
      if (guestToken && restoredSession({ auth0Authenticated: false, guestToken }) === "guest") {
        authToken.set(guestToken);
        isAuthenticated.set(true);
      }
    }

    isLoading.set(false);

    // Cleanup on component unmount
    return () => {
      if (intervalId) {
        clearInterval(intervalId);
      }
    };
  });

  const login = async (redirectPage, options = {}) => {
    if (!auth0) {
      console.error("Auth0 client not initialized");
      return;
    }

    await auth0.loginWithRedirect({
      redirect_uri: redirectPage || window.location.origin + "/play",
      ...options,
    });
  };

  const logout = async () => {
    // Clear interval before logout
    if (intervalId) {
      clearInterval(intervalId);
    }

    // Guest token and Auth0 cache both have to go. Leaving the guest token
    // made the next load skip the welcome choice and re-enter as a guest.
    clearLocalSession();
    authToken.set("");
    isAuthenticated.set(false);
    userInfo.set({});
    authError.set(null);

    if (!auth0) {
      window.location.assign(window.location.origin + "/play/");
      return;
    }

    auth0.logout({
      returnTo: window.location.origin + "/play",
    });
  };

  // Check if session is still valid (useful for components to call)
  const checkSession = async () => {
    if (!auth0) return false;

    try {
      const token = await auth0.getTokenSilently({ audience: config.audience });
      authToken.set(token);
      isAuthenticated.set(true);
      return true;
    } catch (error) {
      console.error("Session check failed:", error);
      isAuthenticated.set(false);
      authToken.set("");
      return false;
    }
  };

  const auth = {
    isLoading,
    isAuthenticated,
    authToken,
    authError,
    login,
    logout,
    userInfo,
    checkSession,
  };

  // Put everything in context so that child
  // components can access the state
  setContext(AUTH_KEY, auth);

  return auth;
}

// Helper function for child components to access the auth context
function getAuth() {
  return getContext(AUTH_KEY);
}

export { createAuth, getAuth };
