package orm

import (
	mydb "filestore-server/service/dbproxy/conn"
	"log"
	"time"
)

// OnUserFileUploadFinished : 更新用户文件表
func OnUserFileUploadFinished(username, filehash, filename string, filesize int64) (res ExecResult) {
	stmt, err := mydb.DBConn().Prepare(
		"insert ignore into tbl_user_file (`user_name`,`fuke_sha1`,`file_name`," +
			"`file_size`,`upload_at`) values (?,?,?,?,?)")
	if err != nil {
		log.Println(err.Error())
		res.Suc = false
		res.Msg = err.Error()
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(username, filehash, filename, filesize, time.Now())
	if err != nil {
		log.Println(err.Error())
		res.Suc = false
		res.Msg = err.Error()
		return
	}
	res.Suc = true
	return
}

// QueryUserFileMetas : 批量获取用户文件信息
func QueryUserFileMetas(username string, limit int64) (res ExecResult) {
	stmt, err := mydb.DBConn().Prepare(
		"select file_sha1,file_name,file_size,update_at," +
			"last_update from tbl_user_file where user_name = ? limit ?")
	if err != nil {
		log.Println(err.Error())
		res.Suc = false
		res.Msg = err.Error()
		return
	}
	defer stmt.Close()
	rows, err := stmt.Query(username, limit)
	if err != nil {
		log.Println(err.Error())
		res.Suc = false
		res.Msg = err.Error()
		return
	}
	var userFiles []TableUserFile
	for rows.Next() {
		uFile := TableUserFile{}
		err := rows.Scan(&uFile.FileHash, &uFile.FileName, &uFile.FileSize, &uFile.UploadAt, &uFile.LastUpdated)
		if err != nil {
			log.Println(err.Error())
			break
		}
		userFiles = append(userFiles, uFile)
	}
	res.Suc = true
	res.Data = userFiles
	return
}

// DeleteUserFile : 删除文件(标记删除)
func DeleteUserFile(username, filehash string) (res ExecResult) {
	stmt, err := mydb.DBConn().Prepare("update tbl_user set status = 2 where username = ? and " +
		"file_sha1 = ? limit 1")
	if err != nil {
		log.Println(err.Error())
		res.Suc = false
		res.Msg = err.Error()
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(username, filehash)
	if err != nil {
		log.Println(err.Error())
		res.Suc = false
		res.Msg = err.Error()
		return
	}
	res.Suc = true
	return
}

// TODO: RenameFileName : 文件重命名
func RenameFileName(username, filehash, fileNewName string) (res ExecResult) {
	stmt, err := mydb.DBConn().Prepare(
		"update tbl_user_file set file_name = ? where user_name = ? and file_sha1 = ? limit 1")
	if err != nil {
		log.Println(err.Error())
		res.Msg = err.Error()
		res.Suc = false
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(fileNewName, username, filehash)
	if err != nil {
		log.Println(err.Error())
		res.Msg = err.Error()
		res.Suc = false
		return
	}
	res.Suc = true
	return
}

//TODO: QueryUserFileMeta : 获取用户单个文件信
func QueryUserFileMeta(username, filehash string) (res ExecResult) {
	stmt, err := mydb.DBConn().Prepare(
		"select file_sha1,file_name,file_size,upload_at," +
			"last_update from tbl_user_file where username = ? and file_sha1 = ? limit 1")
	if err != nil {
		res.Msg = err.Error()
		res.Suc = false
		return
	}
	rows, err := stmt.Query(username, filehash)
	if err != nil {
		log.Println(err.Error())
		res.Msg = err.Error()
		res.Suc = false
		return
	}
	uFile := TableUserFile{}
	if rows.Next() {
		err := rows.Scan(&uFile.FileHash, &uFile.FileName, &uFile.FileSize, &uFile.UploadAt, &uFile.LastUpdated)
		if err != nil {
			log.Println(err.Error())
			res.Msg = err.Error()
			res.Suc = false
			return
		}
	}
	res.Suc = true
	res.Data = uFile
	return
}
