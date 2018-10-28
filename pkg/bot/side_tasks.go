package bot

import (
	"eonbot/pkg/control"
)

// execSideTasks executes specified bot/exchange
// tasks.
func (b *botProcess) execSideTasks(c control.Commons) {
	// if cancel all is allowed, cancel all open orders
	// on all streams.
	if c.CancelAll {
		// if execution is unsuccessful, repeat it
		// as many times as user specified.
		for i := 0; i < b.Conf.MainConfig().Get().BotConfig.SideTaskRestarts; i++ {
			if !b.execCancelAll() {
				continue
			}
			break
		}
	}

	// if sell all is allowed, sell all base
	// assets.
	if c.SellAll {
		// if execution is unsuccessful, repeat it
		// as many times as user specified.
		for i := 0; i < b.Conf.MainConfig().Get().BotConfig.SideTaskRestarts; i++ {
			if !b.execSellAll() {
				continue
			}
			break
		}
	}
}
