package telegram

import (
	"eonbot/pkg/control"
	"fmt"
	"strings"

	"eonbot/pkg"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	cmdHelp    = "help"
	cmdVersion = "version"
	cmdStart   = "start"
	cmdStop    = "stop"
	cmdRestart = "restart"
	cmdStatus  = "status"
	cmdNotifs  = "notifs"
)

var helpList = map[string]string{
	cmdHelp:    fmt.Sprintf("/%s - Shows commands list.", cmdHelp),
	cmdVersion: fmt.Sprintf("/%s - Shows bot's version.", cmdVersion),
	cmdStart:   fmt.Sprintf("/%s - Starts the bot.", cmdStart),
	cmdStop:    fmt.Sprintf("/%s - Stops the bot.", cmdStop),
	cmdRestart: fmt.Sprintf("/%s - Restarts the bot.", cmdRestart),
	cmdStatus:  fmt.Sprintf("/%s - Shows current state of the bot.", cmdStatus),
	cmdNotifs:  fmt.Sprintf("/%s - Toggles (enables/disables) all bot notifications in this chat.", cmdNotifs),
}

func (t *Telegram) parseCMD(cmd *tgbotapi.Message) {
	if !cmd.IsCommand() {
		t.sendAndAbsorb("Not a command.\nUse /help command to see possible commands list.", cmd.Chat.ID)
	}

	switch cmd.Command() {
	case cmdHelp:
		t.helpCMD(cmd)
	case cmdVersion:
		t.versionCMD(cmd)
	case cmdStatus:
		t.statusCMD(cmd)
	case cmdStart:
		t.startCMD(cmd)
	case cmdStop:
		t.stopCMD(cmd)
	case cmdRestart:
		t.restartCMD(cmd)
	case cmdNotifs:
		t.notifsCMD(cmd)
	default:
		t.sendAndAbsorb("Command not recognized.\nUse /help command to see possible commands list.", cmd.Chat.ID)
	}
}

func (t *Telegram) helpCMD(cmd *tgbotapi.Message) {
	var str strings.Builder
	str.WriteString("Commands list:\n")
	for _, s := range helpList {
		str.WriteString(s + "\n")
	}

	t.sendAndAbsorb(str.String(), cmd.Chat.ID)
}

func (t *Telegram) versionCMD(cmd *tgbotapi.Message) {
	t.sendAndAbsorb(pkg.FullVersion(), cmd.Chat.ID)
}

func (t *Telegram) statusCMD(cmd *tgbotapi.Message) {
	t.sendAndAbsorb(t.bot.control.State().String(), cmd.Chat.ID)
}

func (t *Telegram) startCMD(cmd *tgbotapi.Message) {
	res := t.bot.control.Start(control.StartInfo{}, control.CauseRC)
	t.sendAndAbsorb(res.String(), cmd.Chat.ID)
}

func (t *Telegram) stopCMD(cmd *tgbotapi.Message) {
	res := t.bot.control.Stop(control.StopInfo{}, control.CauseRC)
	t.sendAndAbsorb(res.String(), cmd.Chat.ID)
}

func (t *Telegram) restartCMD(cmd *tgbotapi.Message) {
	t.bot.control.Restart(control.StartInfo{}, control.StopInfo{}, control.CauseRC)
	t.sendAndAbsorb("Bot is restarting...", cmd.Chat.ID)
}

func (t *Telegram) notifsCMD(cmd *tgbotapi.Message) {
	var msg string
	if t.subscriberExists(cmd.Chat.ID) {
		t.removeSubscriber(cmd.Chat.ID)
		msg = "Notifications for this chat room have been disabled."
	} else {
		t.setSubscriber(cmd.Chat.ID, true)
		msg = "Notifications for this chat room have been enabled."
	}

	t.sendAndAbsorb(msg, cmd.Chat.ID)
}
