package cloudstrage

import (
	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/files"
	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/sharing"
	"io"
	"os"
)

type DropboxHandler struct {
	FilesClient files.Client
	SharingClient sharing.Client
}

// クライアントハンドラの生成
func NewDropboxClient() *DropboxHandler {
	handler := new(DropboxHandler)
	config := dropbox.Config{
		Token: os.Getenv("DROPBOX_TOKEN"),
		LogLevel: dropbox.LogInfo, // if needed, set the desired logging level. Default is off
	}

	handler.FilesClient = files.New(config)
	handler.SharingClient = sharing.New(config)

	return handler
}

// ファイルのアップロード
func (h *DropboxHandler) UploadFile(talkId string, fileName string, content io.Reader) (res *files.FileMetadata, err error) {
	req := files.NewCommitInfo("/" + talkId + "/" + fileName)
	res, err = h.FilesClient.Upload(req, content)
	return res, err
}

// ファイル情報の取得
func (h *DropboxHandler) GetFileMetaData(path string) (metaData *sharing.SharedFileMetadata, err error) {
	arg := sharing.NewGetFileMetadataArg(path)
	metaData, err = h.SharingClient.GetFileMetadata(arg)
	return metaData, err
}

// 新規フォルダー作成 & 共有設定
// TODO: 共有設定が動作していない
func (h *DropboxHandler) NewFolder(talkId string) (err error) {
	folderArg := files.NewCreateFolderArg("/" + talkId)
	_, err = h.FilesClient.CreateFolderV2(folderArg)
	if err != nil {
		return
	}

	sharingArg := sharing.NewShareFolderArg("/" + talkId)
	sharingArg.SharedLinkPolicy = &sharing.SharedLinkPolicy{dropbox.Tagged{sharing.SharedLinkPolicyAnyone}}
	_, err = h.SharingClient.ShareFolder(sharingArg)
	return
}