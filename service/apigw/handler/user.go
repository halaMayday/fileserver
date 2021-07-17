package handler

import (
	cmn "filestore-server/common"
	cfg "filestore-server/config"
	"filestore-server/service/account/proto"
	"filestore-server/util"
	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro"
	"golang.org/x/net/context"
	"log"
	"net/http"
)

var (
	userCli go_micro_service_user.UserService
)

func init() {
	//registry := consul.NewRegistry(func(options *registry.Options) {
	//    options.Addrs = []string{
	//        "192.168.0.112:8500",
	//        //"172.17.0.5:8500",
	//    }
	//})
	// 创建一个新的服务
	service := micro.NewService(
	//micro.Name("user.Client"),
	//micro.Registry(registry)
	)
	// 初始化
	service.Init()

	// 创建 userClient 客户端
	userCli = go_micro_service_user.NewUserService("go.micro.service.user", service.Client())
}

//SignupInHandler：处理用户注册GET请求
func SignupInHandler(c *gin.Context) {
	c.Redirect(http.StatusFound, "/static/view/signup.html")
}

//DoSignupHandler：处理用户注册POST请求
func DoSignupHandler(c *gin.Context) {
	username := c.Request.FormValue("username")
	passwd := c.Request.FormValue("password")

	resp, err := userCli.Signup(context.TODO(), &go_micro_service_user.ReqSignUp{
		Username: username,
		Password: passwd,
	})
	if err != nil {
		log.Println(err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": resp.Code,
		"msg":  resp.Message,
	})
}

//SignInHandler:处理用户登录GET请求
func SignInHandler(c *gin.Context) {
	c.Redirect(http.StatusFound, "/static/view/signup.html")
}

//DoSigninHandler:处理用户登录POST请求
func DoSigninHandler(c *gin.Context) {
	username := c.Request.FormValue("username")
	password := c.Request.FormValue("password")
	encPassword := util.Sha1([]byte(password + cfg.PasswordSalt))

	rpcResp, err := userCli.Signin(context.TODO(), &go_micro_service_user.ReqSignIn{
		Username: username,
		Password: encPassword,
	})

	if err != nil {
		log.Println(err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}

	if rpcResp.Code != cmn.StatusOK {
		c.JSON(200, gin.H{
			"msg":  "登录失败",
			"code": rpcResp.Code,
		})
		return
	}

	cliResp := util.RespMsg{
		Code: int(cmn.StatusOK),
		Msg:  "登录成功",
		Data: struct {
			Location      string
			Username      string
			Token         string
			UploadEntry   string
			DownloadEntry string
		}{
			Location:      "http://" + c.Request.Host + "/static/view/home.html",
			Username:      username,
			Token:         rpcResp.Token,
			UploadEntry:   cfg.UploadLBHost,
			DownloadEntry: cfg.DownloadLBHost,
		},
	}
	c.Data(http.StatusOK, "application/json", cliResp.JSONBytes())
}

//UserInfoHandler:查询用户信息
func UserInfoHandler(c *gin.Context) {
	//1.解析参数
	username := c.Request.FormValue("username")
	//2.查询用户信息
	resp, err := userCli.UserInfo(context.TODO(), &go_micro_service_user.ReqUserInfo{
		Username: username,
	})
	if err != nil {
		log.Println(err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}

	//3.组装并且相应用户数据
	cliResp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: resp,
	}
	c.Data(http.StatusOK, "application/json", cliResp.JSONBytes())
}
