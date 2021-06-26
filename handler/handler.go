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
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		//返回上传的的
		//html页面
		data, err := ioutil.ReadFile("./static/view/upload.html")
		if err != nil {
			io.WriteString(w, "intelnel server error")
			return
		}
		io.WriteString(w, string(data))
	} else if r.Method == "POST" {
		//接收文件流以及储存到本地目录
		file, head, err := r.FormFile("file")
		if err != nil {
			fmt.Printf("Failed to get data ,err:#{err.Error()}\n")
			return
		}
		defer file.Close()

		fileMeta := meta.FileMeta{
			FileName: head.Filename,
			Location: "/tmp/" + head.Filename,
			UploadAt: time.Now().Format("2006-01-02 15:04:05"),
		}

		newFile, err := os.Create(fileMeta.Location)
		if err != nil {
			fmt.Printf("Faile to create file ,err:#{err.Error()}\n")
		}
		defer newFile.Close()

		fileMeta.FileSize, err = io.Copy(newFile, file)
		if err != nil {
			fmt.Printf("Faile to save data into file ,err:#{err.Error()}\n")
		}

		newFile.Seek(0, 0)
		fileMeta.FileSha1 = util.FileSha1(newFile)

		// 游标重新回到文件头部
		newFile.Seek(0, 0)

		if cfg.CurrentStoreType == cmn.StoreCeph {
			//文件寫入ceph
			data, err := ioutil.ReadAll(newFile)
			if err != nil {
				log.Println("Upload file failed by ceph ,err:{}", err.Error())
				w.Write([]byte("Upload failed!"))
				return
			}
			cephPath := "/ceph" + fileMeta.FileSha1
			_ = ceph.PutObject("userfile", cephPath, data)
			fileMeta.Location = cephPath
		} else if cfg.CurrentStoreType == cmn.StoreOSS {
			//文件吸入OSS儲存
			ossPath := "oss/" + fileMeta.FileSha1
			//判断写入OSS为同步还是异步
			if !cfg.AsyncTransferEnable {
				err := oss.Bucket().PutObject(ossPath, newFile)
				if err != nil {
					log.Println("upload file by oss error,{}", err.Error())
					w.Write([]byte("Upload failed!"))
					return
				}
			} else {
				//写入转移的异步队列
				data := mq.TransferData{
					FileHash:      fileMeta.FileSha1,
					CurLocation:   fileMeta.Location,
					DestLocation:  ossPath,
					DestStoreType: cmn.StoreOSS}

				pubData, _ := json.Marshal(data)
				pubSuc := mq.Publish(
					cfg.TransExchangeName,
					cfg.TransOSSErrQueueName,
					pubData)

				if !pubSuc {
					//todo 可以采取一些死信队列的处理方法等等
					log.Println("当前发送转移信息失败，稍后重试")
				}

			}
		}

		meta.UpdataFileMetaDB(fileMeta)

		//更新用户文件表记录
		r.ParseForm()
		username := r.Form.Get("username")
		suc := dblayer.OnUserFileUploadFinished(username, fileMeta.FileSha1, fileMeta.FileName, fileMeta.FileSize)
		if suc {
			http.Redirect(w, r, "/static/view/home.html", http.StatusFound)
		} else {
			w.Write([]byte("Upload Failed"))
		}
	}
}

//UploadSucHandler:上传已完成
func UploadSucHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Upload finished!")
}

//获取文件元信息
func QueryFileHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.Form.Get("username")
	limitCnt, _ := strconv.Atoi(r.Form.Get("limit"))
	//todo:
	userFiles, err := dblayer.QueryUserFileMetas(username, limitCnt)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	data, err := json.Marshal(userFiles)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

//
func TryFastUploadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	//1.解析参数
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filename := r.Form.Get("filename")
	filesize := r.Form.Get("filesize")
	//2.从文件表中查询相同的hash的文件记录
	fileMeta, err := dblayer.GetFileMeta(filehash)
	if err != nil {
		fmt.Println("action:{},hava a error:{]", "TryFastUploadHandler", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	//3.查不到记录则返回秒传失败
	if fileMeta != nil {
		//todo:秒传失败
		resp := util.RespMsg{
			Code: -1,
			Msg:  "秒传失败，请访问普通接口",
		}
		w.Write(resp.JSONBytes())
		return
	}

	//4.上传过则将文件信息写入用户文件表，返回成功
	fileSize, err := strconv.Atoi(filesize)
	if err != nil {
		w.Write(util.NewRespMsg(-1, "params invalid", nil).JSONBytes())
		return
	}
	suc := dblayer.OnUserFileUploadFinished(username, filehash, filename, int64(fileSize))
	respMsg := util.RespMsg{}
	if suc {
		respMsg = util.RespMsg{
			Code: 0,
			Msg:  "秒传成功",
		}
		w.Write(respMsg.JSONBytes())
		return
	} else {
		respMsg = util.RespMsg{
			Code: -2,
			Msg:  "秒传失败，请稍后再试",
		}
		w.Write(respMsg.JSONBytes())
		return
	}
}

//下载文件
func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fileSha1 := r.Form.Get("filehash")
	fileMeta, _ := meta.GetFileMetaDB(fileSha1)

	wait4DownloadFile, err := os.Open(fileMeta.Location)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer wait4DownloadFile.Close()

	data, err := ioutil.ReadAll(wait4DownloadFile)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octect-stream")
	// attachment表示文件将会提示下载到本地，而不是直接在浏览器中打开
	w.Header().Set("content-disposition", "attachment; filename=\""+fileMeta.FileName+"\"")
	w.Write(data)
}

//FileMetaUpdataHandle: 更新元文件信息接口(重命名)
func FileMetaUpdataHandle(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	opType := r.Form.Get("op")
	fileSha1 := r.Form.Get("fileHash")
	newFileName := r.Form.Get("filename")

	if opType != "0" {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if r.Method != "post" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	curFileMeta := meta.GetFileMeta(fileSha1)
	curFileMeta.FileName = newFileName
	meta.UpdataFileMetaDB(curFileMeta)

	data, err := json.Marshal(curFileMeta)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

const passwd_salt = "halaMayday"

//文件删除的接口
func FileDeleteHandle(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fileSha1 := r.Form.Get("fileHash")
	fMeta := meta.GetFileMeta(fileSha1)
	//todo:这里可能删除失败
	os.Remove(fMeta.Location)
	//删除索引数据
	meta.RemoveFileMeta(fileSha1)
	w.WriteHeader(http.StatusOK)
}

func DownloadURLHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fileHash := r.Form.Get("filehash")

	//从文件表中查询记录
	row, _ := dblayer.GetFileMeta(fileHash)

	//判断文件存放的目录：本地，OSS,还是CEPH
	if strings.HasPrefix(row.FileAddr.String, "/tmp") {
		username := r.Form.Get("username")
		token := r.Form.Get("token")

		downloadUrl := fmt.Sprintf("http://%s/file/download?filehash=%s&username=%s&token=%s",
			r.Host, fileHash, username, token)
		w.Write([]byte(downloadUrl))
	} else if strings.HasPrefix(row.FileAddr.String, "/ceph") {
		//TODO:ceph的下载url
	} else if strings.HasPrefix(row.FileAddr.String, "oss/") {
		//oss下载url
		downloadUrl := oss.DownloadURL(row.FileAddr.String)
		w.Write([]byte(downloadUrl))
	}
}
