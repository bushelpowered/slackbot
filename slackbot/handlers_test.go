package slackbot

import (
	"github.com/gin-gonic/gin"
	"github.com/slack-go/slack"
	"net/http"
	"testing"
)

func TestCommandHandlerWithNoContent(t *testing.T) {
	engine := gin.New()

	bot := newBot()
	bot.RegisterCommand("test", func(bot *Bot, command slack.SlashCommand) *slack.Msg {
		return nil
	})
	bot.prepareEngine(engine, false)

	e := getHttpExpect(t, engine)
	e.POST("/slack/commands/test").
		WithFormField("command", "/test").
		Expect().
		Status(http.StatusOK).NoContent()
}

func TestCommandHandlerWithMessage(t *testing.T) {
	var textContent = "hello world!"

	bot := newBot()
	bot.RegisterCommand("test", func(bot *Bot, command slack.SlashCommand) *slack.Msg {
		return &slack.Msg{Text: textContent}
	})

	engine := gin.New()
	bot.prepareEngine(engine, false)

	e := getHttpExpect(t, engine)
	e.POST("/slack/commands/test").
		WithFormField("command", "/test").
		Expect().
		Status(http.StatusOK).JSON().Object().ValueEqual("text", textContent)
}

func TestCommandHandlerWithInvalidRequest(t *testing.T) {
	bot := newBot()
	bot.RegisterCommand("test", func(bot *Bot, command slack.SlashCommand) *slack.Msg {
		return nil
	})

	engine := gin.New()
	bot.prepareEngine(engine, false)

	e := getHttpExpect(t, engine)
	e.POST("/slack/commands/test").
		Expect().
		Status(http.StatusBadRequest)
}
