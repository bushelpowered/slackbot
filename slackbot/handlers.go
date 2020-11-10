package slackbot

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"io/ioutil"
	"net/http"
)

func (b *Bot) newCommandHandler(callback CommandCallback) gin.HandlerFunc {
	return func(c *gin.Context) {
		command, err := slack.SlashCommandParse(c.Request)
		if err != nil {
			_ = c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		if command.Command == "" {
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("invalid command"))
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

func (b *Bot) newEventHandler(callbacks map[string][]EventCallback) gin.HandlerFunc {
	return func(c *gin.Context) {
		body, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		event, err := slackevents.ParseEvent(body, slackevents.OptionNoVerifyToken()) // verification handled by middleware
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		if event.Type == slackevents.URLVerification {
			var r *slackevents.ChallengeResponse
			err := json.Unmarshal(body, &r)
			if err != nil {
				c.AbortWithError(http.StatusBadRequest, err)
				return
			}
			c.Data(http.StatusOK, "text/plain", []byte(r.Challenge))
			return
		}

		if event.Type == slackevents.CallbackEvent {
			innerEvent := event.InnerEvent
			for eventType, eventCallbacks := range callbacks {
				if innerEvent.Type == eventType {
					for _, callback := range eventCallbacks {
						callback(b, event)
					}
				}
			}
			c.Status(http.StatusOK)
			return
		}

		if event.Type == slackevents.AppRateLimited {
			_ = c.Error(errors.New("app rate limited"))
			c.Status(http.StatusOK)
			return
		}
	}
}
