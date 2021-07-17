package route

import (
	"filestore-server/service/apigw/handler"
	"github.com/gin-gonic/gin"
)

func Router() *gin.Engine {
	router := gin.Default()

	router.Static("/static", "./static")

	//注册
	router.GET("/user/signup", handler.SignupInHandler)
	router.POST("/user/signup", handler.DoSignupHandler)

	//登录
	router.GET("/user/signin", handler.SignInHandler)
	router.POST("/user/signin", handler.DoSigninHandler)

	router.Use(handler.Authorize())

	// 用户查询 TODO:POST请求是否正确
	router.POST("/user/info", handler.UserInfoHandler)

	// 用户文件查询
	router.POST("/file/query", handler.FileQueryHandler)
	// 用户文件修改(重命名)
	router.POST("/file/update", handler.FileMetaUpdateHandler)

	return router
}
