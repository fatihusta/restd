package gind

import (
	"github.com/gin-gonic/gin"
)

var engine *gin.Engine
var logsrc = "gin"

func Startup() {
	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()
	
	engine = gin.New()
}
