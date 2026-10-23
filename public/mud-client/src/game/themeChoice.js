/** Named themes a deployment or a player can select. `auto` follows the server. */
export const THEME_IDS = ["dark-fantasy", "clean-hud", "gilded", "quiet-dark"];
export const THEME_AUTO = "auto";
export const ENGINE_DEFAULT_THEME = "dark-fantasy";

export function knownTheme(value) {
  return THEME_IDS.includes(value) ? value : "";
}

export function serverThemeOrFallback(serverDefault) {
  return knownTheme(serverDefault) || ENGINE_DEFAULT_THEME;
}

/**
 * Stored theme plus the explicit flag, after one migration.
 * A saved `dark-fantasy` with no explicit flag was the old engine default,
 * so it becomes `auto`. `clean-hud` and `gilded` were picker choices.
 */
export function migrateThemeSettings(theme, themeExplicit) {
  const explicit = themeExplicit === true;
  if (theme === THEME_AUTO || theme == null || theme === "") {
    return { theme: THEME_AUTO, themeExplicit: false };
  }
  if (theme === ENGINE_DEFAULT_THEME && !explicit) {
    return { theme: THEME_AUTO, themeExplicit: false };
  }
  if (!knownTheme(theme)) {
    return { theme: THEME_AUTO, themeExplicit: false };
  }
  if (explicit) return { theme, themeExplicit: true };
  return { theme, themeExplicit: true };
}

/** Theme id written onto `document.body.dataset.theme`. */
export function resolveAppliedTheme(theme, serverDefault) {
  if (theme && theme !== THEME_AUTO && knownTheme(theme)) return theme;
  return serverThemeOrFallback(serverDefault);
}
