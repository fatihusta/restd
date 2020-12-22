package gind

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jsommerville-untangle/golang-shared/services/logger"
)

var engine *gin.Engine
var logsrc = "gin"

func Startup() {

	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()
	gin.DefaultWriter = logger.NewLogWriter(logsrc)
	gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {
		logger.LogMessageSource(logger.LogLevelDebug, logsrc, "%v %v %v %v\n", httpMethod, absolutePath, handlerName, nuHandlers)
	}
	

	engine = gin.New()
	engine.Use(ginlogger())
	engine.Use(gin.Recovery())
	engine.Use(addHeaders)

	engine.GET("/test", statusSessions)

	// files
	engine.Static("/admin", "/www/admin")
	engine.Static("/settings", "/www/settings")
	engine.Static("/reports", "/www/reports")
	engine.Static("/setup", "/www/setup")
	engine.Static("/static", "/www/static")
	// handle 404 routes
	engine.NoRoute(noRouteHandler)

	// listen and serve on 0.0.0.0:80
	// TODO change to :80 once take out packetd restd
	go engine.Run(":8080")

	logger.Info("The RestD engine has been started\n")
	

	
}

func Shutdown() {

}

// GenerateRandomString generates a random string of the specified length
func GenerateRandomString(n int) string {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		// TODO logger.Warn("Failed to generated secure key: %v\n", err)
		return "secret"
	}
	return base64.URLEncoding.EncodeToString(b)
}

// handles 404 routes
func noRouteHandler(c *gin.Context) {
	// MFW-704 - return 200 for JS map files requested by Safari on Mac
	if strings.Contains(c.Request.URL.Path, ".js.map") {
		c.String(http.StatusOK, "")
	}
	// otherwise browser will default to its 404 handler
}

func addHeaders(c *gin.Context) {
	c.Header("Cache-Control", "must-revalidate")
	// c.Header("Example-Header", "foo")
	// c.Header("Access-Control-Allow-Origin", "*")
	// c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE")
	// c.Header("Access-Control-Allow-Headers", "X-Custom-Header")
	c.Next()
}

func ginlogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.LogMessageSource(logger.LogLevelDebug, logsrc, "%v %v\n", c.Request.Method, c.Request.RequestURI)
		c.Next()
	}
}