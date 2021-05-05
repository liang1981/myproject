//@Title mysql.go
//@Description 完成对MySQL数据库的初始化，提供全局访问的数据库对象Db

package databases

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)
//定义全局数据库对象，用于数据库操作
var Db *sqlx.DB

//初始话数据库连接
func init() {
	database, err := sqlx.Open("mysql", "root:CloudWalk@0819@tcp(127.0.0.1:43306)/rhznsp_database")
	if err != nil {
		fmt.Println("数据库连接失败", err)
		return
	}

	Db = database
}

