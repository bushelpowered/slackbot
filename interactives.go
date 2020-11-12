package slackbot

import (
	"github.com/slack-go/slack"
)

type InteractionCallback = func(bot *Bot, event slack.InteractionCallback)
type ViewSubmissionInteractionCallback = func(bot *Bot, event slack.InteractionCallback) *slack.ViewSubmissionResponse

// Register a callback for message_action interactions with a specific callbackId
func (b *Bot) RegisterMessageActionInteraction(callbackId string, callback InteractionCallback) {
	b.registerInteractive(slack.InteractionTypeMessageAction, func(bot *Bot, interaction slack.InteractionCallback) (response interface{}) {
		if interaction.CallbackID == callbackId {
			callback(b, interaction)
		}
		return nil
	})
}

// Register a callback for shortcut interactions with a specific callbackId
func (b *Bot) RegisterShortcutInteraction(callbackId string, callback InteractionCallback) {
	b.registerInteractive(slack.InteractionTypeShortcut, func(bot *Bot, interaction slack.InteractionCallback) (response interface{}) {
		if interaction.CallbackID == callbackId {
			callback(b, interaction)
		}
		return nil
	})
}

// Filter for RegisterBlockActionsInteraction
//   - specify an ActionID to filter for only certain actions
//   - specify a BlockID for filter for only certain blocks
type BlockActionFilter struct {
	ActionID string
	BlockID  string
}

// Register a callback for block_actions interactions with a specified BlockActionFilter
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

// Register a callback for view_submission interactions with a specific callbackId
// Callback may return a slack.ViewSubmissionResponse or nil for no response
func (b *Bot) RegisterViewSubmissionInteraction(callbackId string, callback ViewSubmissionInteractionCallback) {
	b.registerInteractive(slack.InteractionTypeViewSubmission, func(bot *Bot, interaction slack.InteractionCallback) (response interface{}) {
		if interaction.View.CallbackID == callbackId {
			return callback(b, interaction)
		}
		return nil
	})
}

// Register a callback for view_closed interactions with a specific callbackId
func (b *Bot) RegisterViewClosedInteraction(callbackId string, callback InteractionCallback) {
	b.registerInteractive(slack.InteractionTypeViewClosed, func(bot *Bot, interaction slack.InteractionCallback) (response interface{}) {
		if interaction.View.CallbackID == callbackId {
			callback(b, interaction)
		}
		return nil
	})
}
