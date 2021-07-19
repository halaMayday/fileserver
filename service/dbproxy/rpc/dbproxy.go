package rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"filestore-server/service/dbproxy/mapper"
	"filestore-server/service/dbproxy/orm"
	dbProxy "filestore-server/service/dbproxy/proto"
)

type DBProxy struct{}

//ExecuteAction:请求执行SQL函数
func ExecuteAction(ctx context.Context, req *dbProxy.ReqExec, resp *dbProxy.RespExec) error {
	resList := make([]orm.ExecResult, len(req.Action))

	//TODO:檢查 req.Sequence req.transaction兩個參數，執行不同的流程
	for idx, singleAction := range req.Action {
		var params []interface{}
		dec := json.NewDecoder(bytes.NewReader(singleAction.Params))
		// 避免int/int32/int64等自动转换为float64
		dec.UseNumber()
		if err := dec.Decode(&params); err != nil {
			resList[idx] = orm.ExecResult{
				Suc: false,
				Msg: "请求参数错误",
			}
			continue
		}

		for k, v := range params {
			if _, ok := v.(json.Number); ok {
				params[k], _ = v.(json.Number).Int64()
			}
		}
		//默认串行执行sql函数
		//TODO:这块不太懂
		execRes, err := mapper.FuncCall(singleAction.Name, params...)
		if err != nil {
			resList[idx] = orm.ExecResult{
				Suc: false,
				Msg: "函数调用错误",
			}
			continue
		}
		resList[idx] = execRes[0].Interface().(orm.ExecResult)
	}
	//TODO:处理异常
	resp.Data, _ = json.Marshal(resList)
	return nil
}
