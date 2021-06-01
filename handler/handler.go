package handler

import (
	"encoding/json"
	dblayer "filestore-server/db"
	"filestore-server/meta"
	"filestore-server/util"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

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
		//meta.UpdataFileMeta(fileMeta)
		meta.UpdataFileMetaDB(fileMeta)

		r.ParseForm()
		username := r.Form.Get("username")
		suc := dblayer.OnUserFileUploadFinished(username, fileMeta.FileSha1, fileMeta.FileName, fileMeta.FileSize)

		if suc {
			http.Redirect(w, r, "/static/view/home.html", http.StatusFound)
		} else {
			w.Write([]byte("Upload Failed"))
		}
		http.Redirect(w, r, "/file/upload/suc", http.StatusFound)
	}
}

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
	suc := dblayer.OnUserFileUploadFinished(username, filehash, filename, int64(filesize))
	respMsg := util.RespMsg{}
	if suc {
		respMsg = util.RespMsg{
			Code: 0,
			Msg:  "秒传成功",
		}
	} else {
		respMsg = util.RespMsg{
			Code: -2,
			Msg:  "秒传失败，请稍后再试",
		}
	}
	w.Write(respMsg.JSONBytes())
	return
}

func UploadSucHandler(w http.ResponseWriter, r *http.Request) {
	//todo:
	io.WriteString(w, "upload finshed")
}

//获取文件元信息
func QueryFilehandler(w http.ResponseWriter, r *http.Request) {
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

//下载文件
func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	filehash := r.Form.Get("filehash")
	downFileMeta := meta.GetFileMeta(filehash)

	f, err := os.Open(downFileMeta.Location)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer f.Close()

	//需要优化，如果文件比较大，应该使用流的方式
	data, err := ioutil.ReadAll(f)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/octect-stream")
	w.Header().Set("Cotext-Descrption", "attachment;filename=\""+downFileMeta.FileName+"\"")
	w.Write(data)
}

//更新元文件信息接口
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
func FileDeletaHandle(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fileSha1 := r.Form.Get("fileHash")
	fMeta := meta.GetFileMeta(fileSha1)
	//todo:这里可能删除失败
	os.Remove(fMeta.Location)
	//删除索引数据
	meta.RemoveFileMeta(fileSha1)
}
