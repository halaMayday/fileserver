package orm

import (
	"database/sql"
	mydb "filestore-server/db/mysql"
	"log"
)

//OnFileUploadFinshed:文件上传完成，信息同步到mysql数据库
func OnFileUploadFinshed(filehash string, filename string,
	filesize int64, fileaddr string) (res ExecResult) {
	//sql脚本预编译
	stmt, err := mydb.DBConn().Prepare(
		"insert ignore into tbl_file(`file_sha1`,`file_name`,`file_size`,`file_addr`,`status`)" +
			" values(?,?,?,?,1)")
	if err != nil {
		log.Println("Failed to prepare statement, err:" + err.Error())
		res.Suc = false
		res.Msg = err.Error()
		return
	}
	defer stmt.Close()
	//执行sql语句
	ret, err := stmt.Exec(filehash, filename, filesize, fileaddr)
	if err != nil {
		log.Println(err.Error())
		res.Suc = false
		res.Msg = err.Error()
		return
	}
	if rf, err := ret.RowsAffected(); nil == err {
		if rf <= 0 {
			log.Printf("File with hash:%s has been uploaded before", filehash)
		}
		res.Suc = true
		return
	}
	res.Suc = false
	return
}

//从mysql中获取文件元信息
func GetFileMeta(filehash string) (res ExecResult) {
	stmt, err := mydb.DBConn().Prepare(
		"select file_sha1,file_addr,file_name,file_size from tbl_file " +
			"where file_sha1=? and status =1 limit 1")

	if err != nil {
		log.Println("Failed to prepare statement,err:", err.Error())
		res.Suc = false
		res.Msg = err.Error()
		return
	}
	defer stmt.Close()
	tFile := TableFile{}
	//执行sql语句
	err = stmt.QueryRow(filehash).Scan(&tFile.FileHash, &tFile.FileAddr, &tFile.FileName, &tFile.FileSize)

	if err != nil {
		if err == sql.ErrNoRows {
			//查询不到对应的记录，返回参数以及错误均为nil
			res.Suc = true
			res.Data = nil
			return
		} else {
			log.Println(err.Error())
			res.Suc = false
			res.Msg = err.Error()
			return
		}
	}
	res.Suc = true
	res.Data = tFile
	return
}

// GetFileMetaList : 从mysql批量获取文件元信息
func GetFileMetaList(limit int64) (res ExecResult) {
	stmt, err := mydb.DBConn().Prepare(
		"select file_sha1,file_addr,file_name,file_size from tbl_file " +
			"where status=1 limit ?")
	if err != nil {
		log.Println(err.Error())
		res.Suc = false
		res.Msg = err.Error()
		return
	}

	defer stmt.Close()

	rows, err := stmt.Query(limit)
	if err != nil {
		log.Println(err.Error())
		res.Suc = false
		res.Msg = err.Error()
		return
	}

	columns, err := rows.Columns()
	if err != nil {
		log.Println(err.Error())
		res.Suc = false
		res.Msg = err.Error()
		return
	}

	values := make([]sql.RawBytes, len(columns))
	var tfiles []TableFile
	for i := 0; i < len(values) && rows.Next(); i++ {
		tfile := TableFile{}
		err = rows.Scan(&tfile.FileHash, &tfile.FileAddr,
			&tfile.FileName, &tfile.FileSize)
		if err != nil {
			log.Println(err.Error())
			break
		}
		tfiles = append(tfiles, tfile)
	}
	res.Suc = true
	res.Data = tfiles
	return
}

// UpdateFileLocation : 更新文件的存储地址(如文件被转移了)
func UpdateFileLocation(filehash string, fileaddr string) (res ExecResult) {
	sqlStr := "update tbl_file set `file_addr` = ? where `file_sha1` = ? limit 1"
	stmt, err := mydb.DBConn().Prepare(sqlStr)
	if err != nil {
		log.Println("预编译sql失败, err:" + err.Error())
		res.Suc = false
		res.Msg = err.Error()
		return
	}

	defer stmt.Close()

	ret, err := stmt.Exec(fileaddr, filehash)

	if err != nil {
		log.Println(err.Error())
		res.Suc = false
		res.Msg = err.Error()
		return
	}

	if rf, err := ret.RowsAffected(); nil == err {
		if rf <= 0 {
			log.Printf("更新文件location失败, filehash:%s", filehash)
			res.Suc = false
			res.Msg = "无记录更新"
			return
		}
		res.Suc = true
		return
	} else {
		res.Suc = false
		res.Msg = err.Error()
		return
	}
}
