package db

import (
	"database/sql"
	mydb "filestore-server/db/mysql"
	"fmt"
	"log"
)

//文件上传完成，信息同步到mysql数据库
func OnFileUploadFinshed(filehash string, filename string,
	filesize int64, fileaddr string) bool {
	//sql脚本预编译
	stmt, err := mydb.DBConn().Prepare(
		"insert ignore into tbl_file(`file_sha1`,`file_name`,`file_size`,`file_addr`,`status`)" +
			" values(?,?,?,?,1)")
	if err != nil {
		fmt.Println("Failed to prepare statement,err:", err.Error())
		return false
	}
	defer stmt.Close()
	//执行sql语句
	ret, err := stmt.Exec(filehash, filename, filesize, fileaddr)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	if rf, err := ret.RowsAffected(); nil == err {
		if rf <= 0 {
			fmt.Printf("File with hash:%s has been uploaded before", filehash)
		}
		return true
	}
	return false
}

type TableFile struct {
	FileHash string
	FileName sql.NullString
	FileSize sql.NullInt64
	FileAddr sql.NullString
}

//从mysql中获取文件元信息
func GetFileMeta(filehash string) (*TableFile, error) {
	stmt, err := mydb.DBConn().Prepare(
		"select file_sha1,file_addr,file_name,file_size from tabl_file " +
			"where file_hash=? and status =1 limit 1")

	if err != nil {
		fmt.Println("Failed to prepare statement,err:", err.Error())
		return nil, err
	}
	defer stmt.Close()
	tFile := TableFile{}
	//执行sql语句
	err = stmt.QueryRow(filehash).Scan(&tFile.FileHash, &tFile.FileName, &tFile.FileSize, &tFile.FileAddr)

	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	return &tFile, nil
}

//UpdateFileLocation:更新文件的储存地址(例如文件被转移了)
func UpdateFileLocation(filehash, fileaddr string) bool {
	sqlStr := "update tbl_file set `file_addr` = ? where `file_sha1` = ? limit 1"
	stmt, err := mydb.DBConn().Prepare(sqlStr)
	if err != nil {
		log.Println("预编译sql失败,err:{}", err.Error())
		return false
	}

	defer stmt.Close()

	ret, err := stmt.Exec(fileaddr, filehash)

	if err != nil {
		log.Println("执行sql:{}失败,err:{}", sqlStr, err.Error())
		return false
	}

	if rf, err := ret.RowsAffected(); nil == err {
		if rf <= 0 {
			log.Printf("更新文件location失败, filehash:%s", filehash)
			return false
		}
		return true
	}
	return false
}
