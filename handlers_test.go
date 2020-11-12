package slackbot

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/stretchr/testify/assert"
	"net/http"
	"reflect"
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

func TestEventHandlerErrorsWithNoContent(t *testing.T) {
	engine := gin.New()

	bot := newBot()
	bot.RegisterEvent(slackevents.AppMention, func(bot *Bot, event slackevents.EventsAPIEvent) {

	})
	bot.prepareEngine(engine, false)

	e := getHttpExpect(t, engine)
	e.POST("/slack/events").
		Expect().
		Status(http.StatusBadRequest).NoContent()
}

func TestEventHandlerUrlVerification(t *testing.T) {
	var challengeString = "challenge"

	engine := gin.New()

	bot := newBot()
	bot.RegisterEvent(slackevents.AppMention, func(bot *Bot, event slackevents.EventsAPIEvent) {

	})
	bot.prepareEngine(engine, false)

	e := getHttpExpect(t, engine)
	e.POST("/slack/events").
		WithJSON(slackevents.EventsAPIURLVerificationEvent{
			Token:     "token",
			Challenge: challengeString,
			Type:      slackevents.URLVerification,
		}).
		Expect().
		Status(http.StatusOK).Text().Equal(challengeString)
}

func TestEventHandlerSingleCallback(t *testing.T) {
	hitCallbackOne := false

	engine := gin.New()

	bot := newBot()
	bot.RegisterEvent(slackevents.AppMention, func(bot *Bot, event slackevents.EventsAPIEvent) {
		hitCallbackOne = true
	})
	bot.prepareEngine(engine, false)

	e := getHttpExpect(t, engine)
	e.POST("/slack/events").
		WithJSON(newFakeEvent(slackevents.AppMention)).
		Expect().
		Status(http.StatusOK).NoContent()

	assert.True(t, hitCallbackOne)
}

func TestEventHandlerHitsAllCallbacksOfTheSameType(t *testing.T) {
	hitCallbackOne := false
	hitCallbackTwo := false

	engine := gin.New()

	bot := newBot()
	bot.RegisterEvent(slackevents.AppMention, func(bot *Bot, event slackevents.EventsAPIEvent) {
		hitCallbackOne = true
	})
	bot.RegisterEvent(slackevents.AppMention, func(bot *Bot, event slackevents.EventsAPIEvent) {
		hitCallbackTwo = true
	})
	bot.prepareEngine(engine, false)

	e := getHttpExpect(t, engine)
	e.POST("/slack/events").
		WithJSON(newFakeEvent(slackevents.AppMention)).
		Expect().
		Status(http.StatusOK).NoContent()

	assert.True(t, hitCallbackOne)
	assert.True(t, hitCallbackTwo)
}

func TestEventHandlerHitsCorrectCallback(t *testing.T) {
	hitCallbackOne := false
	hitCallbackTwo := false

	engine := gin.New()

	bot := newBot()
	bot.RegisterEvent(slackevents.AppMention, func(bot *Bot, event slackevents.EventsAPIEvent) {
		hitCallbackOne = true
	})
	bot.RegisterEvent(slackevents.Message, func(bot *Bot, event slackevents.EventsAPIEvent) {
		hitCallbackTwo = true
	})
	bot.prepareEngine(engine, false)

	e := getHttpExpect(t, engine)
	e.POST("/slack/events").
		WithJSON(newFakeEvent(slackevents.AppMention)).
		Expect().
		Status(http.StatusOK).NoContent()

	assert.True(t, hitCallbackOne)
	assert.False(t, hitCallbackTwo)
}

func TestEventHandlerRateLimited(t *testing.T) {
	engine := gin.New()

	bot := newBot()
	bot.prepareEngine(engine, false)

	e := getHttpExpect(t, engine)
	e.POST("/slack/events").
		WithJSON(slackevents.EventsAPIAppRateLimited{Type: slackevents.AppRateLimited}).
		Expect().
		Status(http.StatusOK).NoContent()
}

func newFakeEvent(eventType string) fakeEvent {
	return newFakeEventWithData(eventType, nil)
}

func newFakeEventWithData(eventType string, event interface{}) fakeEvent {
	if event == nil {
		event = slackevents.EventsAPIInnerEventMapping[eventType]
	}

	// set the type field
	v := reflect.ValueOf(&event).Elem()
	tmp := reflect.New(v.Elem().Type()).Elem()
	tmp.Set(v.Elem())
	tmp.FieldByName("Type").SetString(eventType)
	v.Set(tmp)

	return fakeEvent{
		Type:  slackevents.CallbackEvent,
		Event: event,
	}
}

type fakeEvent struct {
	Type  string      `json:"type"`
	Event interface{} `json:"event"`
}

