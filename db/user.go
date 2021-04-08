package db

import (
	mydb "filestore-server/db/mysql"
	"fmt"
)

//用户注册
func UserSingUp(userName string, passWord string) bool {
	stmt, err := mydb.DBConn().Prepare(
		"insert ignore into tbl_user(`user_name`,`user_pwd`) values (?,?)")

	if err != nil {
		fmt.Println("Failed to insert,err:", err.Error())
		return false
	}
	defer stmt.Close()

	ret, err := stmt.Exec(userName, passWord)

	if err != nil {
		fmt.Println("Failed to insert,err:", err.Error())
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
		fmt.Println("username not found: " + username)
		return false
	}

	//pRows := mydb.ParseRows(rows)
	//if len(pRows) > 0 && string(pRows[0]["user_pwd"].([]byte)) == encpwd {
	//	return true
	//}
	return false
}
