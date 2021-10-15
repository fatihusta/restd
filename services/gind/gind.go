package gind

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"reflect"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/untangle/golang-shared/services/logger"
	"github.com/untangle/golang-shared/services/settings"
	"github.com/untangle/restd/services/certmanager"
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

	store := cookie.NewStore([]byte(GenerateRandomString(32)))

	engine.Use(sessions.Sessions("auth_session", store))
	engine.Use(addTokenToSession)

	engine.GET("/", rootHandler)

	// API endpoints
	engine.GET("/testSessions", statusSessions)
	engine.GET("/testInfo", testInfo)
	//engine.GET("/testError")

	engine.POST("/account/login", authRequired())
	engine.POST("/account/logout", authLogout)
	engine.GET("/account/logout", authLogout)
	engine.GET("/account/status", authStatus)

	api := engine.Group("/api")
	api.Use(authRequired())
	api.GET("/status/uid", statusUID)

	// replace packetdProxy with handlers
	api.GET("/status/sessions", packetdProxy)
	api.GET("/status/system", statusSystem)
	api.GET("/status/hardware", statusHardware)
	api.GET("/status/upgrade", statusUpgradeAvailable)
	api.GET("/status/build", statusBuild)
	api.GET("/status/license", statusLicense)
	api.GET("/status/wantest/:device", statusWANTest)
	api.GET("/status/command/find_account", statusCommandFindAccount)
	api.GET("/status/interfaces/:device", packetdProxy)
	api.GET("/status/arp/", statusArp)
	api.GET("/status/arp/:device", statusArp)
	api.GET("/status/dhcp", statusDHCP)
	api.GET("/status/route", statusRoute)
	api.GET("/status/routetables", statusRouteTables)
	api.GET("/status/route/:table", statusRoute)
	api.GET("/status/rules", statusRules)
	api.GET("/status/routerules", statusRouteRules)
	api.GET("/status/wwan/:device", statusWwan)
	api.GET("/status/wifichannels/:device", statusWifiChannels)
	api.GET("/status/wifimodelist/:device", statusWifiModelist)
	api.GET("/status/diagnostics", packetdProxy)

	// todo replace with threatprevention host
	api.GET("/threatprevention/lookup/:host", packetdProxy)

	// todo replace with settings routes
	api.GET("/settings", getSettings)
	api.GET("/settings/*path", getSettings)
	api.POST("/settings", setSettings)
	api.POST("/settings/*path", setSettings)
	api.DELETE("/settings", deleteSettings)
	api.DELETE("/settings/*path", deleteSettings)

	api.GET("/defaults", getDefaultSettings)
	api.GET("/defaults/*path", getDefaultSettings)

	// todo replace with reports routes
	api.Any("/reports/*path", packetdProxy)

	// todo replace with warehouse routes
	api.Any("/warehouse/*path", packetdProxy)

	// todo replace with netspace routess
	api.Any("/netspace/*path", packetdProxy)

	// todo replace with license routes
	api.Any("/license/*path", packetdProxy)

	// todo replace with logging routes
	api.Any("/logging/*path", packetdProxy)

	// todo replace with wireguard routes
	api.Any("/wireguard/*path", packetdProxy)

	// todo replace with classify routes
	api.Any("/classify/*path", packetdProxy)

	// todo replace with logger routes
	api.Any("/logger/*path", packetdProxy)

	// todo replace with logging routes
	api.Any("/debug", packetdProxy)

	// todo replace with gc routes
	api.Any("/gc", packetdProxy)

	// todo replace with fetch-licenses routes
	api.Any("/fetch-licenses", packetdProxy)

	// todo replace with fetch-licenses routes
	api.Any("/factory-reset", packetdProxy)

	// todo replace with upgrade handlers
	api.POST("/sysupgrade", packetdProxy)
	api.POST("/upgrade", packetdProxy)

	// todo replace with reboot/shutdown handlers
	api.Any("/reboot", packetdProxy)
	api.Any("/shutdown", packetdProxy)

	// todo replace with dhcp handlers
	api.POST("/releasedhcp/:device", packetdProxy)
	api.POST("/renewdhcp/:device", packetdProxy)

	prof := engine.Group("/pprof")
	// todo replace with prof handlers
	prof.Any("/*path", packetdProxy)

	// files
	engine.Static("/admin", "/www/admin")

	// handle 404 routes
	engine.NoRoute(noRouteHandler)

	// listen and serve on 0.0.0.0:80
	go engine.Run(":80")

	cert, key := certmanager.GetConfiguredCert()
	go engine.RunTLS(":443", cert, key)

	logger.Info("The RestD engine has been started\n")

}

// Shutdown function here to stop gind service
func Shutdown() {

}

func packetdProxy(c *gin.Context) {
	remote, err := url.Parse("http://localhost:81")
	if err != nil {
		panic(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)
	//Define the director func

	proxy.Director = func(req *http.Request) {
		req.Header = c.Request.Header
		req.Host = remote.Host
		req.URL.Scheme = remote.Scheme
		req.URL.Host = remote.Host
		req.URL.Path = c.Request.URL.Path
	}

	proxy.ServeHTTP(c.Writer, c.Request)
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
		return
	}

	// check if the route is for the admin SPA
	if strings.HasPrefix(c.Request.URL.Path, "/admin/") {
		// check if it is a tidy URL route and not a file request
		ext := path.Ext(c.Request.RequestURI)
		if ext == "" {
			c.File("/www/admin/index.html")
			return
		}
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

func rootHandler(c *gin.Context) {
	if isSetupWizardCompleted() {
		c.Redirect(http.StatusTemporaryRedirect, "/admin")
	} else {
		c.Redirect(http.StatusTemporaryRedirect, "/admin/setup")
	}
}

// returns true if the setup wizard is completed, or false if not
// if any error occurs it returns true (assumes the wizard is completed)
func isSetupWizardCompleted() bool {
	wizardCompletedJSON, err := settings.GetSettings([]string{"system", "setupWizard", "completed"})
	if err != nil {
		logger.Warn("Failed to read setup wizard completed settings: %v\n", err.Error())
		return true
	}
	if wizardCompletedJSON == nil {
		logger.Warn("Failed to read setup wizard completed settings: %v\n", wizardCompletedJSON)
		return true
	}
	wizardCompletedBool, ok := wizardCompletedJSON.(bool)
	if !ok {
		logger.Warn("Invalid type of setup wizard completed setting: %v %v\n", wizardCompletedJSON, reflect.TypeOf(wizardCompletedJSON))
		return true
	}

	return wizardCompletedBool
}

// statusUID returns the UID of the system
func statusUID(c *gin.Context) {
	logger.Debug("statusUID()\n")

	uid, err := settings.GetUIDOpenwrt()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	c.String(http.StatusOK, uid)
}

// addTokenToSession checks for a "token" argument, and adds it to the session
// this is easier than passing it around among redirects
func addTokenToSession(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		return
	}
	logger.Info("Saving token in session: %v\n", token)
	session := sessions.Default(c)
	session.Set("token", token)
	err := session.Save()
	if err != nil {
		logger.Warn("Error saving session: %s\n", err.Error())
	}
	authRequired()
}
