package slackbot

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin/render"
	"github.com/slack-go/slack"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestSlackVerifierReturnsServerErrorWhenSecretsVerifierFailsInitialization(t *testing.T) {
	bot := newBot()

	engine := gin.New()
	engine.Use(bot.newSlackVerifierMiddleware())
	engine.POST("/test", func(context *gin.Context) {
		context.Render(http.StatusOK, render.Data{ContentType: "application/json; charset=utf-8", Data: []byte("")})
	})

	e := getHttpExpect(t, engine)

	// Assert response
	e.POST("/test").
		WithJSON(slack.SlashCommand{}).
		Expect().
		Status(http.StatusInternalServerError)
}

func TestSlackVerifierSucceedsWhenSignaturesMatch(t *testing.T) {
	requestTimestamp := fmt.Sprintf("%d", time.Now().Unix())
	signingSecret := "e6b19c573432dcc6b075501d51b51bb8"

	bot := NewBot("token", signingSecret)

	// create a request
	request := slack.SlashCommand{}
	requestBytes, _ := json.Marshal(request)

	// create the signature for the request
	hash := hmac.New(sha256.New, []byte(signingSecret))
	hash.Write([]byte(fmt.Sprintf("v0:%s:", requestTimestamp)))
	hash.Write(requestBytes)
	computedHash := hex.EncodeToString(hash.Sum(nil))

	// set up gin with the verifier middleware
	engine := gin.New()
	engine.Use(bot.newSlackVerifierMiddleware())
	engine.POST("/test", func(context *gin.Context) {
		context.Status(http.StatusOK)
	})

	// make a request with valid signature headers
	e := getHttpExpect(t, engine)
	e.POST("/test").
		WithHeader("X-Slack-Signature", fmt.Sprintf("v0=%s", computedHash)).
		WithHeader("X-Slack-Request-Timestamp", requestTimestamp).
		WithHeader("Content-Type", "application/json; charset=utf-8").
		WithBytes(requestBytes).
		Expect().
		Status(http.StatusOK)
}

func TestSlackVerifierReturnsUnauthorizedErrorWhenSignaturesDoNotMatch(t *testing.T) {
	requestTimestamp := fmt.Sprintf("%d", time.Now().Unix())

	bot := NewBot("token", "secret1")

	// create a request
	request := slack.SlashCommand{}
	requestBytes, _ := json.Marshal(request)

	// create the signature for the request
	hash := hmac.New(sha256.New, []byte("secret2"))
	hash.Write([]byte(fmt.Sprintf("v0:%s:", requestTimestamp)))
	hash.Write(requestBytes)
	computedHash := hex.EncodeToString(hash.Sum(nil))

	// set up gin with the verifier middleware
	engine := gin.New()
	engine.Use(bot.newSlackVerifierMiddleware())
	engine.POST("/test", func(context *gin.Context) {
		context.Status(http.StatusOK)
	})

	// make a request with valid signature headers
	e := getHttpExpect(t, engine)
	e.POST("/test").
		WithHeader("X-Slack-Signature", fmt.Sprintf("v0=%s", computedHash)).
		WithHeader("X-Slack-Request-Timestamp", requestTimestamp).
		WithHeader("Content-Type", "application/json; charset=utf-8").
		WithBytes(requestBytes).
		Expect().
		Status(http.StatusUnauthorized)
}
