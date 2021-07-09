package handler

import (
	"encoding/json"
	cmn "filestore-server/common"
	cfg "filestore-server/config"
	dblayer "filestore-server/db"
	"filestore-server/meta"
	"filestore-server/mq"
	"filestore-server/store/ceph"
	"filestore-server/store/oss"
	"filestore-server/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

//处理文件上传
func UploadHandler(c *gin.Context) {
	c.Redirect(http.StatusFound, "/static/view/upload.html")
}

//DoUploadHandler:处理文件上传POST请求
func DoUploadHandler(c *gin.Context) {
	errCode := 0
	defer func() {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		if errCode < 0 {
			c.JSON(http.StatusOK, gin.H{
				"code": errCode,
				"msg":  "上传失败",
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"code": errCode,
				"msg":  "上传成功",
			})
		}
	}()

	//接收文件流以及储存到本地目录
	file, head, err := c.Request.FormFile("file")
	if err != nil {
		log.Printf("Failed to get form data, err:%s\n", err.Error())
		errCode = -1
		return
	}
	defer file.Close()

	fileMeta := meta.FileMeta{
		FileName: head.Filename,
		Location: cmn.TmpFile + head.Filename,
		UploadAt: time.Now().Format("2006-01-02 15:04:05"),
	}

	newFile, err := os.Create(fileMeta.Location)
	if err != nil {
		log.Printf("Failed to get file data, err:%s\n", err.Error())
		errCode = -2
		return
	}
	defer newFile.Close()

	fileMeta.FileSize, err = io.Copy(newFile, file)
	if err != nil {
		log.Printf("Failed to save data into file,  err:%s\n", err.Error())
		errCode = -3
		return
	}

	newFile.Seek(0, 0)
	fileMeta.FileSha1 = util.FileSha1(newFile)

	// 游标重新回到文件头部
	newFile.Seek(0, 0)

	uploadPath := ""
	if cfg.CurrentStoreType == cmn.StoreCeph {
		uploadPath = "/ceph/" + fileMeta.FileSha1
		uploadbyCephSuc := doUploadbyCeph(newFile, uploadPath, fileMeta)
		if !uploadbyCephSuc {
			errCode = -4
			return
		}

	} else if cfg.CurrentStoreType == cmn.StoreOSS {
		//文件吸入OSS儲存
		uploadPath = "oss/" + fileMeta.FileSha1

		uploadByOssSuc := doUploadByOSS(newFile, uploadPath, fileMeta)
		if !uploadByOssSuc {
			errCode = -5
			return
		}
	}

	fileMeta.Location = uploadPath
	meta.UpdataFileMetaDB(fileMeta)

	//更新用户文件表记录
	username := c.Request.FormValue("username")
	suc := dblayer.OnUserFileUploadFinished(username, fileMeta.FileSha1, fileMeta.FileName, fileMeta.FileSize)
	if !suc {
		log.Println("更新用户文件表记录失败")
		errCode = -7
		return
	}
}

func doUploadbyCeph(newFile *os.File, path string, fileMeta meta.FileMeta) bool {
	//如果同步的话就直接上传到ceph
	if !cfg.AsyncTransferEnable {
		//文件寫入ceph
		data, err := ioutil.ReadAll(newFile)
		if err != nil {
			log.Println(err.Error())
			return false
		}
		err = ceph.PutObject(cfg.OSSBucket, path, data)
		if err != nil {
			log.Println("upload file by CEPH error,{}", err.Error())
			return false
		}
	} else {
		//如果异步的话，写入转移的异步队列
		transferDataSuc := doTransferData(fileMeta, path, cmn.StoreCeph)
		if !transferDataSuc {
			return false
		}
	}
	return true
}

func doUploadByOSS(newFile *os.File, path string, fileMeta meta.FileMeta) bool {
	//如果同步的话就直接上传到ceph
	if !cfg.AsyncTransferEnable {
		err := oss.Bucket().PutObject(path, newFile)
		if err != nil {
			log.Println("upload file by oss error,{}", err.Error())
			return false

		}
	} else {
		//如果异步的话，写入转移的异步队列
		transferDataSuc := doTransferData(fileMeta, path, cmn.StoreOSS)
		if !transferDataSuc {
			return false
		}
	}
	return true
}

func doTransferData(fileMeta meta.FileMeta, path string, storeType cmn.StoreType) bool {
	//写入转移的异步队列
	data := mq.TransferData{
		FileHash:      fileMeta.FileSha1,
		CurLocation:   fileMeta.Location,
		DestLocation:  path,
		DestStoreType: storeType}

	pubData, _ := json.Marshal(data)
	pubSuc := mq.Publish(
		cfg.TransExchangeName,
		cfg.TransOSSErrQueueName,
		pubData)

	if !pubSuc {
		//todo 可以采取一些死信队列的处理方法等等
		log.Println("当前发送转移信息失败，稍后重试")
		return false
	}
	return true
}

//QueryFileHandler:获取文件元信息
func QueryFileHandler(c *gin.Context) {
	username := c.Request.FormValue("username")
	limitCnt, _ := strconv.Atoi(c.Request.FormValue("limit"))
	userFiles, err := dblayer.QueryUserFileMetas(username, limitCnt)

	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{
				"msg":  "query file error",
				"code": -6,
			})
		return
	}
	data, err := json.Marshal(userFiles)
	if err != nil {
		fmt.Printf("json.Marshal user:%+v\n error:%s", userFiles, err.Error())
		c.JSON(http.StatusInternalServerError,
			gin.H{
				"msg":  "json.Marshal error",
				"code": -7,
			})
		return
	}
	c.Data(http.StatusOK, "application/json", data)
}

