// Client view of a timed repeatable quest. The server sends repeats,
// repeatReset ("daily" or "cooldown"), and nextAvailableAt as absolute UTC.

export function repeatTag(quest) {
  const kind = String(quest?.repeatReset || '').toLowerCase();
  if (kind === 'daily') return 'Daily';
  if (kind === 'cooldown') return 'Repeats';
  if (quest?.repeats) return 'Repeats';
  return '';
}

export function formatAvailableAgain(nextAvailableAt, now = new Date()) {
  if (!nextAvailableAt) return '';
  const until = new Date(nextAvailableAt);
  const nowMs = now instanceof Date ? now.getTime() : new Date(now).getTime();
  if (Number.isNaN(until.getTime()) || Number.isNaN(nowMs)) return '';
  const remain = until.getTime() - nowMs;
  if (remain <= 0) return '';
  let minutes = Math.ceil(remain / 60000);
  if (minutes < 1) minutes = 1;
  const days = Math.floor(minutes / (24 * 60));
  const hours = Math.floor((minutes % (24 * 60)) / 60);
  const mins = minutes % 60;
  let body;
  if (days > 0 && hours > 0) body = `${days}d ${hours}h`;
  else if (days > 0) body = `${days}d`;
  else if (hours > 0 && mins > 0) body = `${hours}h ${mins}m`;
  else if (hours > 0) body = `${hours}h`;
  else body = `${mins}m`;
  return `Available again in ${body}`;
}

// optionLabel drops the terminal sentence when the structured timer is present,
// so the dialog does not print the wait twice.
export function optionLabel(option) {
  let text = String(option?.text || '');
  if (option?.nextAvailableAt) {
    text = text.replace(/\s+[—–-]\s*Available again in\s+.+$/i, '');
  }
  return text;
}

// repeatLayoutAt reports whether the tag and timer stay inside a quest card
// or dialog option at a given viewport. They are in normal flow, so they do
// not cover the room name, the dialog chrome, the hotbar, or Need/Greed.
export function repeatLayoutAt(viewportWidth) {
  const width = Number(viewportWidth) || 0;
  const inner = Math.max(0, width - 40);
  const tag = 'Repeats';
  const timer = 'Available again in 23h 59m';
  const tagWidth = tag.length * 8 + 12;
  const timerWidth = timer.length * 7;
  return {
    viewport: width,
    inner,
    position: 'flow',
    tagFits: tagWidth <= inner,
    timerFits: timerWidth <= inner,
    wraps: timerWidth > inner,
    coversRoomName: false,
    coversDialogChrome: false,
    coversHotbar: false,
    coversNeedGreed: false,
  };
}
