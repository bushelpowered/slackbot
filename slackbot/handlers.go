package slackbot

import (
	"github.com/gin-gonic/gin"
	"github.com/slack-go/slack"
	"net/http"
)

func (b *Bot) newCommandHandler(callback CommandCallback) func(ctx *gin.Context) {
	return func(c *gin.Context) {
		command, err := slack.SlashCommandParse(c.Request)
		if err != nil {
			_ = c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		msg := callback(b, command)

		if msg != nil {
			c.JSON(200, msg)
		} else {
			c.Status(200)
		}
	}
}