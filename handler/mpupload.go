package handler

import (
	redisPool "filestore-server/cache/redis"
	dblayer "filestore-server/db"
	"filestore-server/util"
	"fmt"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
)

//MultipartUploadInfo:初始化信息
type MultipartUploadInfo struct {
	FileHash   string
	FileSize   int
	UploadID   string
	ChunkSize  int
	ChunkCount int
}

//初始化分块上传
func InitalMultipartUploadHandler(w http.ResponseWriter, r *http.Request) {
	//1.解析用户请求参数
	r.ParseForm()
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filesize, err := strconv.Atoi(r.Form.Get("filesize"))

	if err != nil {
		w.Write(util.NewRespMsg(-1, "params invalid", nil).JSONBytes())
	}

	//2.获得一个redis链接
	redisConn := redisPool.RedisPool().Get()
	defer redisConn.Close()
	//3.生成分块上传的初始化信息

	upInfo := MultipartUploadInfo{
		FileHash:   filehash,
		FileSize:   filesize,
		UploadID:   username + fmt.Sprintf("%x", time.Now().UnixNano()),
		ChunkSize:  5 * 1024 * 1024,
		ChunkCount: int(math.Ceil(float64(filesize) / (5 * 1024 * 1024))),
	}

	//4.将初始化信息写入到redis缓存
	redisConn.Do("HSET", "MP_"+upInfo.UploadID, "chunkcount", upInfo.ChunkCount)
	redisConn.Do("HSET", "MP_"+upInfo.UploadID, "filehash", upInfo.FileHash)
	redisConn.Do("HSET", "MP_"+upInfo.UploadID, "filesize", upInfo.FileSize)

	//5.将响应初始化数据返回到客户端
	w.Write(util.NewRespMsg(0, "OK", upInfo).JSONBytes())
}

//上传文件分块
func UploadPartHandler(w http.ResponseWriter, r *http.Request) {
	//1.解析用户请求
	r.ParseForm()
	// username := r.Form.Get("username")

	uploadID := r.Form.Get("uploadid")
	chunkIndex := r.Form.Get("index")

	//2.获取redis链接池中的一个链接
	redisConn := redisPool.RedisPool().Get()
	defer redisConn.Close()

	//3.获得文件句柄，用于储存分块内容
	//先创建目录，再创建文件
	filePath := "/data/" + uploadID + "/" + chunkIndex
	os.MkdirAll(path.Dir(filePath), 0744)
	file, err := os.Create(filePath)
	if err != nil {
		w.Write(util.NewRespMsg(-1, "Upload part failed", nil).JSONBytes())
		return
	}
	defer file.Close()

	buf := make([]byte, 1024*1024)

	for {
		n, err := r.Body.Read(buf)
		file.Write(buf[:n])
		if err != nil {
			break
		}
	}
	//4.更新redis缓存状态
	redisConn.Do("HSET", "MP_"+uploadID, "chunkIndex_"+chunkIndex, 1)

	//5.返回处理结果到客户端
	w.Write(util.NewRespMsg(0, "OK", nil).JSONBytes())
}

//todo:合并上传功能
func CompleteUploadHandler(w http.ResponseWriter, r *http.Request) {
	//1.解析请求参数
	r.ParseForm()
	upid := r.Form.Get("upid")
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filename := r.Form.Get("filename")
	filesize, err := strconv.Atoi(r.Form.Get("filesize"))
	if err != nil {
		w.Write(util.NewRespMsg(-1, "params invalid", nil).JSONBytes())
		return
	}
	//2.获得redis链接池中的一个链接
	redisConn := redisPool.RedisPool().Get()
	defer redisConn.Close()
	//3.通过uploadid 查询redis并判断是否所有分块上传完成
	data, err := redis.Values(redisConn.Do("HGETALL", "MP+"+upid))
	if err != nil {
		w.Write(util.NewRespMsg(-1, "complete upload failed", nil).JSONBytes())
		return
	}
	totalCount := 0
	chunkCount := 0

	for i := 0; i <= len(data); i += 2 {
		key := string(data[i].([]byte))
		value := string(data[i+1].([]byte))
		if key == "chunkcount" {
			totalCount, _ = strconv.Atoi(value)
		} else if strings.HasPrefix(key, "chunkIndex_") && value == "1" {
			chunkCount++
		}
	}
	if totalCount != chunkCount {
		w.Write(util.NewRespMsg(-2, "invalid request", nil).JSONBytes())
		return
	}
	//4.todo:合并分块

	//5.更新唯一文件表和用户文件表
	dblayer.OnFileUploadFinshed(filehash, filename, int64(filesize), "")
	dblayer.OnUserFileUploadFinished(username, filehash, filename, int64(filesize))
	//6.响应处理结果
	w.Write(util.NewRespMsg(0, "OK", nil).JSONBytes())
}

//TODO:取消上传分块
func CancelUploadPartHandler(w http.ResponseWriter, r *http.Request) {

}

//TODO：查看文件上传的整体状态
func MultipartUploadStatusHanlder(w http.ResponseWriter, r *http.Request) {

}
