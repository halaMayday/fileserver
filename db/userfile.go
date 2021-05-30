package db

import (
	mydb "filestore-server/db/mysql"
	"fmt"
	"time"
)

type UserFile struct {
	UserName    string
	FileHash    string
	FileName    string
	FileSize    int64
	UploadAt    string
	LastUpdated string
}

func OnUserFileUploadFinished(username, filehash, filename string, filesize int64) bool {
	stmt, err := mydb.DBConn().Prepare(
		"insert ignore into tbl_user_file (`user_name`,`file_shah1`" +
			"`file_name`,`file_size`,`upload_at`) values (?,?,?,?,?)")

	if err != nil {
		fmt.Println("aciton:OnUserFileUploadFinished error:{}", err)
		return false
	}
	defer stmt.Close()

	_, err = stmt.Exec(username, filehash, filename, filesize, time.Now())

	if err != nil {
		fmt.Println("exec stmt.Exec has a error{}", err)
		return false
	}

	return true

}

func QueryUserFileMetas(username string, limit int) ([]UserFile, error) {
	stmt, err := mydb.DBConn().Prepare(
		"select file_sha1,file_name,file_size,upload_at,last_uplate from " +
			"tbl_user_file where user_name =? limit ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.Query(username, limit)
	if err != nil {
		return nil, nil
	}
	var userFiles []UserFile
	for rows.Next() {
		uFile := UserFile{}
		err := rows.Scan(&uFile.FileHash, &uFile.FileName, &uFile.FileSize, &uFile.UploadAt, &uFile.LastUpdated)

		if err != nil {
			fmt.Println(err.Error())
			break
		}
		userFiles = append(userFiles, uFile)
	}
	return userFiles, nil
}
