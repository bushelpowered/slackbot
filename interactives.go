package slackbot

import (
	"github.com/slack-go/slack"
)

type InteractionCallback = func(bot *Bot, event slack.InteractionCallback)
type ViewSubmissionInteractionCallback = func(bot *Bot, event slack.InteractionCallback) *slack.ViewSubmissionResponse

func (b *Bot) RegisterMessageActionInteraction(callbackId string, callback InteractionCallback) {
	b.registerInteractive(slack.InteractionTypeMessageAction, func(bot *Bot, interaction slack.InteractionCallback) (response interface{}) {
		if interaction.CallbackID == callbackId {
			callback(b, interaction)
		}
		return nil
	})
}

func (b *Bot) RegisterShortcutInteraction(callbackId string, callback InteractionCallback) {
	b.registerInteractive(slack.InteractionTypeShortcut, func(bot *Bot, interaction slack.InteractionCallback) (response interface{}) {
		if interaction.CallbackID == callbackId {
			callback(b, interaction)
		}
		return nil
	})
}

type BlockActionFilter struct {
	ActionID string
	BlockID  string
}

func (b *Bot) RegisterBlockActionsInteraction(filter BlockActionFilter, callback InteractionCallback) {
	b.registerInteractive(slack.InteractionTypeBlockActions, func(bot *Bot, interaction slack.InteractionCallback) (response interface{}) {
		actions := interaction.ActionCallback.BlockActions
		if len(actions) > 0 {
			action := actions[0]
			actionMatch := (filter.ActionID == "") || (action.ActionID == filter.ActionID)
			blockMatch := (filter.BlockID == "") || (action.BlockID == filter.BlockID)
			if actionMatch && blockMatch {
				callback(b, interaction)
			}
		}
		return nil
	})
}

func (b *Bot) RegisterViewSubmissionInteraction(callbackId string, callback ViewSubmissionInteractionCallback) {
	b.registerInteractive(slack.InteractionTypeViewSubmission, func(bot *Bot, interaction slack.InteractionCallback) (response interface{}) {
		if interaction.View.CallbackID == callbackId {
			return callback(b, interaction)
		}
		return nil
	})
}

func (b *Bot) RegisterViewClosedInteraction(callbackId string, callback InteractionCallback) {
	b.registerInteractive(slack.InteractionTypeViewClosed, func(bot *Bot, interaction slack.InteractionCallback) (response interface{}) {
		if interaction.View.CallbackID == callbackId {
			callback(b, interaction)
		}
		return nil
	})
}
