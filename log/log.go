//@Title log.go
//@Description 服务服务器的日志记录

package log

import (
	"io"
	"os"

	"github.com/gin-gonic/gin"
)

func MyLog() {
	gin.DisableConsoleColor()

	// Logging to a file.
	f, _ := os.Create("gin.log")
	gin.DefaultWriter = io.MultiWriter(f)

	// 如果需要同时将日志写入文件和控制台，请使用以下代码。
	// gin.DefaultWriter = io.MultiWriter(f, os.Stdout)
}