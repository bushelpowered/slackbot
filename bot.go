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
	log    *logrus.Logger

	commands      map[string]CommandCallback
	events        map[string][]eventCallback
	interactives  map[slack.InteractionType][]interactiveCallback
	selectOptions map[string]interface{}

	sync.RWMutex
}

// Create a new Bot with given bot token and app signing secret
func NewBot(token string, signingSecret string) *Bot {
	return &Bot{
		token:         token,
		signingSecret: signingSecret,
	}
}

// Get a slack.Client for interacting with the API
func (b *Bot) Api() *slack.Client {
	return slack.New(b.token)
}

func (b *Bot) logger() *logrus.Logger {
	if b.log == nil {
		b.Lock()
		defer b.Unlock()
		if b.log == nil {
			b.log = logrus.StandardLogger()
		}
	}
	return b.log
}

// Set a pre-configured logrus.Logger to provide a consistent logging experience in your bot.
func (b *Bot) SetLogger(logger *logrus.Logger) {
	b.Lock()
	defer b.Unlock()

	b.log = logger
}

// Register a slash command callback
func (b *Bot) RegisterCommand(name string, callback CommandCallback) {
	b.logger().Debugf("RegisterCommand %s", name)

	b.Lock()
	defer b.Unlock()

	if b.commands == nil {
		b.commands = make(map[string]CommandCallback)
	}

	b.commands[name] = callback
}

func (b *Bot) registerEvent(eventType string, callback eventCallback) {
	b.logger().Debugf("RegisterEvent %s", eventType)

	b.Lock()
	defer b.Unlock()

	if b.events == nil {
		b.events = make(map[string][]eventCallback)
	}
	b.events[eventType] = append(b.events[eventType], callback)
}

// Register a message event keyword regex callback.
func (b *Bot) RegisterKeyword(regex *regexp.Regexp, callback KeywordCallback) {
	b.logger().Debugf("RegisterKeyword %s", regex)
	b.RegisterMessageEvent(b.newKeywordEventCallback(regex, callback))
}

func (b *Bot) newKeywordEventCallback(regex *regexp.Regexp, callback KeywordCallback) MessageEventCallback {
	return func(bot *Bot, c MessageEventContainer) {
		if regex.FindString(c.Event.Text) != "" {
			callback(b, c)
		}
	}
}

func (b *Bot) registerInteractive(interactionType slack.InteractionType, callback interactiveCallback) {
	b.logger().Debugf("RegisterInteractive %s", interactionType)

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

// Register a select options callback
func (b *Bot) RegisterSelectOptions(callbackId string, callback SelectMenuOptionsCallback) {
	b.logger().Debugf("RegisterSelectOptions %s", callbackId)

	b.registerSelectOptions(callbackId, callback)
}

// Register a select option groups callback
func (b *Bot) RegisterSelectOptionGroups(callbackId string, callback SelectMenuOptionsGroupCallback) {
	b.logger().Debugf("RegisterSelectOptions %s", callbackId)

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

// Start the bot on the given listen address
func (b *Bot) Boot(listenAddr string) error {
	engine := gin.New()
	engine.Use(gin.Recovery())

	return b.BootWithEngine(listenAddr, engine)
}

// Start the bot on the given listen address with a pre-configured instance of gin.Engine.
func (b *Bot) BootWithEngine(listenAddr string, engine *gin.Engine) error {
	b.logger().Infof("Booting slackbot on %s", listenAddr)

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
			b.logger().WithError(err).Fatalln("Failed to start server")
		}
	}()

	return nil
}

func (b *Bot) prepareEngine(engine *gin.Engine, verify bool) {
	engine.Use(ginlogrus.Logger(b.logger()))

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
		b.logger().Infof("Wired command \"%s\" to %s/commands/%s", name, group.BasePath(), name)
		group.POST("/commands/"+name, b.newCommandHandler(callback))
	}
}

func (b *Bot) wireEvents(group *gin.RouterGroup) {
	b.logger().Infof("Wired events to %s/events", group.BasePath())
	group.POST("/events", b.newEventHandler())
}

func (b *Bot) wireInteractives(group *gin.RouterGroup) {
	b.logger().Infof("Wired interactives to %s/interactives", group.BasePath())
	group.POST("/interactives", b.newInteractiveHandler())
}

func (b *Bot) wireSelectMenus(group *gin.RouterGroup) {
	b.logger().Infof("Wired select menus to %s/menus", group.BasePath())
	group.POST("/menus", b.newSelectMenusHandler())
}

// Shutdown the bot gracefully with a given timeout
func (b *Bot) Shutdown(timeout time.Duration) {
	b.Lock()
	defer b.Unlock()

	if b.server == nil {
		return
	}

	defer func() {
		b.server = nil
	}()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := b.server.Shutdown(ctx); err != nil {
		b.logger().WithError(err).Fatalln("Server forced to shutdown")
	}
}
