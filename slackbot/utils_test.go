package slackbot

import (
	"github.com/gavv/httpexpect/v2"
	"github.com/gin-gonic/gin"
	"net/http"
	"testing"
)

func getHttpExpect(t *testing.T, engine *gin.Engine) *httpexpect.Expect {
	return httpexpect.WithConfig(httpexpect.Config{
		Client: &http.Client{
			Transport: httpexpect.NewBinder(engine),
			Jar:       httpexpect.NewJar(),
		},
		Reporter: httpexpect.NewAssertReporter(t),
		Printers: []httpexpect.Printer{
			httpexpect.NewDebugPrinter(t, true),
		},
	})
}