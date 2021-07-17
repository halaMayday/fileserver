package handler

import (
	"context"
	"encoding/json"
	go_micro_service_user "filestore-server/service/account/proto"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
)

//QueryFileHandler:获取文件元信息
func FileQueryHandler(c *gin.Context) {
	username := c.Request.FormValue("username")
	limitCnt, _ := strconv.Atoi(c.Request.FormValue("limit"))
	rpcResp, err := userCli.UserFiles(context.TODO(), &go_micro_service_user.ReqUserFile{
		Username: username,
		Limit:    int32(limitCnt),
	})

	if err != nil {
		log.Println(err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(rpcResp)
	if err != nil {
		log.Println(err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Data(http.StatusOK, "application/json", data)
}

//FileMetaUpdateHandler: 更新元文件信息接口(重命名)  TODO:需要修改
func FileMetaUpdateHandler(c *gin.Context) {
	opType := c.Request.FormValue("op")
	fileSha1 := c.Request.FormValue("filehash")
	username := c.Request.FormValue("username")
	newFileName := c.Request.FormValue("filename")

	if opType != "0" || len(newFileName) < 1 {
		c.Status(http.StatusForbidden)
		return
	}

	rpcResp, err := userCli.UserFileRename(context.TODO(), &go_micro_service_user.ReqUserFileRename{
		Username:    username,
		Filehash:    fileSha1,
		NewFileName: newFileName,
	})

	if err != nil {
		log.Println(err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}

	if len(rpcResp.FileData) <= 0 {
		rpcResp.FileData = []byte("[]")
	}
	c.Data(http.StatusOK, "application/json", rpcResp.FileData)
}