func TestInteractiveHandlerWithNoPayload(t *testing.T) {
	engine := gin.New()

	bot := newBot()
	bot.prepareEngine(engine, false)

	e := getHttpExpect(t, engine)
	e.POST("/slack/interactives").
		Expect().
		Status(http.StatusBadRequest).NoContent()
}

func TestInteractiveHandlerWithBadPayload(t *testing.T) {
	engine := gin.New()

	bot := newBot()
	bot.prepareEngine(engine, false)

	e := getHttpExpect(t, engine)
	e.POST("/slack/interactives").
		WithFormField("payload", "not json").
		Expect().
		Status(http.StatusBadRequest).NoContent()
}

func TestInteractiveHandlerWithNoHandlers(t *testing.T) {
	engine := gin.New()

	bot := newBot()
	bot.prepareEngine(engine, false)

	payload, _ := json.Marshal(slack.InteractionCallback{})

	e := getHttpExpect(t, engine)
	e.POST("/slack/interactives").
		WithFormField("payload", string(payload)).
		Expect().
		Status(http.StatusOK).NoContent()
}

func TestSelectMenuOptionsHandlerWithNoPayload(t *testing.T) {
	engine := gin.New()

	bot := newBot()
	bot.prepareEngine(engine, false)

	e := getHttpExpect(t, engine)
	e.POST("/slack/menus").
		Expect().
		Status(http.StatusBadRequest).NoContent()
}

func TestSelectMenuOptionsHandlerWithBadPayload(t *testing.T) {
	engine := gin.New()

	bot := newBot()
	bot.prepareEngine(engine, false)

	e := getHttpExpect(t, engine)
	e.POST("/slack/menus").
		WithFormField("payload", "not json").
		Expect().
		Status(http.StatusBadRequest).NoContent()
}

func TestSelectMenuOptionsHandlerWithBadType(t *testing.T) {
	engine := gin.New()

	bot := newBot()
	bot.prepareEngine(engine, false)

	payload, _ := json.Marshal(slack.InteractionCallback{
		Type:       "not a real type",
		CallbackID: "callback1",
	})

	e := getHttpExpect(t, engine)
	e.POST("/slack/menus").
		WithFormField("payload", string(payload)).
		Expect().
		Status(http.StatusBadRequest).NoContent()
}

func TestSelectOptionsHandlerWithOptions(t *testing.T) {
	engine := gin.New()

	bot := newBot()
	bot.RegisterSelectOptions("callback1", func(bot *Bot) slack.OptionsResponse {
		return slack.OptionsResponse{Options: []*slack.OptionBlockObject{&slack.OptionBlockObject{Value: "callback1"}}}
	})
	bot.prepareEngine(engine, false)

	payload, _ := json.Marshal(slack.InteractionCallback{
		Type:       slack.InteractionTypeInteractionMessage,
		CallbackID: "callback1",
	})

	e := getHttpExpect(t, engine)
	e.POST("/slack/menus").
		WithFormField("payload", string(payload)).
		Expect().
		Status(http.StatusOK).JSON().Object().Value("options").Array().First().Object().ValueEqual("value", "callback1")
}

func TestSelectOptionsHandlerWithOptionsGroup(t *testing.T) {
	engine := gin.New()

	bot := newBot()
	bot.RegisterSelectOptionGroups("callback1", func(bot *Bot) slack.OptionGroupsResponse {
		return slack.OptionGroupsResponse{OptionGroups: []*slack.OptionGroupBlockObject{&slack.OptionGroupBlockObject{Options: []*slack.OptionBlockObject{&slack.OptionBlockObject{Value: "callback1"}}}}}
	})
	bot.prepareEngine(engine, false)

	payload, _ := json.Marshal(slack.InteractionCallback{
		Type:       slack.InteractionTypeInteractionMessage,
		CallbackID: "callback1",
	})

	e := getHttpExpect(t, engine)
	e.POST("/slack/menus").
		WithFormField("payload", string(payload)).
		Expect().
		Status(http.StatusOK).JSON().Object().Value("option_groups").Array().First().Object().Value("options").Array().First().Object().ValueEqual("value", "callback1")
}

func TestSelectOptionsHandlerWithBadCallback(t *testing.T) {
	engine := gin.New()

	bot := newBot()
	bot.registerSelectOptions("callback1", func() {})
	bot.prepareEngine(engine, false)

	payload, _ := json.Marshal(slack.InteractionCallback{
		Type:       slack.InteractionTypeInteractionMessage,
		CallbackID: "callback1",
	})

	e := getHttpExpect(t, engine)
	e.POST("/slack/menus").
		WithFormField("payload", string(payload)).
		Expect().
		Status(http.StatusInternalServerError).NoContent()
}
