package handler

import (
	"context"
	"encoding/json"
	"filestore-server/common"
	dblayer "filestore-server/db"
	"filestore-server/meta"
	proto "filestore-server/service/account/proto"
	"log"
)

func (u *User) UserFiles(ctx context.Context, req *proto.ReqUserFile, resp *proto.RespUserFile) error {
	username := req.Username
	limit := int(req.Limit)

	userFiles, err := dblayer.QueryUserFileMetas(username, limit)

	if err != nil {
		resp.Code = common.StatusServerError
		return err
	}

	data, err := json.Marshal(userFiles)
	if err != nil {
		resp.Code = common.StatusServerError
		return nil
	}
	resp.FileData = data
	return nil
}

//文件重命名
func (u *User) UserFileRename(ctx context.Context, req *proto.ReqUserFileRename, resp *proto.RespUserFileRename) error {
	//username := req.Username
	fileHash := req.Filehash
	newFileName := req.NewFileName

	curFileMeta, err := meta.GetFileMetaDB(fileHash)
	if err != nil {
		log.Println(err.Error())
		resp.Code = common.StatusServerError
		return err
	}
	curFileMeta.FileName = newFileName
	meta.UpdataFileMetaDB(*curFileMeta)
	data, err := json.Marshal(*curFileMeta)
	if err != nil {
		log.Println(err.Error())
		resp.Code = common.StatusServerError
		return err
	}
	resp.FileData = data
	return nil
}
