package main

import (
	"github.com/line/line-bot-sdk-go/linebot"
	"time"
)

type ReceivedMessage struct {
	Type linebot.EventType		`json:"type"`
	Message *MessageBody		`json:"message"`
	Source linebot.EventSource	`json:"source"`
	Timestamp time.Time			`json:"timestamp"`
	ReplyToken string			`json:"replyToken"`
}

type MessageBody struct {
	ID string					`json:"id"`
	FileName *string			`json:"fileName"`
}