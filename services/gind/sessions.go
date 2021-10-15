package gind

// import (
// 	"net/http"

// 	"github.com/gin-gonic/gin"
// 	"github.com/untangle/golang-shared/services/logger"
// )

// // statusSessions is the RESTD /api/status/sessions handler
// func statusSessions(c *gin.Context) {
// 	logger.Debug("statusSession()\n")

// 	sessions, err := getSessions()
// 	if err != nil {
// 		logger.Warn(err.Error(), "\n")
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, sessions)
// }

// // getSessions sends the GET_SESSIONS, gets reply, and retrives the Session item
// func getSessions() ([]map[string]interface{}, error) {
// 	reply, err := messenger.SendRequestAndGetReply(messenger.Packetd, messenger.GetSessions)
// 	if err != nil {
// 		return nil, err
// 	}

// 	logger.Debug("received reply: ", reply)

// 	sessions, err := messenger.RetrievePacketdReplyItem(reply, messenger.GetSessions)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return sessions, nil
// }
