package handler

import (
	dblayer "filestore-server/db"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

//SignupInHandler：处理用户注册GET请求
func SignupInHandler(c *gin.Context) {
	c.Redirect(http.StatusFound, "/static/view/signup.html")
}

////DoSignupHandler：处理用户注册POST请求
//func DoSignupHandler(c *gin.Context) {
//    username := c.Request.FormValue("username")
//    passwd := c.Request.FormValue("password")
//
//    //1.校验用户名和密码
//    if len(username) < 3 || len(passwd) < 5 {
//        c.JSON(
//            http.StatusOK,
//            gin.H{
//                "msg": "Invalid parameter",
//                //这里需要定义枚举
//                "code": -1,
//            })
//        return
//    }
//
//    //2.对密码进行加密
//    encPassword := util.Sha1([]byte(passwd + cfg.PasswordSalt))
//    suc := dblayer.UserSingUp(username, encPassword)
//
//    if suc {
//        c.JSON(
//            http.StatusOK,
//            gin.H{
//                "msg": "Signup successed",
//                //这里需要定义枚举
//                "code": 0,
//            })
//    } else {
//        log.Println("注册失败")
//        c.Status(http.StatusInternalServerError)
//        return
//    }
//}
//
////SignInHandler:处理用户登录GET请求
//func SignInHandler(c *gin.Context) {
//    c.Redirect(http.StatusFound, "/static/view/signup.html")
//}

//DoSigninHandler：处理用户注册POST请求

//UserInfoHandler:查询用户信息
//func UserInfoHandler(c *gin.Context) {
//    //1.解析参数
//    username := c.Request.FormValue("username")
//
//    //2.查询用户信息
//    user, err := dblayer.GetUserInfo(username)
//    if err != nil {
//        log.Printf("query user info error:%s\n", err.Error())
//        c.Status(http.StatusInternalServerError)
//        return
//    }
//    //3.组装并且相应用户数据
//    resp := util.RespMsg{
//        Code: 0,
//        Msg:  "OK",
//        Data: user,
//    }
//    c.Data(http.StatusOK, "application/json", resp.JSONBytes())
//}

//IsTokenValid:验证token有效性
func IsTokenValid(token string, username string) bool {
	if len(token) < 40 {
		return false
	}
	//判断token的时效性 token 有效期为半个小时
	tokenApplyTime := token[32:]
	//
	ts := fmt.Sprintf("%x", time.Now().Unix()-30*60)
	if tokenApplyTime < ts {
		fmt.Println("token is expire")
		return false
	}
	//查询username对应的token信息
	userToken := dblayer.GetUserToken(username)
	//对比两个token是否一致
	if token != userToken {
		fmt.Println("token is invalid 请重新登录")
		return false
	}
	return true
}
