package rpc

import (
	"context"
	"filestore-server/service/download/config"
	dlProto "filestore-server/service/download/proto"
)

//DownLoad: download 结构体
type Download struct{}

//DownloadEntry:获取下载入口
func (u *Download) DownloadEntry(
	ctx context.Context,
	req *dlProto.ReqEntry,
	res *dlProto.RespEntry) error {
	res.Entry = config.DownloadEntry
	return nil
}
