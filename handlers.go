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
	return func(ctx *gin.Context) {
		command, err := slack.SlashCommandParse(ctx.Request)
		if err != nil {
			_ = ctx.AbortWithError(http.StatusBadRequest, err)
			return
		}
		if command.Command == "" {
			_ = ctx.AbortWithError(http.StatusBadRequest, errors.New("invalid command"))
			return
		}

		msg := callback(b, command)

		if msg != nil {
			ctx.JSON(http.StatusOK, msg)
		} else {
			ctx.Status(http.StatusOK)
		}
	}
}

func (b *Bot) newEventHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		body, err := ioutil.ReadAll(ctx.Request.Body)
		if err != nil {
			_ = ctx.AbortWithError(http.StatusBadRequest, err)
			return
		}

		event, err := slackevents.ParseEvent(body, slackevents.OptionNoVerifyToken()) // verification handled by middleware
		if err != nil {
			_ = ctx.AbortWithError(http.StatusBadRequest, err)
			return
		}

		if event.Type == slackevents.URLVerification {
			var r *slackevents.ChallengeResponse
			err := json.Unmarshal(body, &r)
			if err != nil {
				_ = ctx.AbortWithError(http.StatusBadRequest, err)
				return
			}
			ctx.Data(http.StatusOK, "text/plain", []byte(r.Challenge))
			return
		}

		if event.Type == slackevents.CallbackEvent {
			innerEvent := event.InnerEvent
			b.RLock()
			defer b.RUnlock()
			for eventType, eventCallbacks := range b.events {
				if innerEvent.Type == eventType {
					for _, callback := range eventCallbacks {
						callback(b, event)
					}
				}
			}
			ctx.Status(http.StatusOK)
			return
		}

		if event.Type == slackevents.AppRateLimited {
			_ = ctx.Error(errors.New("app rate limited"))
			ctx.Status(http.StatusOK)
			return
		}
	}
}

func (b *Bot) newInteractiveHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		payload := ctx.PostForm("payload")
		if payload == "" {
			_ = ctx.AbortWithError(http.StatusBadRequest, ErrEmptyPayload)
			return
		}

		var interactionCallback slack.InteractionCallback
		err := json.Unmarshal([]byte(payload), &interactionCallback)
		if err != nil {
			_ = ctx.AbortWithError(http.StatusBadRequest, ErrBadPayload)
			return
		}

		b.RLock()
		defer b.RUnlock()
		callbacks, typeExists := b.interactives[interactionCallback.Type]
		if typeExists {
			for _, callback := range callbacks {
				response := callback(b, interactionCallback)
				if response != nil {
					ctx.JSON(http.StatusOK, response)
					return
				}
			}
		}

		ctx.Status(http.StatusOK)
	}
}
