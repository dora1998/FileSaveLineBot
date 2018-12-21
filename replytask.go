package main

import (
	"encoding/base64"
	"encoding/json"
	"github.com/dora1998/FileSaveLineBot/cloudstrage"
	"github.com/line/line-bot-sdk-go/linebot"
	"golang.org/x/net/context"
	"log"
	"net/http"
	"time"
)

const MESSEAGE_FILE_SAVED="ファイルを保存したぱん！¥r¥n保存期限が切れちゃったら、下のアドレスから見れるぱん。¥r¥n%s"

// 受け取ったメッセージの処理
// TODO: 現在これを通さずに処理しているので、TaskQueueに通す
func handleTask(w http.ResponseWriter, r *http.Request) {
	c := newContext(r)
	data := r.FormValue("data")
	if data == "" {
		errorf(c, "No data")
		return
	}

	j, err := base64.StdEncoding.DecodeString(data)
	logf(c, (string)(j))
	if err != nil {
		errorf(c, "base64 DecodeString: %v", err)
		return
	}

	mes := new(ReceivedMessage)
	err = json.Unmarshal(j, mes)
	if err != nil {
		errorf(c, "json.Unmarshal: %v", err)
		return
	}

	bot, err := newLINEBot(c)
	if err != nil {
		errorf(c, "newLINEBot: %v", err)
		return
	}

	logf(c, "EventType: %s\nMessage: %#v", mes.Type, mes.Message)

	replyMessage(c, bot, mes)

	w.WriteHeader(200)
}

func replyMessage(c context.Context, bot *linebot.Client, mes *ReceivedMessage) {
	// トークルームID（個別チャット：相手ユーザーID, グループチャット：グループID）
	var talkId string
	if mes.Source.Type == linebot.EventSourceTypeGroup {
		talkId = mes.Source.GroupID
	} else {
		talkId = mes.Source.UserID
	}
	logf(c, "TalkId: %s", talkId)

	dHandler := cloudstrage.NewDropboxClient()

	// 万が一処理する種類以外のメッセージタスクが飛んできたら、200返して終わりにする
	if mes.Type == linebot.EventTypeJoin {
		dHandler := cloudstrage.NewDropboxClient()
		dHandler.NewFolder(talkId)
	} else if mes.Type != linebot.EventTypeMessage {
		return
	}

	var fileName string
	if mes.Message.FileName != nil {
		fileName = *mes.Message.FileName
	} else {
		// 画像の場合、ファイル名がないので、日時で名前を生成
		jst := time.FixedZone("Asia/Tokyo", 9*60*60)
		fileName = mes.Timestamp.In(jst).Format("20060102-150305")
	}
	logf(c, "FileName: %s", fileName)

	// ファイル本体を並列で取得
	fileBody, err := handleContent(bot, mes.Message.ID)
	defer fileBody.Content.Close()
	if err != nil {
		errorf(c, "HandleContent: %v", err)
	}

	log.Printf("Upload File")
	switch fileBody.ContentType {
	case "image/jpeg":
		fileName += ".jpg"
	case "image/gif":
		fileName += ".gif"
	case "image/png":
		fileName += ".png"
	}
	res, err := dHandler.UploadFile(talkId, fileName, fileBody.Content)
	logf(c, "Uploaded: %v", res)
	if err != nil {
		errorf(c, "Upload: %v", err)
	}

	replyMessage := linebot.NewTextMessage("ok")
	if _, err := bot.ReplyMessage(mes.ReplyToken, replyMessage).WithContext(c).Do(); err != nil {
		errorf(c, "ReplayMessage: %v", err)
		return
	}
}

func handleContent(bot *linebot.Client, messageID string) (content *linebot.MessageContentResponse, err error) {
	content, err = bot.GetMessageContent(messageID).Do()
	if err != nil {
		return
	}

	log.Printf("Got file: %s", content.ContentType)
	return
}