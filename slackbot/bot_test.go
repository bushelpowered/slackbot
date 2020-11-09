package slackbot

import (
	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"net/http"
	"reflect"
	"testing"
	"time"
)
import "github.com/stretchr/testify/assert"

func TestConstruction(t *testing.T) {
	tokenVal := "token"
	signingSecretVal := "secret"

	bot := newBot()

	assert.Equal(t, tokenVal, bot.token)
	assert.Equal(t, signingSecretVal, bot.signingSecret)
}

func newBot() *Bot {
	tokenVal := "token"
	signingSecretVal := "secret"

	return NewBot(tokenVal, signingSecretVal)
}

func TestSetLogger(t *testing.T) {
	bot := newBot()

	newLogger := logrus.New()
	bot.SetLogger(newLogger)

	assert.Equal(t, newLogger, bot.logger)
}

func TestGetLoggerReturnsStandardLoggerWhenUnset(t *testing.T) {
	bot := newBot()

	assert.Equal(t, logrus.StandardLogger(), bot.getLogger())
}

func TestGetLoggerReturnsSetLogger(t *testing.T) {
	bot := newBot()

	newLogger := logrus.New()
	bot.SetLogger(newLogger)

	assert.Equal(t, newLogger, bot.getLogger())
}

func TestRegisterCommand(t *testing.T) {
	bot := newBot()

	test1Callback := func(b *Bot, command slack.SlashCommand) (*slack.Msg) {
		return nil
	}
	bot.RegisterCommand("command1", test1Callback)
	assert.Equal(t, reflect.ValueOf(test1Callback).Pointer(), reflect.ValueOf(bot.commands["command1"]).Pointer())

	test2Callback := func(b *Bot, command slack.SlashCommand) (*slack.Msg) {
		return nil
	}
	bot.RegisterCommand("command2", test2Callback)
	assert.Equal(t, reflect.ValueOf(test2Callback).Pointer(), reflect.ValueOf(bot.commands["command2"]).Pointer())
}

func TestRegisterEvent(t *testing.T) {
	bot := newBot()

	test1Callback := func(b *Bot, event slackevents.EventsAPIEvent) {}
	bot.RegisterEvent(slackevents.Message, test1Callback)
	assert.Equal(t, reflect.ValueOf(test1Callback).Pointer(), reflect.ValueOf(bot.events[slackevents.Message][0]).Pointer())

	test2Callback := func(b *Bot, event slackevents.EventsAPIEvent) {}
	bot.RegisterEvent(slackevents.Message, test2Callback)
	assert.Equal(t, reflect.ValueOf(test2Callback).Pointer(), reflect.ValueOf(bot.events[slackevents.Message][1]).Pointer())

	test3Callback := func(b *Bot, event slackevents.EventsAPIEvent) {}
	bot.RegisterEvent(slackevents.AppMention, test3Callback)
	assert.Equal(t, reflect.ValueOf(test3Callback).Pointer(), reflect.ValueOf(bot.events[slackevents.AppMention][0]).Pointer())
}

func TestRegisterKeyword(t *testing.T) {
	bot := newBot()

	bot.RegisterKeyword("keyword", func(b *Bot, command slackevents.MessageEvent) {})

	assert.Equal(t, 1, len(bot.events[slackevents.Message]))
}

func TestKeywordCallbackMatch(t *testing.T) {
	keyword := "keyword"
	text := "this text contains the keyword"

	matched := false
	callback := newKeywordEventCallback(keyword, func(b *Bot, command slackevents.MessageEvent) {
		matched = true
	})

	callback(newBot(), newMessageEvent(text))

	assert.True(t, matched)
}

func TestKeywordCallbackMiss(t *testing.T) {
	keyword := "keyword"
	text := "this text does not contain the special word"

	matched := false
	callback := newKeywordEventCallback(keyword, func(b *Bot, command slackevents.MessageEvent) {
		matched = true
	})

	callback(newBot(), newMessageEvent(text))

	assert.False(t, matched)
}

func TestKeywordCallbackIsCaseInsensitive(t *testing.T) {
	keyword := "Keyword"
	text := "this text contains the keyWorD"

	matched := false
	callback := newKeywordEventCallback(keyword, func(b *Bot, command slackevents.MessageEvent) {
		matched = true
	})

	callback(newBot(), newMessageEvent(text))

	assert.True(t, matched)
}

func newMessageEvent(text string) slackevents.EventsAPIEvent {
	return slackevents.EventsAPIEvent{
		Type: slackevents.CallbackEvent,
		InnerEvent: slackevents.EventsAPIInnerEvent{
			Type: slackevents.Message,
			Data: &slackevents.MessageEvent{
				Type: slackevents.Message,
				Text: text,
			},
		},
	}
}

func TestRegisterInteractive(t *testing.T) {
	bot := newBot()

	testCallback1 := func(b *Bot, interaction slack.InteractionCallback) {}
	bot.RegisterInteractive(slack.InteractionTypeBlockActions, "test_action_1", testCallback1)
	assert.Equal(t, reflect.ValueOf(testCallback1).Pointer(), reflect.ValueOf(bot.interactives[slack.InteractionTypeBlockActions]["test_action_1"]).Pointer())

	testCallback2 := func(b *Bot, interaction slack.InteractionCallback) {}
	bot.RegisterInteractive(slack.InteractionTypeBlockActions, "test_action_2", testCallback2)
	assert.Equal(t, reflect.ValueOf(testCallback2).Pointer(), reflect.ValueOf(bot.interactives[slack.InteractionTypeBlockActions]["test_action_2"]).Pointer())
}

func TestRegisterSelectOptions(t *testing.T) {
	bot := newBot()

	options1 := func(b *Bot) slack.OptionsResponse {
		return slack.OptionsResponse{}
	}
	bot.RegisterSelectOptions("action_1", options1)
	assert.Equal(t, reflect.ValueOf(options1).Pointer(), reflect.ValueOf(bot.selectOptions["action_1"]).Pointer())

	options2 := func(b *Bot) slack.OptionsResponse {
		return slack.OptionsResponse{}
	}
	bot.RegisterSelectOptions("action_2", options2)
	assert.Equal(t, reflect.ValueOf(options2).Pointer(), reflect.ValueOf(bot.selectOptions["action_2"]).Pointer())
}

func TestRegisterSelectOptionGroups(t *testing.T) {
	bot := newBot()

	options1 := func(b *Bot) slack.OptionGroupsResponse {
		return slack.OptionGroupsResponse{}
	}
	bot.RegisterSelectOptionGroups("action_1", options1)
	assert.Equal(t, reflect.ValueOf(options1).Pointer(), reflect.ValueOf(bot.selectOptions["action_1"]).Pointer())

	options2 := func(b *Bot) slack.OptionGroupsResponse {
		return slack.OptionGroupsResponse{}
	}
	bot.RegisterSelectOptionGroups("action_2", options2)
	assert.Equal(t, reflect.ValueOf(options2).Pointer(), reflect.ValueOf(bot.selectOptions["action_2"]).Pointer())
}

func TestBoot(t *testing.T) {
	bot := newBot()

	err := bot.Boot(":51356")
	defer bot.Shutdown()
	assert.NoError(t, err)

	err = bot.Boot(":51356")
	assert.EqualError(t, err, ErrAlreadyBooted.Error())

	client := http.Client{Timeout: time.Second}
	resp, err := client.Get("http://localhost:51356")

	assert.NoError(t, err)
	if resp != nil {
		assert.Equal(t, 404, resp.StatusCode)
		_ = resp.Body.Close()
	}

	bot.Shutdown()
}
