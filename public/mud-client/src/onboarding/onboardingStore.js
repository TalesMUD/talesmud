import { writable } from "svelte/store";

// Onboarding phases:
// "loading"    - Auth0 still initializing
// "welcome"    - Not authenticated, show welcome/login
// "nickname"   - Authenticated but needs nickname
// "character"  - Authenticated, has nickname, no characters
// "ready"      - Authenticated, has nickname, has characters -> show Game
const onboardingPhase = writable("loading");
const userData = writable(null);
const userCharacters = writable([]);

export { onboardingPhase, userData, userCharacters };
