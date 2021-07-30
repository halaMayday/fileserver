package api

import (
	"filestore-server/common"
	"filestore-server/config"
	dbcli "filestore-server/service/dbproxy/client"
	"filestore-server/store/ceph"
	"filestore-server/store/oss"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

//DownloadURLHandler:生成文件的地址
func DownloadURLHandler(c *gin.Context) {

	fileHash := c.Request.FormValue("filehash")

	//从文件表中查询记录
	dbResp, err := dbcli.GetFileMeta(fileHash)
	if err != nil {
		c.JSON(
			http.StatusOK,
			gin.H{
				"code": common.StatusServerError,
				"msg":  "server error",
			})
		return
	}
	tblFile := dbcli.ToTableFile(dbResp.Data)

	//判断文件存放的目录：本地，OSS,还是CEPH
	if strings.HasPrefix(tblFile.FileAddr.String, config.TempLocalRootDir) {
		username := c.Request.FormValue("username")
		token := c.Request.FormValue("token")

		downloadUrl := fmt.Sprintf("http://%s/file/download?filehash=%s&username=%s&token=%s",
			c.Request.Host, fileHash, username, token)

		c.Data(http.StatusOK, "application/octet-stream", []byte(downloadUrl))
	} else if strings.HasPrefix(tblFile.FileAddr.String, config.CephRootDir) {
		//TODO:ceph的下载url
	} else if strings.HasPrefix(tblFile.FileAddr.String, config.OSSRootDir) {
		//oss下载url
		downloadUrl := oss.DownloadURL(tblFile.FileAddr.String)
		c.Data(http.StatusOK, "application/octet-stream", []byte(downloadUrl))
	}
}

//下载文件
func DownloadHandler(c *gin.Context) {
	fileSha1 := c.Request.FormValue("filehash")
	username := c.Request.FormValue("username")

	//todo:处理异常情况
	fileMetaResp, ferr := dbcli.GetFileMeta(fileSha1)
	userFileMetaResp, uferr := dbcli.QueryUserFileMeta(username, fileSha1)
	if ferr != nil || uferr != nil || !fileMetaResp.Suc || !userFileMetaResp.Suc {
		c.JSON(
			http.StatusOK,
			gin.H{
				"code": common.StatusServerError,
				"msg":  "server error",
			})
		return
	}

	tableFile := dbcli.ToTableFile(fileMetaResp.Data)
	tableUserFile := dbcli.ToTableUserFile(userFileMetaResp.Data)

	if strings.HasPrefix(tableFile.FileAddr.String, config.TempLocalRootDir) {
		//本地文件直接下载
		c.FileAttachment(tableFile.FileAddr.String, tableUserFile.FileName)
	} else if strings.HasPrefix(tableFile.FileAddr.String, config.CephRootDir) {
		//ceph中的文件，通过cephapi下载
		//TODO:ceph桶的名字
		bucket := ceph.GetCephBucket("userfile")
		data, _ := bucket.Get(tableFile.FileAddr.String)
		c.Header("content-disposition", "attachment; filename=\""+tableUserFile.FileName+"\"")
		c.Data(http.StatusOK, "application/octect-stream", data)
	} else {
		c.JSON(
			http.StatusOK,
			gin.H{
				"code": common.StatusServerError,
				"msg":  "文件路径=" + tableFile.FileAddr.String + "无法下载",
			})
		return
	}
}
