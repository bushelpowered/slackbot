package slackbot

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/toorop/gin-logrus"
	"net/http"
	"regexp"
	"sync"
	"time"
)

//go:generate go run events.go

type CommandCallback = func(bot *Bot, command slack.SlashCommand) *slack.Msg
type eventCallback = func(bot *Bot, event slackevents.EventsAPIEvent)
type KeywordCallback = func(bot *Bot, container MessageEventContainer)
type interactiveCallback = func(bot *Bot, interaction slack.InteractionCallback) (response interface{})
type SelectMenuOptionsCallback = func(bot *Bot, interaction slack.InteractionCallback) slack.OptionsResponse
type SelectMenuOptionsGroupCallback = func(bot *Bot, interaction slack.InteractionCallback) slack.OptionGroupsResponse

type Bot struct {
	token         string
	signingSecret string

	server *http.Server
	logger *logrus.Logger

	commands      map[string]CommandCallback
	events        map[string][]eventCallback
	interactives  map[slack.InteractionType][]interactiveCallback
	selectOptions map[string]interface{}

	sync.RWMutex
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

func (b *Bot) Logger() *logrus.Logger {
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
	b.Logger().Debugf("RegisterCommand %s", name)

	b.Lock()
	defer b.Unlock()

	if b.commands == nil {
		b.commands = make(map[string]CommandCallback)
	}

	b.commands[name] = callback
}

func (b *Bot) registerEvent(eventType string, callback eventCallback) {
	b.Logger().Debugf("RegisterEvent %s", eventType)

	b.Lock()
	defer b.Unlock()

	if b.events == nil {
		b.events = make(map[string][]eventCallback)
	}
	b.events[eventType] = append(b.events[eventType], callback)
}

func (b *Bot) RegisterKeyword(regex *regexp.Regexp, callback KeywordCallback) {
	b.Logger().Debugf("RegisterKeyword %s", regex)

	b.registerEvent(slackevents.Message, newKeywordEventCallback(regex, callback))
}

func newKeywordEventCallback(regex *regexp.Regexp, callback KeywordCallback) eventCallback {
	return func(b *Bot, event slackevents.EventsAPIEvent) {
		switch ev := event.InnerEvent.Data.(type) {
		case *slackevents.MessageEvent:
			if regex.FindString(ev.Text) != "" {
				callback(b, MessageEventContainer{APIEvent: event, Event: *ev})
			}
		}
	}
}

func (b *Bot) registerInteractive(interactionType slack.InteractionType, callback interactiveCallback) {
	b.Logger().Debugf("RegisterInteractive %s", interactionType)

	b.Lock()
	defer b.Unlock()

	if b.interactives == nil {
		b.interactives = make(map[slack.InteractionType][]interactiveCallback)
	}
	if _, exists := b.interactives[interactionType]; !exists {
		b.interactives[interactionType] = make([]interactiveCallback, 0)
	}
	b.interactives[interactionType] = append(b.interactives[interactionType], callback)
}

func (b *Bot) RegisterSelectOptions(callbackId string, callback SelectMenuOptionsCallback) {
	b.Logger().Debugf("RegisterSelectOptions %s", callbackId)

	b.registerSelectOptions(callbackId, callback)
}

func (b *Bot) RegisterSelectOptionGroups(callbackId string, callback SelectMenuOptionsGroupCallback) {
	b.Logger().Debugf("RegisterSelectOptions %s", callbackId)

	b.registerSelectOptions(callbackId, callback)
}

func (b *Bot) registerSelectOptions(callbackId string, callback interface{}) {
	b.Lock()
	defer b.Unlock()

	if b.selectOptions == nil {
		b.selectOptions = make(map[string]interface{})
	}

	b.selectOptions[callbackId] = callback
}

func (b *Bot) Boot(listenAddr string) error {
	engine := gin.New()
	engine.Use(gin.Recovery())

	return b.BootWithEngine(listenAddr, engine)
}

func (b *Bot) BootWithEngine(listenAddr string, engine *gin.Engine) error {
	b.Logger().Infof("Booting slackbot on %s", listenAddr)

	b.Lock()
	defer b.Unlock()

	if b.server != nil {
		return ErrAlreadyBooted
	}

	b.prepareEngine(engine, true)

	b.server = &http.Server{
		Addr:    listenAddr,
		Handler: engine,
	}

	go func() {
		if err := b.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			b.Logger().WithError(err).Fatalln("Failed to start server")
		}
	}()

	return nil
}

func (b *Bot) prepareEngine(engine *gin.Engine, verify bool) {
	engine.Use(ginlogrus.Logger(b.Logger()))

	slackGroup := engine.Group("/slack")
	if verify {
		slackGroup.Use(b.newSlackVerifierMiddleware())
	}

	b.wireCallbacks(slackGroup)
}

func (b *Bot) wireCallbacks(group *gin.RouterGroup) {
	b.wireCommands(group)
	b.wireEvents(group)
	b.wireInteractives(group)
	b.wireSelectMenus(group)
}

func (b *Bot) wireCommands(group *gin.RouterGroup) {
	for name, callback := range b.commands {
		b.Logger().Infof("Wired command \"%s\" to %s/commands/%s", name, group.BasePath(), name)
		group.POST("/commands/"+name, b.newCommandHandler(callback))
	}
}

func (b *Bot) wireEvents(group *gin.RouterGroup) {
	b.Logger().Infof("Wired events to %s/events", group.BasePath())
	group.POST("/events", b.newEventHandler())
}

func (b *Bot) wireInteractives(group *gin.RouterGroup) {
	b.Logger().Infof("Wired interactives to %s/interactives", group.BasePath())
	group.POST("/interactives", b.newInteractiveHandler())
}

func (b *Bot) wireSelectMenus(group *gin.RouterGroup) {
	b.Logger().Infof("Wired select menus to %s/menus", group.BasePath())
	group.POST("/menus", b.newSelectMenusHandler())
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
		b.Logger().WithError(err).Fatalln("Server forced to shutdown")
	}

	b.server = nil
}
