package bot

import (
	"eonbot/pkg/control"
	"eonbot/pkg/remote/inner"
)

// onStateChange is used as a callback when the state
// of the bot changes.
func (b *botProcess) onStateChange(s control.StateInfoer) {
	// omit restart notifications on config change.
	if s.GetCause() == control.CauseConfigChange {
		return
	}

	// send notifications to remote controllers.
	b.RC.TelegramSend(s.StringShort())
	b.RC.InternalSend(inner.StateChangeEvent)
}
