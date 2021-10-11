package gind

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/untangle/golang-shared/services/logger"
	"github.com/untangle/restd/services/messenger"
)

var engine *gin.Engine
var logsrc = "gin"

// Startup starts the gin server
func Startup() {
	// Set some gin properties
	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()
	gin.DefaultWriter = logger.NewLogWriter(logsrc)
	gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {
		logger.LogMessageSource(logger.LogLevelDebug, logsrc, "%v %v %v %v\n", httpMethod, absolutePath, handlerName, nuHandlers)
	}

	// Create gin engine
	engine = gin.New()
	engine.Use(ginlogger())
	engine.Use(gin.Recovery())
	engine.Use(addHeaders)

	engine.GET("/ping", pingHandler)

	// API endpoints
	engine.GET("/testSessions", statusSessions)
	engine.GET("/testInfo", testInfo)
	//engine.GET("/testError")

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

// Shutdown function here to stop gind service
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

// addHeaders adds the gin headers
func addHeaders(c *gin.Context) {
	c.Header("Cache-Control", "must-revalidate")
	// c.Header("Example-Header", "foo")
	// c.Header("Access-Control-Allow-Origin", "*")
	// c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE")
	// c.Header("Access-Control-Allow-Headers", "X-Custom-Header")
	c.Next()
}

// ginLogger creates function for logging
func ginlogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.LogMessageSource(logger.LogLevelDebug, logsrc, "%v %v\n", c.Request.Method, c.Request.RequestURI)
		c.Next()
	}
}

// testInfo sends request and parses the testInfo packetd response for testing ZMQ and restd
// basic format for gin handlers
func testInfo(c *gin.Context) {
	logger.Debug("testInfo()\n")

	// Send the PACKETD TEST_INFO request and get the reply
	reply, err := messenger.SendRequestAndGetReply(messenger.Packetd, messenger.TestInfo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Debug("received reply: ", reply)

	// Retrieve the TEST_INFO information
	info, err := messenger.RetrievePacketdReplyItem(reply, messenger.TestInfo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, info)
}

func pingHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}