//
func TryFastUploadHandler(c *gin.Context) {

	//1.解析参数
	username := c.Request.FormValue("username")
	filehash := c.Request.FormValue("filehash")
	filename := c.Request.FormValue("filename")
	filesize := c.Request.FormValue("filesize")
	//2.从文件表中查询相同的hash的文件记录
	fileMeta, err := dblayer.GetFileMeta(filehash)
	if err != nil {
		fmt.Println("action:{},hava a error:{]", "TryFastUploadHandler", err)
		c.Status(http.StatusInternalServerError)
		return
	}
	//3.查不到记录则返回秒传失败
	if fileMeta != nil {
		//todo:秒传失败
		resp := util.RespMsg{
			Code: -1,
			Msg:  "秒传失败，请访问普通接口",
		}
		c.Data(http.StatusOK, "applicaiton/json", resp.JSONBytes())
		return
	}

	//4.上传过则将文件信息写入用户文件表，返回成功
	fileSize, err := strconv.Atoi(filesize)
	if err != nil {
		resp := util.NewRespMsg(-1, "params invalid", nil)
		c.Data(http.StatusOK, "applicaiton/json", resp.JSONBytes())
		return
	}
	suc := dblayer.OnUserFileUploadFinished(username, filehash, filename, int64(fileSize))
	if suc {
		resp := util.RespMsg{
			Code: 0,
			Msg:  "秒传成功",
		}
		c.Data(http.StatusOK, "applicaiton/json", resp.JSONBytes())
		return
	} else {
		resp := util.RespMsg{
			Code: -2,
			Msg:  "秒传失败，请稍后再试",
		}
		c.Data(http.StatusOK, "applicaiton/json", resp.JSONBytes())
		return
	}
}

//FileMetaUpdataHandle: 更新元文件信息接口(重命名)
func FileMetaUpdataHandle(c *gin.Context) {
	opType := c.Request.FormValue("op")
	fileSha1 := c.Request.FormValue("filehash")
	newFileName := c.Request.FormValue("filename")

	if opType != "0" {
		c.Status(http.StatusForbidden)
		return
	}
	curFileMeta, err := meta.GetFileMetaDB(fileSha1)
	if err != nil {
		log.Println(err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}

	curFileMeta.FileName = newFileName
	meta.UpdataFileMetaDB(*curFileMeta)
	data, err := json.Marshal(*curFileMeta)
	if err != nil {
		log.Println(err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Data(http.StatusOK, "application/json", data)
}

//const passwd_salt = "halaMayday"

//文件删除的接口
//目前只删除了本地文件，云上的文件没有删除。删除云上的文件，应该交给mq.
func FileDeleteHandle(c *gin.Context) {
	fileSha1 := c.Request.FormValue("fileHash")
	fMeta, err := meta.GetFileMetaDB(fileSha1)
	//todo:这里可能删除失败
	os.Remove(fMeta.Location)
	//删除索引数据
	meta.RemoveFileMeta(fileSha1)
	body := gin.H{"msg": "delete successed", "code": 0}
	marshal, err := json.Marshal(body)
	if err != nil {
		log.Println(err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Data(http.StatusOK, "application/json", marshal)
}

//下载文件
func DownloadHandler(c *gin.Context) {
	fileSha1 := c.Request.FormValue("filehash")
	fileMeta, _ := meta.GetFileMetaDB(fileSha1)

	wait4DownloadFile, err := os.Open(fileMeta.Location)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	defer wait4DownloadFile.Close()

	data, err := ioutil.ReadAll(wait4DownloadFile)
	if err != nil {
		c.JSON(
			http.StatusOK,
			gin.H{
				"code": cmn.StatusServerError,
				"msg":  "server error",
			})
		return
	}

	c.Header("content-disposition", "attachment; filename=\""+fileMeta.FileName+"\"")
	c.Data(http.StatusOK, "application/octect-stream", data)
}

func DownloadURLHandler(c *gin.Context) {

	fileHash := c.Request.FormValue("filehash")

	//从文件表中查询记录
	row, _ := dblayer.GetFileMeta(fileHash)

	//判断文件存放的目录：本地，OSS,还是CEPH
	if strings.HasPrefix(row.FileAddr.String, "/tmp") {
		username := c.Request.FormValue("username")
		token := c.Request.FormValue("token")

		downloadUrl := fmt.Sprintf("http://%s/file/download?filehash=%s&username=%s&token=%s",
			c.Request.Host, fileHash, username, token)

		c.Data(http.StatusOK, "application/octet-stream", []byte(downloadUrl))
	} else if strings.HasPrefix(row.FileAddr.String, "/ceph") {
		//TODO:ceph的下载url
	} else if strings.HasPrefix(row.FileAddr.String, "oss/") {
		//oss下载url
		downloadUrl := oss.DownloadURL(row.FileAddr.String)
		c.Data(http.StatusOK, "application/octet-stream", []byte(downloadUrl))
	}
}
