package gind

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/untangle/golang-shared/services/logger"
	"github.com/untangle/restd/services/messenger"
)

func reportsGetData(c *gin.Context) {
	queryStr := c.Param("query_id")
	if queryStr == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query_id not found"})
		return
	}

	reply, err := messenger.SendRequestAndGetReply(messenger.Reportd, messenger.QueryData, queryStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	result, err := messenger.RetrieveReportdReplyItem(reply, messenger.QueryData)
	// StatusOK has to be used so UI doesn't fail to finish loading reports
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}
	queryData, ok := result["result"].(string)
	if !ok {
		c.JSON(http.StatusOK, gin.H{"error": "failed_to_get_data_query"})
		return
	}

	c.Header("Content-Type", "application/json")
	c.String(http.StatusOK, queryData)
	return
}

func reportsCreateQuery(c *gin.Context) {
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	reply, err := messenger.SendRequestAndGetReply(messenger.Reportd, messenger.QueryCreate, string(body))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	result, err := messenger.RetrieveReportdReplyItem(reply, messenger.QueryCreate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	queryID, ok := result["result"].(uint64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_create_query"})
		return
	}
	str := fmt.Sprintf("%d", queryID)
	logger.Debug("CreateQuery(%s)\n", str)
	c.String(http.StatusOK, str)
}

func reportsCloseQuery(c *gin.Context) {
	queryStr := c.Param("query_id")
	if queryStr == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query_id not found"})
		return
	}
	result, err := messenger.SendRequestAndGetReply(messenger.Reportd, messenger.QueryClose, queryStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	reply, err := messenger.RetrieveReportdReplyItem(result, messenger.QueryClose)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	success, ok := reply["result"].(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_close_query"})
		return
	}
	c.String(http.StatusOK, success)
	return
}
