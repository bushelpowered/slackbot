package main

import (
	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"os"
	"os/signal"
	"slackbot"
	"syscall"
)

// Boot a bot that responds with select menu options
func main() {
	bot := slackbot.NewBot(os.Getenv("SLACK_TOKEN"), os.Getenv("SLACK_SIGNING_SECRET"))

	// register select options
	bot.RegisterSelectOptions("callback1", exampleSelectOptionsCallback)
	bot.RegisterSelectOptionGroups("callback2", exampleSelectOptionGroupsCallback)

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

func exampleSelectOptionsCallback(bot *slackbot.Bot, interaction slack.InteractionCallback) slack.OptionsResponse {
	logrus.Infoln(interaction)
	return slack.OptionsResponse{}
}

func exampleSelectOptionGroupsCallback(bot *slackbot.Bot, interaction slack.InteractionCallback) slack.OptionGroupsResponse {
	logrus.Infoln(interaction)
	return slack.OptionGroupsResponse{}
}
