package route

import (
	"filestore-server/handler"
	"github.com/gin-gonic/gin"
)

func Router() *gin.Engine {
	//gin framework 包括Logger,Recovery
	router := gin.Default()
	//处理静态资源
	router.Static("/static/", "./static")

	//注册
	router.GET("/user/signup", handler.SignupInHandler)
	router.POST("/user/signup", handler.DoSignupHandler)
	//登录
	router.GET("/user/signin", handler.SignInHandler)
	router.POST("/user/signin", handler.DoSigninHandler)

	//加入中间件，用于效验token的拦截器
	router.Use(handler.HTTPInterceptor())

	//use之后的所有handler都会进去拦截

	//用户信息查询
	router.GET("/user/info", handler.UserInfoHandler)

	//文件存取接口
	router.GET("/file/upload", handler.UploadHandler)
	router.POST("/file/upload", handler.DoUploadHandler)
	router.GET("/file/query", handler.QueryFileHandler)
	router.POST("/file/update", handler.FileMetaUpdataHandle)
	router.POST("/file/delete", handler.FileDeleteHandle)
	router.POST("/file/download", handler.DownloadHandler)
	router.POST("/file/downloadurl", handler.DownloadURLHandler)

	//秒传接口
	router.POST("/file/fastupload", handler.TryFastUploadHandler)

	//分块上传接口:TODO
	//分块上传接口
	router.POST("/file/mpupload/init", handler.InitalMultipartUploadHandler)
	router.POST("/file/mpupload/uppart", handler.UploadPartHandler)
	router.POST("/file/mpupload/finshed", handler.CompleteUploadHandler)
	router.POST("/file/mpupload/cancel", handler.CancelUploadPartHandler)
	router.POST("/file/mpupload/status", handler.MultipartUploadStatusHanlder)

	return router
}
