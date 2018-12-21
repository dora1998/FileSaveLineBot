package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/line/line-bot-sdk-go/linebot/httphandler"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"log"
	"net/http"
	"os"
)

const MESSEAGE_GROUP_JOINED="こんにちは！しおパンダbotだぱん！\nファイルが送信されたら、Dropboxに自動で取っておくから是非使って欲しいぱん。"

var botHandler *httphandler.WebhookHandler

func main() {
	err := godotenv.Load("line.env")
	if err != nil {
		panic(err)
	}

	botHandler, err = httphandler.New(
		os.Getenv("LINE_BOT_CHANNEL_SECRET"),
		os.Getenv("LINE_BOT_CHANNEL_TOKEN"),
	)
	botHandler.HandleEvents(handleCallback)

	http.Handle("/callback", botHandler)
	http.HandleFunc("/task", handleTask)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

// Webhook を受け取り、TaskQueueに流す
func handleCallback(evs []*linebot.Event, r *http.Request) {
	c := newContext(r)
	bot, err := newLINEBot(c)
	if err != nil {
		errorf(c, "newLINEBot: %v", err)
		return
	}

	for _, e := range evs {
		logf(c, "Webhook recieved.\nEventType: %s\nMesseage: %#v", e.Type, e.Message)

		switch e.Type {
		case linebot.EventTypeJoin:
			// グループへ投入された時はメッセージをすぐ返す
			_, err := bot.ReplyMessage(
				e.ReplyToken,
				linebot.NewTextMessage(MESSEAGE_GROUP_JOINED),
			).Do()

			if err != nil {
				errorf(c, "ReplayMessage: %v", err)
				continue
			}

			taskData := &ReceivedMessage{
				Type: e.Type,
				Message: nil,
				Source: *e.Source,
				Timestamp: e.Timestamp,
				ReplyToken: e.ReplyToken,
			}
			replyMessage(c, bot, taskData)
		case linebot.EventTypeMessage:
			var taskData *ReceivedMessage = nil

			// 画像とファイル送信のみ反応し、TaskQueueへ投げるデータを生成
			switch mes := e.Message.(type) {
			case *linebot.FileMessage:
				logf(c, "FileMessage Received.")
				taskData = &ReceivedMessage{
					Type: e.Type,
					Message: &MessageBody{
						ID: mes.ID,
						FileName: &mes.FileName,
					},
					Source: *e.Source,
					Timestamp: e.Timestamp,
					ReplyToken: e.ReplyToken,
				}
			case *linebot.ImageMessage:
				logf(c, "ImageMessage Received.")
				taskData = &ReceivedMessage{
					Type: e.Type,
					Message: &MessageBody{
						ID: mes.ID,
						FileName: nil,
					},
					Source: *e.Source,
					Timestamp: e.Timestamp,
					ReplyToken: e.ReplyToken,
				}
			}

			// 保存処理をするメッセージだけTaskQueueに投げる
			// TODO: 現在TaskQueueにAddしても実行されないため、直接処理中
			if taskData != nil {
				//j, err := json.Marshal(taskData)
				//if err != nil {
				//	errorf(c, "json.Marshal: %v", err)
				//	return
				//}
				//b64data := base64.StdEncoding.EncodeToString(j)
				//t := taskqueue.NewPOSTTask("/task", map[string][]string{"data": {b64data}})
				//taskqueue.Add(c, t, "")
				//logf(c, "TaskQueue Sent.")
				replyMessage(c, bot, taskData)
			}
		}
	}
}

func logf(c context.Context, format string, args ...interface{}) {
	log.Printf(format, args...)
}

func errorf(c context.Context, format string, args ...interface{}) {
	log.Fatalf(format, args...)
}

func newContext(r *http.Request) context.Context {
	return appengine.NewContext(r)
}

func newLINEBot(c context.Context) (*linebot.Client, error) {
	//return botHandler.NewClient(
	//	linebot.WithHTTPClient(urlfetch.Client(c)),
	//)
	return linebot.New(
		os.Getenv("LINE_BOT_CHANNEL_SECRET"),
		os.Getenv("LINE_BOT_CHANNEL_TOKEN"),
	)
}

func isDevServer() bool {
	return os.Getenv("RUN_WITH_DEVAPPSERVER") != ""
}