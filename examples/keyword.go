package main

import (
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"regexp"
	"slackbot"
	"syscall"
)

// Boot a bot that listens for the keyword "fire"
func main() {
	bot := slackbot.NewBot(os.Getenv("SLACK_TOKEN"), os.Getenv("SLACK_SIGNING_SECRET"))

	// register keywords
	keyword, _ := regexp.Compile("(?i)fire") // case insensitive "fire"
	bot.RegisterKeyword(keyword, testKeywordHandler)

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

func testKeywordHandler(bot *slackbot.Bot, container slackbot.MessageEventContainer) {
	logrus.Infoln(container.Event.Text)
}
