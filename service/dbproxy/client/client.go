package client

import (
	"context"
	"encoding/json"
	"filestore-server/service/dbproxy/orm"
	dbProto "filestore-server/service/dbproxy/proto"
	"github.com/micro/go-micro"
)

//文件信息结构
type FileMeta struct {
	FileSha1 string
	FileName string
	FileSize int64
	Location string
	UploadAt string
}

var (
	dbCli dbProto.DBProxyService
)

func Init(service micro.Service) {
	//初始化一个dbproxy的服务端
	dbCli = dbProto.NewDBProxyService("go.micro.service.dbproxy", service.Client())
}

//execAction:向dbproxy请求执行action
func execAction(funcName string, paramJson []byte) (*dbProto.RespExec, error) {
	return dbCli.ExecuteAction(context.TODO(), &dbProto.ReqExec{
		Action: []*dbProto.SingleAction{
			&dbProto.SingleAction{
				Name:   funcName,
				Params: paramJson,
			},
		},
	})
}

//parseBody:转换rpc返回的结果
func parseBody(resp *dbProto.RespExec) *orm.ExecResult {
	if resp == nil || resp.Data == nil {
		return nil
	}
	resList := []orm.ExecResult{}
	_ = json.Unmarshal(resp.Data, &resList)
	//TODO:
	if len(resList) > 0 {
		return &resList[0]
	}
	return nil
}
