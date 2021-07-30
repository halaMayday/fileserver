package api

import (
	"bytes"
	"encoding/json"
	redisPool "filestore-server/cache/redis"
	cmn "filestore-server/common"
	cfg "filestore-server/config"
	dblayer "filestore-server/db"
	"filestore-server/meta"
	"filestore-server/mq"
	dbcli "filestore-server/service/dbproxy/client"
	"filestore-server/service/dbproxy/orm"
	"filestore-server/store/ceph"
	"filestore-server/store/oss"
	"filestore-server/util"
	"github.com/garyburd/redigo/redis"
	"github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

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

	// 1. 从from表单中获取文件内容句柄
	file, head, err := c.Request.FormFile("file")
	if err != nil {
		log.Printf("Failed to get form data, err:%s\n", err.Error())
		errCode = -1
		return
	}
	defer file.Close()

	// 2. 把文件内容转为[]byte
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		log.Printf("Failed to get file data, err:%s\n", err.Error())
		errCode = -2
		return
	}

	//3.构建文件元信息
	fileMeta := dbcli.FileMeta{
		FileName: head.Filename,
		FileSha1: util.Sha1(buf.Bytes()),
		FileSize: int64(len(buf.Bytes())),
		UploadAt: time.Now().Format("2021-0728-02 15:04:05"),
	}

	//4.将文件写入临时储存的位置
	fileMeta.Location = cfg.TempLocalRootDir + fileMeta.FileSha1
	newFile, err := os.Create(fileMeta.Location)
	if err != nil {
		log.Printf("Failed to create file, err:%s\n", err.Error())
		errCode = -3
		return
	}
	defer newFile.Close()

	nByte, err := newFile.Write(buf.Bytes())
	if int64(nByte) != fileMeta.FileSize || err != nil {
		log.Printf("Failed to save data into file, writtenSize:%d, err:%s\n", nByte, err.Error())
		errCode = -4
		return
	}

	//5. 同步或者异步将文件转移到Ceph/OSS
	// 游标重新回到文件头部
	newFile.Seek(0, 0)

	uploadPath := ""
	if cfg.CurrentStoreType == cmn.StoreCeph {
		uploadPath = cfg.CephRootDir + fileMeta.FileSha1
		uploadbyCephSuc := doUploadbyCeph(newFile, uploadPath, fileMeta)
		if !uploadbyCephSuc {
			errCode = -4
			return
		}

	} else if cfg.CurrentStoreType == cmn.StoreOSS {
		//文件吸入OSS儲存
		uploadPath = cfg.OSSRootDir + fileMeta.FileSha1

		uploadByOssSuc := doUploadByOSS(newFile, uploadPath, fileMeta)
		if !uploadByOssSuc {
			errCode = -5
			return
		}
	}

	//6.更新文件表记录
	_, err = dbcli.OnFileUploadFinished(fileMeta)
	if err != nil {
		errCode = -6
		return
	}

	//7.更新用户文件表记录
	username := c.Request.FormValue("username")
	upResp, err := dbcli.OnUserFileUploadFinished(username, fileMeta)
	if err == nil && upResp.Suc {
		errCode = 0
	} else {
		errCode = -6
	}
}

//通过ceph上传文件
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

//通过OSS上传文件
func doUploadByOSS(newFile *os.File, path string, fileMeta meta.FileMeta) bool {
	//如果同步的话就直接上传到ceph
	if !cfg.AsyncTransferEnable {
		//TODO:设置oss中的文件名，方便指定文件名下载
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

//异步转移任务
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

//TryFastUploadHandler:尝试快传接口
func TryFastUploadHandler(c *gin.Context) {

	//1.解析参数
	username := c.Request.FormValue("username")
	filehash := c.Request.FormValue("filehash")
	filename := c.Request.FormValue("filename")
	//2.从文件表中查询相同的hash的文件记录
	fileMetaResp, err := dbcli.GetFileMeta(filehash)
	if err != nil {
		log.Println("action:{},hava a error:{]", "TryFastUploadHandler", err)
		c.Status(http.StatusInternalServerError)
		return
	}
	//3.查不到记录则返回秒传失败
	if !fileMetaResp.Suc {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "秒传失败，请访问普通接口",
		}
		c.Data(http.StatusOK, "applicaiton/json", resp.JSONBytes())
		return
	}
	//4.上传过则将文件信息写入用户文件表，返回成功
	fileMeta := dbcli.TableFileToFileMeta(fileMetaResp.Data.(orm.TableFile))
	fileMeta.FileName = filename
	upResp, err := dbcli.OnUserFileUploadFinished(username, fileMeta)

	if err == nil && upResp.Suc {
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
