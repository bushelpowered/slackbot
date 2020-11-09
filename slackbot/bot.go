package slackbot

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"github.com/toorop/gin-logrus"
	"net/http"
	"strings"
	"sync"
	"time"
)
import "github.com/slack-go/slack/slackevents"

type CommandCallback = func(bot *Bot, command slack.SlashCommand) *slack.Msg
type EventCallback = func(bot *Bot, event slackevents.EventsAPIEvent)
type KeywordCallback = func(abot *Bot, command slackevents.MessageEvent)
type InteractiveCallback = func(bot *Bot, interaction slack.InteractionCallback)
type SelectMenuOptionsCallback = func(bot *Bot) slack.OptionsResponse
type SelectMenuOptionsGroupCallback = func(bot *Bot) slack.OptionGroupsResponse

type Bot struct {
	token         string
	signingSecret string

	server *http.Server
	logger *logrus.Logger

	commands      map[string]CommandCallback
	events        map[string][]EventCallback
	interactives  map[slack.InteractionType]map[string]InteractiveCallback
	selectOptions map[string]interface{}

	sync.Mutex
}

func NewBot(token string, signingSecret string) *Bot {
	return &Bot{
		token:         token,
		signingSecret: signingSecret,
	}
}

func (b *Bot) Api() *slack.Client {
	return slack.New(b.token)
}

func (b *Bot) getLogger() *logrus.Logger {
	if b.logger == nil {
		b.Lock()
		defer b.Unlock()
		if b.logger == nil {
			b.logger = logrus.StandardLogger()
		}
	}
	return b.logger
}

func (b *Bot) SetLogger(logger *logrus.Logger) {
	b.Lock()
	defer b.Unlock()

	b.logger = logger
}

func (b *Bot) RegisterCommand(name string, callback CommandCallback) {
	b.getLogger().Debugf("RegisterCommand %s", name)

	b.Lock()
	defer b.Unlock()

	if b.commands == nil {
		b.commands = make(map[string]CommandCallback)
	}

	b.commands[name] = callback
}

func (b *Bot) RegisterEvent(eventType string, callback EventCallback) {
	b.getLogger().Debugf("RegisterEvent %s", eventType)

	b.Lock()
	defer b.Unlock()

	if b.events == nil {
		b.events = make(map[string][]EventCallback)
	}
	b.events[eventType] = append(b.events[eventType], callback)
}

func (b *Bot) RegisterKeyword(keyword string, callback KeywordCallback) {
	b.getLogger().Debugf("RegisterKeyword %s", keyword)

	b.RegisterEvent(slackevents.Message, newKeywordEventCallback(keyword, callback))
}

func newKeywordEventCallback(keyword string, callback KeywordCallback) EventCallback {
	return func(b *Bot, event slackevents.EventsAPIEvent) {
		switch ev := event.InnerEvent.Data.(type) {
		case *slackevents.MessageEvent:
			if strings.Contains(strings.ToLower(ev.Text), strings.ToLower(keyword)) {
				callback(b, *ev)
			}
		}
	}
}

func (b *Bot) RegisterInteractive(interactionType slack.InteractionType, filterId string, callback InteractiveCallback) {
	b.getLogger().Debugf("RegisterInteractive %s:%s", interactionType, filterId)

	b.Lock()
	defer b.Unlock()

	if b.interactives == nil {
		b.interactives = make(map[slack.InteractionType]map[string]InteractiveCallback)
	}
	if _, exists := b.interactives[interactionType]; !exists {
		b.interactives[interactionType] = make(map[string]InteractiveCallback)
	}
	b.interactives[interactionType][filterId] = callback
}

func (b *Bot) RegisterSelectOptions(actionId string, callback SelectMenuOptionsCallback) {
	b.getLogger().Debugf("RegisterSelectOptions %s", actionId)

	b.registerSelectOptions(actionId, callback)
}

func (b *Bot) RegisterSelectOptionGroups(actionId string, callback SelectMenuOptionsGroupCallback) {
	b.getLogger().Debugf("RegisterSelectOptions %s", actionId)

	b.registerSelectOptions(actionId, callback)
}

func (b *Bot) registerSelectOptions(actionId string, callback interface{}) {
	b.Lock()
	defer b.Unlock()

	if b.selectOptions == nil {
		b.selectOptions = make(map[string]interface{})
	}

	b.selectOptions[actionId] = callback
}

func (b *Bot) Boot(listenAddr string) error {
	engine := gin.New()
	engine.Use(gin.Recovery())

	return b.BootWithEngine(listenAddr, engine)
}

func (b *Bot) BootWithEngine(listenAddr string, engine *gin.Engine) error {
	b.getLogger().Infof("Booting slackbot on %s", listenAddr)

	b.Lock()
	defer b.Unlock()

	if b.server != nil {
		return ErrAlreadyBooted
	}

	engine.Use(ginlogrus.Logger(b.getLogger()))

	slackGroup := engine.Group("/slack")
	slackGroup.Use(b.newSlackVerifierMiddleware())

	b.wireCallbacks(slackGroup)

	b.server = &http.Server{
		Addr:    listenAddr,
		Handler: engine,
	}

	go func() {
		if err := b.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			b.getLogger().WithError(err).Fatalln("Failed to start server")
		}
	}()

	return nil
}

func (b *Bot) wireCallbacks(group *gin.RouterGroup) {
	b.wireCommands(group)
	b.wireEvents(group)
	b.wireInteractives(group)
	b.wireSelectMenus(group)
}

func (b *Bot) wireCommands(group *gin.RouterGroup) {
	for name, callback := range b.commands {
		b.getLogger().Infof("Wired command \"%s\" to %s/commands/%s", name, group.BasePath(), name)
		group.POST("/commands/" + name, b.newCommandHandler(callback))
	}
}

func (b *Bot) wireEvents(group *gin.RouterGroup) {

}

func (b *Bot) wireInteractives(group *gin.RouterGroup) {

}

func (b *Bot) wireSelectMenus(group *gin.RouterGroup) {

}

func (b *Bot) Shutdown() {
	b.Lock()
	defer b.Unlock()

	if b.server == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := b.server.Shutdown(ctx); err != nil {
		b.getLogger().WithError(err).Fatalln("Server forced to shutdown")
	}

	b.server = nil
}
