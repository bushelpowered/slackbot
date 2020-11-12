package slackbot

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestRegisterMessageActionInteraction(t *testing.T) {
	engine := gin.New()

	bot := newBot()
	callback1Hit1 := false
	bot.RegisterMessageActionInteraction("callback1", func(bot *Bot, event slack.InteractionCallback) {
		callback1Hit1 = true
	})
	callback1Hit2 := false
	bot.RegisterMessageActionInteraction("callback1", func(bot *Bot, event slack.InteractionCallback) {
		callback1Hit2 = true
	})
	callback3Hit := false
	bot.RegisterMessageActionInteraction("callback2", func(bot *Bot, event slack.InteractionCallback) {
		callback3Hit = true
	})
	bot.prepareEngine(engine, false)

	payload, _ := json.Marshal(slack.InteractionCallback{
		Type:       slack.InteractionTypeMessageAction,
		CallbackID: "callback1",
	})

	e := getHttpExpect(t, engine)
	e.POST("/slack/interactives").
		WithFormField("payload", string(payload)).
		Expect().
		Status(http.StatusOK).NoContent()

	assert.True(t, callback1Hit1)
	assert.True(t, callback1Hit2)
	assert.False(t, callback3Hit)
}

func TestRegisterShortcutInteraction(t *testing.T) {
	engine := gin.New()

	bot := newBot()
	callback1Hit := false
	bot.RegisterShortcutInteraction("callback1", func(bot *Bot, event slack.InteractionCallback) {
		callback1Hit = true
	})
	callback2Hit := false
	bot.RegisterShortcutInteraction("callback2", func(bot *Bot, event slack.InteractionCallback) {
		callback2Hit = true
	})
	bot.prepareEngine(engine, false)

	payload, _ := json.Marshal(slack.InteractionCallback{
		Type:       slack.InteractionTypeShortcut,
		CallbackID: "callback1",
	})

	e := getHttpExpect(t, engine)
	e.POST("/slack/interactives").
		WithFormField("payload", string(payload)).
		Expect().
		Status(http.StatusOK).NoContent()

	assert.True(t, callback1Hit)
	assert.False(t, callback2Hit)
}

func TestRegisterBlockActionsInteraction(t *testing.T) {
	engine := gin.New()

	bot := newBot()
	hits := make(map[string]bool)
	bot.RegisterBlockActionsInteraction(BlockActionFilter{ActionID: "action1"}, func(bot *Bot, event slack.InteractionCallback) {
		hits["onlyAction1"] = true
	})
	bot.RegisterBlockActionsInteraction(BlockActionFilter{ActionID: "action2"}, func(bot *Bot, event slack.InteractionCallback) {
		hits["onlyAction2"] = true
	})
	bot.RegisterBlockActionsInteraction(BlockActionFilter{BlockID: "block1"}, func(bot *Bot, event slack.InteractionCallback) {
		hits["onlyBlock1"] = true
	})
	bot.RegisterBlockActionsInteraction(BlockActionFilter{BlockID: "block2"}, func(bot *Bot, event slack.InteractionCallback) {
		hits["onlyBlock2"] = true
	})
	bot.RegisterBlockActionsInteraction(BlockActionFilter{ActionID: "action1", BlockID: "block1"}, func(bot *Bot, event slack.InteractionCallback) {
		hits["bothAction1AndBlock1"] = true
	})
	bot.RegisterBlockActionsInteraction(BlockActionFilter{ActionID: "action1", BlockID: "block2"}, func(bot *Bot, event slack.InteractionCallback) {
		hits["bothAction1AndBlock2"] = true
	})
	bot.RegisterBlockActionsInteraction(BlockActionFilter{ActionID: "action2", BlockID: "block1"}, func(bot *Bot, event slack.InteractionCallback) {
		hits["bothAction2AndBlock1"] = true
	})
	bot.prepareEngine(engine, false)

	payload, _ := json.Marshal(slack.InteractionCallback{
		Type: slack.InteractionTypeBlockActions,
		ActionCallback: slack.ActionCallbacks{
			BlockActions: []*slack.BlockAction{&slack.BlockAction{ActionID: "action1", BlockID: "block1"}},
		},
	})

	e := getHttpExpect(t, engine)
	e.POST("/slack/interactives").
		WithFormField("payload", string(payload)).
		Expect().
		Status(http.StatusOK).NoContent()

	assert.True(t, hits["onlyAction1"])
	assert.False(t, hits["onlyAction2"])
	assert.True(t, hits["onlyBlock1"])
	assert.False(t, hits["onlyBlock2"])
	assert.True(t, hits["bothAction1AndBlock1"])
	assert.False(t, hits["bothAction1AndBlock2"])
	assert.False(t, hits["bothAction2AndBlock1"])
}

func TestRegisterViewSubmissionInteractionWithoutResponse(t *testing.T) {
	engine := gin.New()

	bot := newBot()
	callback1Hit := false
	bot.RegisterViewSubmissionInteraction("callback1", func(bot *Bot, event slack.InteractionCallback) *slack.ViewSubmissionResponse {
		callback1Hit = true
		return nil
	})
	callback2Hit := false
	bot.RegisterViewSubmissionInteraction("callback2", func(bot *Bot, event slack.InteractionCallback) *slack.ViewSubmissionResponse {
		callback2Hit = true
		return nil
	})
	bot.prepareEngine(engine, false)

	payload, _ := json.Marshal(slack.InteractionCallback{
		Type: slack.InteractionTypeViewSubmission,
		View: slack.View{
			CallbackID: "callback1",
		},
	})

	e := getHttpExpect(t, engine)
	e.POST("/slack/interactives").
		WithFormField("payload", string(payload)).
		Expect().
		Status(http.StatusOK).NoContent()

	assert.True(t, callback1Hit)
	assert.False(t, callback2Hit)
}

func TestRegisterViewSubmissionInteractionWithResponse(t *testing.T) {
	engine := gin.New()

	bot := newBot()
	bot.RegisterViewSubmissionInteraction("callback1", func(bot *Bot, event slack.InteractionCallback) *slack.ViewSubmissionResponse {
		return &slack.ViewSubmissionResponse{
			ResponseAction: slack.RAClear,
		}
	})
	bot.prepareEngine(engine, false)

	payload, _ := json.Marshal(slack.InteractionCallback{
		Type: slack.InteractionTypeViewSubmission,
		View: slack.View{
			CallbackID: "callback1",
		},
	})

	e := getHttpExpect(t, engine)
	e.POST("/slack/interactives").
		WithFormField("payload", string(payload)).
		Expect().
		Status(http.StatusOK).JSON().Object().ValueEqual("response_action", slack.RAClear)
}

func TestRegisterViewClosedInteraction(t *testing.T) {
	engine := gin.New()

	bot := newBot()
	callback1Hit := false
	bot.RegisterViewClosedInteraction("callback1", func(bot *Bot, event slack.InteractionCallback) {
		callback1Hit = true
	})
	callback2Hit := false
	bot.RegisterViewClosedInteraction("callback2", func(bot *Bot, event slack.InteractionCallback) {
		callback2Hit = true
	})
	bot.prepareEngine(engine, false)

	payload, _ := json.Marshal(slack.InteractionCallback{
		Type: slack.InteractionTypeViewClosed,
		View: slack.View{
			CallbackID: "callback1",
		},
	})

	e := getHttpExpect(t, engine)
	e.POST("/slack/interactives").
		WithFormField("payload", string(payload)).
		Expect().
		Status(http.StatusOK).NoContent()

	assert.True(t, callback1Hit)
	assert.False(t, callback2Hit)
}
