package main

import (
	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack/slackevents"
	"os"
	"os/signal"
	"slackbot/slackbot"
	"syscall"
)

// Boot a bot with a slash command that echos Hello World!
func main() {
	bot := slackbot.NewBot("bot token", "signing secret")

	// register events
	bot.RegisterEvent(slackevents.AppMention, testAppMentionHandler)

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

func testAppMentionHandler(bot *slackbot.Bot, event slackevents.EventsAPIEvent) {
	appMentionEvent := event.InnerEvent.Data.(*slackevents.AppMentionEvent)
	logrus.Infoln(appMentionEvent)
}
