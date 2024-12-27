/*
 * Copyright 2023 The RuleGo Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package action

import (
	"database/sql"
	"fmt"
	"github.com/2018yuli/rulego/api/types"
	"github.com/2018yuli/rulego/utils/maps"
	"github.com/2018yuli/rulego/utils/str"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"strings"
)

// 注册节点
func init() {
	Registry.Add(&DbClientNode{})
}

const (
	SELECT = "SELECT"
	INSERT = "INSERT"
	DELETE = "DELETE"
	UPDATE = "UPDATE"
)
const (
	rowsAffectedKey = "rowsAffected"
	lastInsertIdKey = "lastInsertId"
)

// DbClientNodeConfiguration 节点配置
type DbClientNodeConfiguration struct {
	// Sql 操作语句，可以使用${}占位符
	Sql string
	// Params 操作参数，可以是数组或对象
	Params []interface{}
	// GetOne 是否只返回一条记录，返回结构非slice结构
	GetOne bool
	// PoolSize 连接池大小
	PoolSize int
	// DbType 数据库类型，mysql或postgres
	DbType string
	// Dsn 数据库连接配置，参考sql.Open参数
	Dsn string
}

type DbClientNode struct {
	config DbClientNodeConfiguration
	db     *sql.DB
	//操作类型 SELECT\UPDATE\INSERT\DELETE
	opType string
	//参数是否有变量
	paramsHasVar bool
}

// Type 返回组件类型
func (x *DbClientNode) Type() string {
	return "dbClient"
}

func (x *DbClientNode) New() types.Node {
	return &DbClientNode{}
}

// Init 初始化组件
func (x *DbClientNode) Init(ruleConfig types.Config, configuration types.Configuration) error {
	err := maps.Map2Struct(configuration, &x.config)
	if err == nil {
		if x.config.DbType == "" {
			x.config.DbType = "mysql"
		}
		x.db, err = sql.Open(x.config.DbType, x.config.Dsn)
		if err == nil {
			x.db.SetMaxOpenConns(x.config.PoolSize)
			x.db.SetMaxIdleConns(x.config.PoolSize / 2)
			err = x.db.Ping()
			words := strings.Fields(x.config.Sql)
			// opType = SELECT\UPDATE\INSERT\DELETE
			x.opType = strings.ToUpper(words[0])
			//检查操作类型是否支持
			switch x.opType {
			case SELECT, UPDATE, INSERT, DELETE:
				// do nothing
			default:
				err = fmt.Errorf("unsupported sql statement: %s", x.config.Sql)
			}

			//检查是参数否有变量
			for _, item := range x.config.Params {
				if v, ok := item.(string); ok && str.CheckHasVar(v) {
					x.paramsHasVar = true
					break
				}
			}

			//检查是否需要转换成$1风格占位符
			x.config.Sql = str.ConvertDollarPlaceholder(x.config.Sql, x.config.DbType)
		}
	}
	return err
}

// OnMsg 处理消息
func (x *DbClientNode) OnMsg(ctx types.RuleContext, msg types.RuleMsg) error {
	var data interface{}
	var err error
	var rowsAffected int64
	var lastInsertId int64
	sqlStr := str.SprintfDict(x.config.Sql, msg.Metadata.Values())

	var params []interface{}
	if x.paramsHasVar {
		//转换参数变量
		for _, item := range x.config.Params {
			if v, ok := item.(string); ok {
				params = append(params, str.SprintfDict(v, msg.Metadata.Values()))
			} else {
				params = append(params, item)
			}
		}
	} else {
		params = x.config.Params
	}

	switch x.opType {
	case SELECT:
		data, err = x.query(sqlStr, params, x.config.GetOne)
	case UPDATE:
		rowsAffected, err = x.update(sqlStr, params)
	case INSERT:
		rowsAffected, lastInsertId, err = x.insert(sqlStr, params)
	case DELETE:
		rowsAffected, err = x.delete(sqlStr, params)
	default:
		err = fmt.Errorf("unsupported sql statement: %s", sqlStr)
	}

	if err != nil {
		ctx.TellFailure(msg, err)
	} else {
		switch x.opType {
		case SELECT:
			msg.Data = str.ToString(data)
		case UPDATE, DELETE:
			msg.Metadata.PutValue(rowsAffectedKey, str.ToString(rowsAffected))
		case INSERT:
			msg.Metadata.PutValue(rowsAffectedKey, str.ToString(rowsAffected))
			msg.Metadata.PutValue(lastInsertIdKey, str.ToString(lastInsertId))
		}
		ctx.TellSuccess(msg)
	}
	return err
}

// query 查询数据并返回map或slice类型
func (x *DbClientNode) query(sqlStr string, params []interface{}, getOne bool) (interface{}, error) {
	rows, err := x.db.Query(sqlStr, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	// 获取列名和列类型
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	// 创建一个固定大小的 map 和切片，用于存储每一行的数据
	row := make(map[string]interface{}, len(columns))
	values := make([]interface{}, len(columns))

	// 遍历每一列，初始化 interface{} 切片中的值
	for i := range columns {
		var v interface{}
		values[i] = &v
		row[columns[i]] = &v
	}

	// 创建一个空的 map 切片，用于存储最终结果
	result := make([]map[string]interface{}, 0)

	// 遍历结果集中的每一行数据
	for rows.Next() {
		// 调用 rows.Scan 方法，将结果存储在指针切片中
		err = rows.Scan(values...)
		if err != nil {
			return nil, err
		}

		// 将当前行的 map 深拷贝到一个新的 map 中，避免后续循环覆盖数据
		m := make(map[string]interface{}, len(row))
		for k, v := range row {
			// 如果值是 []byte 类型，转换成 string 类型
			if b1, ok := v.(*interface{}); ok {
				if b, ok := (*b1).([]byte); ok {
					v = string(b)
				}
			}
			m[k] = v
		}
		// 将新的 map 追加到结果切片中
		result = append(result, m)
	}

	// 检查是否有错误发生
	if err = rows.Err(); err != nil {
		return nil, err
	}

	if getOne {
		if len(result) > 0 {
			return result[0], nil // 如果只有一条记录，返回map类型
		} else {
			return nil, nil
		}
	} else {
		return result, nil // 否则返回slice类型
	}

}

// update 修改数据并返回影响行数
func (x *DbClientNode) update(sqlStr string, params []interface{}) (int64, error) {
	result, err := x.db.Exec(sqlStr, params...)
	if err != nil {
		return 0, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return rowsAffected, nil
}

// insert 插入数据并返回自增ID
func (x *DbClientNode) insert(sqlStr string, params []interface{}) (int64, int64, error) {
	result, err := x.db.Exec(sqlStr, params...)
	if err != nil {
		return 0, 0, err
	} else {
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return 0, 0, err
		}

		lastInsertId, _ := result.LastInsertId()
		return rowsAffected, lastInsertId, nil
	}
}

// delete 删除数据并返回影响行数
func (x *DbClientNode) delete(sqlStr string, params []interface{}) (int64, error) {
	result, err := x.db.Exec(sqlStr, params...)
	if err != nil {
		return 0, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return rowsAffected, nil
}

// Destroy 销毁组件
func (x *DbClientNode) Destroy() {
	if x.db != nil {
		x.db.Close()
	}
}
