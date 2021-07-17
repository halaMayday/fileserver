package handler

import (
	"context"
	"filestore-server/common"
	"filestore-server/config"
	dblayer "filestore-server/db"
	proto "filestore-server/service/account/proto"
	"filestore-server/util"
	"fmt"
	"log"
	"strconv"
	"time"
)

type User struct{}

//GenToken:生成token
func GenToken(username string) string {
	//40位字符：md5(username+timestamp+token_salt)+timestamp[:8]
	ts := fmt.Sprintf("%x", time.Now().Unix())
	tokenPrefix := util.MD5([]byte(username + ts + "_tokensalt"))
	return tokenPrefix + ts[:8]
}

//Signup:处理用户注册请求
func (u *User) Signup(ctx context.Context, req *proto.ReqSignUp, resp *proto.RespSignUp) error {
	username := req.Username
	password := req.Password

	//1.校验用户名和密码
	if len(username) < 3 || len(password) < 5 {
		resp.Code = common.StatusParamInvalid
		resp.Message = "注册参数无效"
		return nil
	}

	//2.对密码进行加密
	encPassword := util.Sha1([]byte(password + config.PasswordSalt))
	suc := dblayer.UserSingUp(username, encPassword)

	if suc {
		resp.Code = common.StatusOK
		resp.Message = "注册成功"
	} else {
		log.Println("注册失败")
		resp.Code = common.StatusRegisterFailed
		resp.Message = "注册失败"
	}
	return nil
}

//Signin:处理用户登录请求
func (u *User) Signin(ctx context.Context, req *proto.ReqSignIn, resp *proto.RespSignIn) error {
	username := req.Username
	password := req.Password

	//1.效验用户名和密码
	pwdChecked := dblayer.UserSignin(username, password)
	if !pwdChecked {
		resp.Code = common.StatusLoginFailed
		return nil
	}

	//2.生成访问凭证
	token := GenToken(username)
	upRes := dblayer.UpdateToken(username, token)
	if !upRes {
		resp.Code = common.StatusServerError
		return nil
	}
	//3.更新用户最后活跃时间
	updateSuc := dblayer.UpdateUserLastOnLineTime(username)
	if !updateSuc {
		resp.Code = common.StatusServerError
		return nil
	}
	//4.登录成功，返回token
	resp.Code = common.StatusOK
	resp.Token = token
	return nil
}

//UserInfo:获取用户信息
func (u *User) UserInfo(ctx context.Context, req *proto.ReqUserInfo, resp *proto.RespUserInfo) error {
	username := req.GetUsername()
	userInfo, err := dblayer.GetUserInfo(username)
	if err != nil {
		resp.Code = common.StatusServerError
		resp.Message = "服务错误"
		return nil
	}

	//相应组装参数
	resp.Code = common.StatusOK
	resp.Username = userInfo.Username
	resp.SignupAt = userInfo.SignupAt
	resp.LastActiveAt = userInfo.LastActiveAt
	userStatus, err := strconv.ParseInt(userInfo.Status, 10, 32)
	if err != nil {
		resp.Code = common.StatusServerError
		resp.Message = "服务错误"
		return nil
	}
	resp.Status = int32(userStatus)
	//todo:需要增加接口，完善用户信息（email/phone等）
	resp.Email = userInfo.Email
	resp.Phone = userInfo.Phone
	return nil
}
