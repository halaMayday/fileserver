package client

import (
	"context"
	"encoding/json"
	"filestore-server/service/dbproxy/orm"
	dbProto "filestore-server/service/dbproxy/proto"
	"github.com/micro/go-micro"
	"github.com/mitchellh/mapstructure"
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

func TableFileToFileMeta(tfile orm.TableFile) FileMeta {
	return FileMeta{
		FileSha1: tfile.FileHash,
		FileName: tfile.FileName.String,
		FileSize: tfile.FileSize.Int64,
		Location: tfile.FileAddr.String,
	}
}

func ToTableUser(src interface{}) orm.TableUser {
	user := orm.TableUser{}
	mapstructure.Decode(src, &user)
	return user
}

func ToTableFile(src interface{}) orm.TableFile {
	file := orm.TableFile{}
	mapstructure.Decode(src, &file)
	return file
}

func ToTableFiles(src interface{}) []orm.TableFile {
	files := []orm.TableFile{}
	mapstructure.Decode(src, &files)
	return files
}

func ToTableUserFile(src interface{}) orm.TableUserFile {
	ufile := orm.TableUserFile{}
	mapstructure.Decode(src, &ufile)
	return ufile
}

func ToTableUserFiles(src interface{}) []orm.TableUserFile {
	ufile := []orm.TableUserFile{}
	mapstructure.Decode(src, &ufile)
	return ufile
}

//UserSignUp:用户注册
func UserSignUp(username, encPassword string) (*orm.ExecResult, error) {
	userInfo, _ := json.Marshal([]interface{}{username, encPassword})
	resp, err := execAction("/user/Signup", userInfo)
	return parseBody(resp), err
}

//UserSignin:用户登录
func UserSignin(username, encPassword string) (*orm.ExecResult, error) {
	userInfo, _ := json.Marshal([]interface{}{username, encPassword})
	resp, err := execAction("/user/Signin", userInfo)
	return parseBody(resp), err
}

//UpdateToken：更新用户token
func UpdateToken(username, token string) (*orm.ExecResult, error) {
	userInfo, _ := json.Marshal([]interface{}{username, token})
	resp, err := execAction("/user/UpdateToken", userInfo)
	return parseBody(resp), err
}

//UpdateUserLastOnLineTime：更新最后在线时间
func UpdateUserLastOnLineTime(username string) (*orm.ExecResult, error) {
	userInfo, _ := json.Marshal([]interface{}{username, username})
	resp, err := execAction("/user/UpdateOnLineTime", userInfo)
	return parseBody(resp), err
}

//GetUserInfo：获取用户信息
func GetUserInfo(username string) (*orm.ExecResult, error) {
	userInfo, _ := json.Marshal([]interface{}{username, username})
	resp, err := execAction("/user/GetUserInfo", userInfo)
	return parseBody(resp), err
}

//QueryUserFileMetas：查询用户
func QueryUserFileMeta(username, filehash string) (*orm.ExecResult, error) {
	userFileInfo, _ := json.Marshal([]interface{}{username, filehash})
	resp, err := execAction("/ufile/QueryUserFileMeta", userFileInfo)
	return parseBody(resp), err
}

func QueryUserFileMetas(username string, limit int) (*orm.ExecResult, error) {
	uInfo, _ := json.Marshal([]interface{}{username, limit})
	res, err := execAction("/ufile/QueryUserFileMetas", uInfo)
	return parseBody(res), err
}

func RenameFileName(username, filehash, filename string) (*orm.ExecResult, error) {
	uInfo, _ := json.Marshal([]interface{}{username, filehash, filename})
	res, err := execAction("/ufile/RenameFileName", uInfo)
	return parseBody(res), err
}

// OnFileUploadFinished : 新增/更新文件元信息到mysql中
func OnFileUploadFinished(fmeta FileMeta) (*orm.ExecResult, error) {
	uInfo, _ := json.Marshal([]interface{}{fmeta.FileSha1, fmeta.FileName, fmeta.FileSize, fmeta.Location})
	res, err := execAction("/file/OnFileUploadFinished", uInfo)
	return parseBody(res), err
}

// OnUserFileUploadFinished : 新增/更新文件元信息到mysql中
func OnUserFileUploadFinished(username string, fmeta FileMeta) (*orm.ExecResult, error) {
	uInfo, _ := json.Marshal([]interface{}{username, fmeta.FileSha1,
		fmeta.FileName, fmeta.FileSize})
	res, err := execAction("/ufile/OnUserFileUploadFinished", uInfo)
	return parseBody(res), err
}

//GetFileMeta 获取文件信息
func GetFileMeta(filehash string) (*orm.ExecResult, error) {
	uInfo, _ := json.Marshal([]interface{}{filehash})
	res, err := execAction("/file/GetFileMeta", uInfo)
	return parseBody(res), err
}

//UpdateFileLocation:
func UpdateFileLocation(filehash, fileaddr string) (*orm.ExecResult, error) {
	uInfo, _ := json.Marshal([]interface{}{filehash, fileaddr})
	res, err := execAction("/file/UpdateFileLocation", uInfo)
	return parseBody(res), err
}
