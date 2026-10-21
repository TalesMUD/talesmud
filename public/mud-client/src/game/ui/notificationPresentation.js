export function notificationsBlocked(state) {
  return !!(state && (state.inCombat || state.combatPhase === 'active'
    || state.combatPhase === 'ending' || state.combatOutcome || state.needGreed?.pending));
}

export function createVisibleLifetime(expire, clock = {
  now: () => performance.now(),
  setTimeout: (callback, duration) => setTimeout(callback, duration),
  clearTimeout: (timer) => clearTimeout(timer),
}) {
  const remaining = new WeakMap();
  return (node, { notification, duration }) => {
    const balance = remaining.get(notification) ?? duration;
    const started = clock.now();
    let finished = false;
    const timer = clock.setTimeout(() => {
      finished = true;
      remaining.delete(notification);
      expire(notification);
    }, balance);
    return {
      destroy() {
        clock.clearTimeout(timer);
        if (!finished) remaining.set(notification, Math.max(0, balance - (clock.now() - started)));
      },
    };
  };
}
