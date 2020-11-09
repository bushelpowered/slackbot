package slackbot

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/slack-go/slack"
	"io/ioutil"
	"net/http"
)

func (b *Bot) newSlackVerifierMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := b.getLogger().WithField("handler", "slackVerifier")

		logger.Debug("Verifying slack request signature")

		verifier, err := slack.NewSecretsVerifier(c.Request.Header, b.signingSecret)
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		body, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			_ = c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		_, err = verifier.Write(body)
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.Request.Body = ioutil.NopCloser(bytes.NewReader(body))

		err = verifier.Ensure()
		if err != nil {
			_ = c.AbortWithError(http.StatusUnauthorized, err)
			return
		}

		c.Next()
	}
}
