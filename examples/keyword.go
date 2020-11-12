package main

import (
	"github.com/bushelpowered/slackbot"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"
)

// Boot a bot that listens for the keyword "fire"
func main() {
	bot := slackbot.NewBot(os.Getenv("SLACK_TOKEN"), os.Getenv("SLACK_SIGNING_SECRET"))

	// register keywords
	keyword, _ := regexp.Compile("(?i)fire") // case insensitive "fire"
	bot.RegisterKeyword(keyword, exampleKeywordCallback)

	// boot the bot
	err := bot.Boot(":8000")
	if err != nil {
		logrus.WithError(err).Fatalln("Failed to start bot")
		return
	}
	defer bot.Shutdown(time.Second * 10)

	// wait for exit
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Infoln("Shutting down...")
}

func exampleKeywordCallback(bot *slackbot.Bot, container slackbot.MessageEventContainer) {
	logrus.Infoln(container)
}
