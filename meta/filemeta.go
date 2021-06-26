package meta

import (
	_ "expvar"
	mydb "filestore-server/db"
	"sort"
)

//文件信息结构
type FileMeta struct {
	FileSha1 string
	FileName string
	FileSize int64
	Location string
	UploadAt string
}

var fileMetas map[string]FileMeta

func init() {
	fileMetas = make(map[string]FileMeta)
}

//新增/更新元信息 UploadFileMeta
func UpdataFileMeta(fmeta FileMeta) {
	fileMetas[fmeta.FileSha1] = fmeta
}

//新增/更新文件元信息到mysql数据库
func UpdataFileMetaDB(fmeta FileMeta) bool {
	return mydb.OnFileUploadFinshed(
		fmeta.FileSha1, fmeta.FileName, fmeta.FileSize, fmeta.Location)
}

//通过sha1 获取文件的元信息对象
func GetFileMeta(fileSha1 string) FileMeta {
	return fileMetas[fileSha1]
}

//GetFileMetaDB:从mysql中获取文件元信息
func GetFileMetaDB(filesha1 string) (*FileMeta, error) {
	tFile, err := mydb.GetFileMeta(filesha1)
	if err != nil {
		return nil, err
	}
	fmeta := FileMeta{
		FileSha1: tFile.FileHash,
		FileName: tFile.FileName.String,
		FileSize: tFile.FileSize.Int64,
		Location: tFile.FileAddr.String,
	}
	return &fmeta, nil
}

//获取批量的文件元信息接口 : 未全部完成
func GetLastFileMetas(count int) ([]FileMeta, error) {
	fMetaArray := make([]FileMeta, len(fileMetas))
	for _, v := range fileMetas {
		fMetaArray = append(fMetaArray, v)
	}
	sort.Sort(ByUploadTime(fMetaArray))
	return fMetaArray[0:count], nil
}

//删除：todo 需要考虑线程同步
func RemoveFileMeta(fileSha1 string) {
	delete(fileMetas, fileSha1)
}
