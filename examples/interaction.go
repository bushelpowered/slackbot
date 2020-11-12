package main

import (
	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"os"
	"os/signal"
	"slackbot"
	"syscall"
)

// Boot a bot that listens for app mention events
func main() {
	bot := slackbot.NewBot(os.Getenv("SLACK_TOKEN"), os.Getenv("SLACK_SIGNING_SECRET"))

	// register interaction
	bot.RegisterShortcutInteraction("test_id", exampleShortcutInteractionCallback)

	// boot the bot
	err := bot.Boot(":8000")
	if err != nil {
		logrus.WithError(err).Fatalln("Failed to start bot")
		return
	}
	defer bot.Shutdown()

	// wait for exit
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Infoln("Shutting down...")
}

func exampleShortcutInteractionCallback(bot *slackbot.Bot, event slack.InteractionCallback) {
	logrus.Infoln(event)
}