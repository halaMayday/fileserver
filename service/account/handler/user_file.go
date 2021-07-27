package handler

import (
	"context"
	"encoding/json"
	"filestore-server/common"

	proto "filestore-server/service/account/proto"
	dbcli "filestore-server/service/dbproxy/client"
)

func (u *User) UserFiles(ctx context.Context, req *proto.ReqUserFile, resp *proto.RespUserFile) error {
	username := req.Username
	limit := int(req.Limit)

	userFileDbResp, err := dbcli.QueryUserFileMetas(username, limit)

	if err != nil || !userFileDbResp.Suc {
		resp.Code = common.StatusServerError
		return err
	}

	userFiles := dbcli.ToTableUserFiles(userFileDbResp)
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
	dbResp, err := dbcli.RenameFileName(req.Username, req.Filehash, req.NewFileName)

	if err != nil || !dbResp.Suc {
		resp.Code = common.StatusServerError
		return err
	}
	userFiles := dbcli.ToTableUserFiles(dbResp.Data)
	data, err := json.Marshal(userFiles)

	if err != nil {
		resp.Code = common.StatusServerError
		return nil
	}

	resp.FileData = data
	return nil
}
