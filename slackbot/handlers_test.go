package slackbot

//func TestCommandWithResponse(t *testing.T) {
//
//
//
//
//
//
//	bootAndTest(t, func(b *Bot) {
//		b.RegisterCommand("test", func(bot *Bot, command slack.SlashCommand) (*slack.Msg, error) {
//			return &slack.Msg{Text: "hello world!"}, nil
//		})
//	}, func(b *Bot) {
//
//
//
//
//
//	})
//}
//
//
//func bootAndTest(t *testing.T, configure func(b *Bot), test func(b *Bot)) {
//	b := newBot()
//	configure(b)
//	err := b.Boot(":51357")
//	assert.NoError(t, err)
//	if err != nil {
//		return
//	}
//	defer b.Shutdown()
//	test(b)
//}
