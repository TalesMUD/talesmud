// src/auth.js

import { setContext, getContext } from "svelte";
import { writable, get } from "svelte/store";
import createAuth0Client from "@auth0/auth0-spa-js";

const isLoading = writable(true);
const isAuthenticated = writable(false);
const authToken = writable("");
const userInfo = writable({});
const authError = writable(null);
const AUTH_KEY = {};

// Refresh token 30 minutes before typical expiration
const refreshRate = 30 * 60 * 1000; // 30 minutes

// Module-level auth0 client
let auth0 = null;
let initPromise = null;

async function initAuth0(config) {
  if (auth0) return auth0;
  if (initPromise) return initPromise;

  initPromise = (async () => {
    try {
      auth0 = await createAuth0Client({
        domain: config.domain,
        client_id: config.client_id,
        audience: config.audience,
        cacheLocation: "localstorage",
        useRefreshTokens: true,
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
          window.history.replaceState({}, document.title, window.location.pathname);
          authError.set(null);
        } catch (callbackError) {
          console.error("Error handling redirect callback:", callbackError);
          authError.set(callbackError);
        }
      }

      const _isAuthenticated = await auth0.isAuthenticated();
      isAuthenticated.set(_isAuthenticated);

      if (_isAuthenticated) {
        userInfo.set(await auth0.getUser());

        try {
          const token = await auth0.getTokenSilently({ audience: config.audience });
          authToken.set(token);

          // Set up automatic token refresh
          setInterval(async () => {
            try {
              const newToken = await auth0.getTokenSilently({ audience: config.audience });
              authToken.set(newToken);
            } catch (refreshError) {
              console.error("Token refresh failed:", refreshError);
              if (refreshError.error === "login_required" || refreshError.error === "consent_required") {
                isAuthenticated.set(false);
                authToken.set("");
                userInfo.set({});
                authError.set(refreshError);
              }
            }
          }, refreshRate);
        } catch (tokenError) {
          console.error("Failed to get initial token:", tokenError);
          if (tokenError.error === "login_required") {
            isAuthenticated.set(false);
            authError.set(tokenError);
          }
        }
      }
    } catch (initError) {
      console.error("Failed to initialize Auth0:", initError);
      authError.set(initError);
    }

    isLoading.set(false);
    return auth0;
  })();

  return initPromise;
}

function createAuth(config) {
  // Start initialization immediately
  initAuth0(config);

  const login = async (redirectPage) => {
    // Wait for auth0 to be ready
    await initAuth0(config);
    if (!auth0) {
      console.error("Auth0 client not initialized");
      return;
    }

    await auth0.loginWithRedirect({
      redirect_uri: redirectPage || window.location.origin,
    });
  };

  const logout = async () => {
    await initAuth0(config);
    if (!auth0) {
      console.error("Auth0 client not initialized");
      return;
    }

    authToken.set("");
    isAuthenticated.set(false);
    userInfo.set({});
    authError.set(null);

    auth0.logout({
      returnTo: window.location.origin,
    });
  };

  const checkSession = async () => {
    await initAuth0(config);
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

  setContext(AUTH_KEY, auth);
  return auth;
}

function getAuth() {
  return getContext(AUTH_KEY);
}

export { createAuth, getAuth };
