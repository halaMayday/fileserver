package db

import (
	mydb "filestore-server/db/mysql"
	"fmt"
)

//用户注册
func UserSingUp(userName string, passWord string) bool {
	stmt, err := mydb.DBConn().Prepare(
		"insert ignore into tbl_user(`user_name`,`user_pwd`) values (?,?)")

	defer stmt.Close()

	if err != nil {
		fmt.Println("Failed to insert,err:", err)
		return false
	}

	ret, err := stmt.Exec(userName, passWord)

	if err != nil {
		fmt.Println("Failed to insert,err:", err)
		return false
	}

	if rf, err := ret.RowsAffected(); nil == err {
		if rf > 0 {
			return true
		} else {
			fmt.Printf("File with hash:%s has been signup before", userName)
		}
	}

	return false
}

// UserSignin : 判断密码是否一致
func UserSignin(username string, encpwd string) bool {
	stmt, err := mydb.DBConn().Prepare("select * from tbl_user where user_name=? limit 1")
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	defer stmt.Close()

	rows, err := stmt.Query(username)
	if err != nil {
		fmt.Println(err.Error())
		return false
	} else if rows == nil {
		//TODO: 当查询的username为空的时候，rows == nil 也不成立。这里的逻辑有问题
		fmt.Println("username not found: " + username)
		return false
	}
	pRows := mydb.ParseRows(rows)
	if len(pRows) > 0 && string(pRows[0]["user_pwd"].([]byte)) == encpwd {
		return true
	}
	return false
}

func UpdateToken(username string, token string) bool {
	stmt, err := mydb.DBConn().Prepare(
		"replace into tbl_user_token (`user_name`,`user_token`) values (?,?)")
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	defer stmt.Close()

	_, err = stmt.Exec(username, token)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}

//定义user的结构体

type User struct {
	Username     string
	Email        string
	Phone        string
	SignupAt     string
	LastActiveAt string
	Status       string
}

type UserToken struct {
	Username  string
	Usertoken string
}

func GetUserInfo(username string) (User, error) {
	user := User{}

	stmt, err := mydb.DBConn().Prepare(
		"select  user_name,signup_at from tbl_user where user_name = ? limit 1")

	if err != nil {
		fmt.Println(err.Error())
		return user, err
	}

	defer stmt.Close()

	err = stmt.QueryRow(username).Scan(&user.Username, &user.SignupAt)
	if err != nil {
		return user, err
	}
	return user, nil
}

func GetUserToken(username string) string {
	token := ""
	stmt, err := mydb.DBConn().Prepare("select user_token from tbl_user_token where user_name = ? limit 1")

	if err != nil {
		fmt.Println(err.Error())
		return token
	}
	defer stmt.Close()
	err = stmt.QueryRow(username).Scan(&token)
	if err != nil {
		fmt.Println("查询用户:{}的token错误,原因是{}", username, err)
	}
	return token
}
