package api

import (
	cfg "filestore-server/config"
	dblayer "filestore-server/db"
	redisPool "filestore-server/cache/redis"
	dbcli "filestore-server/service/dbproxy/client"
	"filestore-server/util"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/gin-gonic/gin"
	"log"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
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
func InitalMultipartUploadHandler(c *gin.Context) {
	//1.解析用户请求参数
	username := c.Request.FormValue("username")
	filehash := c.Request.FormValue("filehash")
	filesize, err := strconv.Atoi(c.Request.FormValue("filesize"))

	if err != nil {
		c.JSON(
			http.StatusOK,
			gin.H{
				"code": -1,
				"msg":  "params invalid",
			})
		return
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
	c.JSON(
		http.StatusOK,
		gin.H{
			"code": 0,
			"msg":  "ok",
			"data": upInfo,
		})
}

//UploadPartHandler:上传文件分块
func UploadPartHandler(c *gin.Context) {
	//1.解析用户请求
	uploadID := c.Request.FormValue("uploadid")
	chunkIndex := c.Request.FormValue("index")

	//2.获取redis链接池中的一个链接
	redisConn := redisPool.RedisPool().Get()
	defer redisConn.Close()

	//3.获得文件句柄，用于储存分块内容
	//先创建目录，再创建文件
	filePath := cfg.TempLocalRootDir + uploadID + "/" + chunkIndex
	os.MkdirAll(path.Dir(filePath), 0744)
	file, err := os.Create(filePath)
	if err != nil {
		c.JSON(
			http.StatusOK,
			gin.H{
				"coder": -1,
				"msg":   "upload part failed",
				"data":  nil,
			})
		return
	}
	defer file.Close()

	buf := make([]byte, 1024*1024)

	for {
		n, err := c.Request.Body.Read(buf)
		file.Write(buf[:n])
		if err != nil {
			break
		}
	}
	//4.更新redis缓存状态
	redisConn.Do("HSET", "MP_"+uploadID, "chunkIndex_"+chunkIndex, 1)

	//5.返回处理结果到客户端
	c.JSON(
		http.StatusOK,
		gin.H{
			"code": 0,
			"msg":  "OK",
			"data": nil,
		})
}

//CompleteUploadHandler:合并上传通知
func CompleteUploadHandler(c *gin.Context) {
	//1.解析请求参数
	upid := c.Request.FormValue("upid")
	username := c.Request.FormValue("username")
	filehash := c.Request.FormValue("filehash")
	filename := c.Request.FormValue("filename")
	filesize := c.Request.FormValue("filesize")

	//2.获得redis链接池中的一个链接
	redisConn := redisPool.RedisPool().Get()
	defer redisConn.Close()

	//3.通过uploadid 查询redis并判断是否所有分块上传完成
	data, err := redis.Values(redisConn.Do("HGETALL", "MP+"+upid))
	if err != nil {
		c.JSON(
			http.StatusOK,
			gin.H{
				"code": -1,
				"msg":  "complete upload failed",
				"data": nil,
			})
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
		c.JSON(
			http.StatusOK,
			gin.H{
				"code": -2,
				"msg":  "分块不完整",
				"data": nil,
			})
		return
	}

	// 4. TODO：合并分块, 可以将ceph当临时存储，合并时将文件写入ceph;
	// 也可以不用在本地进行合并，转移的时候将分块append到ceph/oss即可
	srcPath := cfg.TempPartRootDir + upid + "/"
	destPath := cfg.TempLocalRootDir + filehash
	cmd := fmt.Sprintf("cd %s && ls | sort -n | xargs cat > %s", srcPath, destPath)
	mergeResp, err := util.ExecLinuxShell(cmd)
	if err != nil {
		log.Println(err)
		c.JSON(
			http.StatusOK,
			gin.H{
				"code": -2,
				"msg":  "合并失败",
				"data": nil,
			})
	}
	log.Println(mergeResp)

	//5.更新唯一文件表和用户文件表
	fSize, _ := strconv.Atoi(filesize)
	fmeta := dbcli.FileMeta{
		FileSha1: filehash,
		FileName: username,
		FileSize: int64(fSize),
		Location: destPath,
	}

	_, ferr := dbcli.OnFileUploadFinished(fmeta)
	_, uferr := dbcli.OnUserFileUploadFinished(username, fmeta)
	if ferr != nil || uferr != nil {
		log.Println(err)
		c.JSON(
			http.StatusOK,
			gin.H{
				"code": -2,
				"msg":  "数据更新失败",
				"data": nil,
			})
		return
	}
	//6.响应处理结果
	c.JSON(
		http.StatusOK,
		gin.H{
			"code": 0,
			"msg":  "OK",
			"data": nil,
		})

}

//TODO:取消上传分块
func CancelUploadPartHandler(c *gin.Context) {}

//TODO：查看文件上传的整体状态
func MultipartUploadStatusHanlder(c *gin.Context) {}
