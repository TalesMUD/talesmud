package messages

import "sync"

// A command reply is stamped with the player's notice generation so a text
// client can drop a line that arrives after the next key. Combat messages
// leave the generation at zero and are not command output.
type noticeBook struct {
	mu   sync.Mutex
	seq  uint64
	busy map[string]uint64
	last map[string]uint64
}

var notices noticeBook

// BeginNotice opens one command for this player and returns its generation.
func BeginNotice(userID string) uint64 {
	if userID == "" {
		return 0
	}
	notices.mu.Lock()
	defer notices.mu.Unlock()
	notices.seq++
	if notices.busy == nil {
		notices.busy = map[string]uint64{}
		notices.last = map[string]uint64{}
	}
	notices.busy[userID] = notices.seq
	notices.last[userID] = notices.seq
	return notices.seq
}

// EndNotice closes the command. Lines already stamped keep their generation.
func EndNotice(userID string) {
	if userID == "" {
		return
	}
	notices.mu.Lock()
	defer notices.mu.Unlock()
	delete(notices.busy, userID)
}

// LastNoticeGen is the generation of this player's most recent command.
func LastNoticeGen(userID string) uint64 {
	notices.mu.Lock()
	defer notices.mu.Unlock()
	if notices.last == nil {
		return 0
	}
	return notices.last[userID]
}

// StampNotice copies the open command generation for this player.
// Zero means the send is not part of that command.
func StampNotice(userID string) uint64 {
	if userID == "" {
		return 0
	}
	notices.mu.Lock()
	defer notices.mu.Unlock()
	if notices.busy == nil {
		return 0
	}
	return notices.busy[userID]
}
