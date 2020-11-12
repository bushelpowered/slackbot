package slackbot

import "errors"

var ErrAlreadyBooted = errors.New("bot already booted")
var ErrEmptyPayload = errors.New("empty payload")
var ErrBadPayload = errors.New("bad payload")
var ErrUnknownOptionsCallback = errors.New("unknown options callback")
