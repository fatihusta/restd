package gind

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/untangle/golang-shared/services/settings"
)

// getSettings will get the settings of the associated path segments
func getSettings(c *gin.Context) {
	var segments []string

	path := c.Param("path")

	if path == "" {
		segments = nil
	} else {
		segments = removeEmptyStrings(strings.Split(path, "/"))
	}

	jsonResult, err := settings.GetSettings(segments)
	if err != nil {
		c.JSON(http.StatusInternalServerError, jsonResult)
	} else {
		c.JSON(http.StatusOK, jsonResult)
	}
	return
}

// getDefaultSettings will get the settings of the associated path segments
func getDefaultSettings(c *gin.Context) {
	var segments []string

	path := c.Param("path")

	if path == "" {
		segments = nil
	} else {
		segments = removeEmptyStrings(strings.Split(path, "/"))
	}

	jsonResult, err := settings.GetDefaultSettings(segments)
	if err != nil {
		c.JSON(http.StatusInternalServerError, jsonResult)
	} else {
		c.JSON(http.StatusOK, jsonResult)
	}
	return
}

// deleteSettings will delete the settings of the associated path segments
func deleteSettings(c *gin.Context) {
	var segments []string
	path := c.Param("path")

	if path == "" {
		segments = nil
	} else {
		segments = removeEmptyStrings(strings.Split(path, "/"))
	}

	jsonResult, err := settings.TrimSettings(segments)
	if err != nil {
		c.JSON(http.StatusInternalServerError, jsonResult)
	} else {
		c.JSON(http.StatusOK, jsonResult)
	}
	return
}

// setSettings will set the settings of the associated path segments
func setSettings(c *gin.Context) {
	var segments []string
	path := c.Param("path")
	force := c.Query("force")
	forceSync := false

	if path == "" {
		segments = nil
	} else {
		segments = removeEmptyStrings(strings.Split(path, "/"))
	}

	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	var bodyJSONObject interface{}
	err = json.Unmarshal(body, &bodyJSONObject)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
	}

	if force != "" {
		var parseErr error
		forceSync, parseErr = strconv.ParseBool(force)

		if parseErr != nil {
			forceSync = false
		}
	}
	jsonResult, err := settings.SetSettings(segments, bodyJSONObject, forceSync)
	if err != nil {
		c.JSON(http.StatusInternalServerError, jsonResult)
	} else {
		c.JSON(http.StatusOK, jsonResult)
	}
	return
}

// removeEmptyStrings removes any empty strings from the string slice and returns a new slice
func removeEmptyStrings(strings []string) []string {
	b := strings[:0]
	for _, x := range strings {
		if x != "" {
			b = append(b, x)
		}
	}
	return b
}
