package gind

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
	"net/http"

	"github.com/gin-gonic/gin"
)

var engine *gin.Engine
var logsrc = "gin"

func Startup() {

	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()
	/*
	TODO
	gin.DefaultWriter = logger.NewLogWriter(logsrc)
	gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {
		logger.LogMessageSource(logger.LogLevelDebug, logsrc, "%v %v %v %v\n", httpMethod, absolutePath, handlerName, nuHandlers)
	}
	*/

	engine = gin.New()
	// TODO engine.Use(ginlogger())
	engine.Use(gin.Recovery())
	engine.Use(addHeaders)

	// Allow cross-site for dev - this should be disabled in production
	// config := cors.DefaultConfig()
	// config.AllowAllOrigins = true
	// engine.Use(cors.New(config))

	// A server-side store would be better IMO, but I can't find one.
	// -dmorris
	/*
	TODO
	store := cookie.NewStore([]byte(GenerateRandomString(32)))
	// store := cookie.NewStore([]byte("secret"))

	engine.Use(sessions.Sessions("auth_session", store))
	engine.Use(addTokenToSession)
	*/

	/*
	TODO
	engine.GET("/", rootHandler)

	engine.GET("/ping", pingHandler)

	engine.POST("/account/login", authRequired())
	engine.POST("/account/logout", authLogout)
	engine.GET("/account/logout", authLogout)
	engine.GET("/account/status", authStatus)

	api := engine.Group("/api")
	api.Use(authRequired())

	api.GET("/settings", getSettings)
	api.GET("/settings/*path", getSettings)
	api.POST("/settings", setSettings)
	api.POST("/settings/*path", setSettings)
	api.DELETE("/settings", trimSettings)
	api.DELETE("/settings/*path", trimSettings)

	api.GET("/logging/:logtype", getLogOutput)

	api.GET("/defaults", getDefaultSettings)
	api.GET("/defaults/*path", getDefaultSettings)

	api.POST("/reports/create_query", reportsCreateQuery)
	api.GET("/reports/get_data/:query_id", reportsGetData)
	api.POST("/reports/close_query/:query_id", reportsCloseQuery)

	api.POST("/warehouse/capture", warehouseCapture)
	api.POST("/warehouse/close", warehouseClose)
	api.POST("/warehouse/playback", warehousePlayback)
	api.POST("/warehouse/cleanup", warehouseCleanup)
	api.GET("/warehouse/status", warehouseStatus)
	api.POST("/control/traffic", trafficControl)

	api.POST("/netspace/request", netspaceRequest)

	api.GET("/status/sessions", statusSessions)
	api.GET("/status/system", statusSystem)
	api.GET("/status/hardware", statusHardware)
	api.GET("/status/upgrade", statusUpgradeAvailable)
	api.GET("/status/build", statusBuild)
	api.GET("/status/license", statusLicense)
	api.GET("/status/wantest/:device", statusWANTest)
	api.GET("/status/uid", statusUID)
	api.GET("/status/command/find_account", statusCommandFindAccount)
	api.GET("/status/interfaces/:device", statusInterfaces)
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

	api.GET("/wireguard/keypair", wireguardKeyPair)
	api.POST("/wireguard/publickey", wireguardPublicKey)

	api.GET("/classify/applications", getClassifyAppTable)
	api.GET("/classify/categories", getClassifyCatTable)

	api.GET("/logger/:source", loggerHandler)
	api.GET("/debug", debugHandler)
	api.POST("/gc", gcHandler)

	api.POST("/fetch-licenses", fetchLicensesHandler)
	api.POST("/factory-reset", factoryResetHandler)
	api.POST("/sysupgrade", sysupgradeHandler)
	api.POST("/upgrade", upgradeHandler)

	api.POST("/reboot", rebootHandler)
	api.POST("/shutdown", shutdownHandler)

	api.POST("/releasedhcp/:device", releaseDhcp)
	api.POST("/renewdhcp/:device", renewDhcp)
	*/
	// files
	engine.Static("/admin", "/www/admin")
	engine.Static("/settings", "/www/settings")
	engine.Static("/reports", "/www/reports")
	engine.Static("/setup", "/www/setup")
	engine.Static("/static", "/www/static")
	// handle 404 routes
	engine.NoRoute(noRouteHandler)

	/*
	TODO
	prof := engine.Group("/pprof")
	prof.Use(authRequired())

	prof.GET("/", pprofHandler(pprof.Index))
	prof.GET("/cmdline", pprofHandler(pprof.Cmdline))
	prof.GET("/profile", pprofHandler(pprof.Profile))
	prof.POST("/symbol", pprofHandler(pprof.Symbol))
	prof.GET("/symbol", pprofHandler(pprof.Symbol))
	prof.GET("/trace", pprofHandler(pprof.Trace))
	prof.GET("/block", pprofHandler(pprof.Handler("block").ServeHTTP))
	prof.GET("/goroutine", pprofHandler(pprof.Handler("goroutine").ServeHTTP))
	prof.GET("/heap", pprofHandler(pprof.Handler("heap").ServeHTTP))
	prof.GET("/mutex", pprofHandler(pprof.Handler("mutex").ServeHTTP))
	prof.GET("/threadcreate", pprofHandler(pprof.Handler("threadcreate").ServeHTTP))
	*/

	// listen and serve on 0.0.0.0:80
	engine.Run(":8080")

	/*
	TODO
	cert, key := certmanager.GetConfiguredCert()
	go engine.RunTLS(":443", cert, key)

	logger.Info("The RestD engine has been started\n")
	*/
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

// addTokenToSession checks for a "token" argument, and adds it to the session
// this is easier than passing it around among redirects
/*
TODO
func addTokenToSession(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		return
	}
	// TODO logger.info
	session := sessions.Default(c)
	session.Set("token", token)
	err := session.Save()
	if err != nil {
		// TODO 
		fmt.Println(err)
	}
	authRequired()
}
*/