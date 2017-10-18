package main

import (
	"fmt"
	"io/ioutil"

	"github.com/gin-gonic/gin"
)

func (s *service) handleBroadcast(context *gin.Context) {
	message, err := ioutil.ReadAll(context.Request.Body)

	if err != nil {
		logger.Errorf(
			"could not read response body, reason: %s",
			err.Error(),
		)
		context.JSON(makeInternalServerError(err.Error(), nil))
		return
	}

	s.sync <- true
	defer func() {
		<-s.sync
	}()

	for authToken := range s.currentSubscriptions {
		s.currentSubscriptions.publishMessage(authToken, message)
	}

	context.JSON(
		makeStatusOk(
			fmt.Sprintf(
				"message sent to %d recipients",
				len(s.currentSubscriptions),
			),
			string(message),
		),
	)
	return
}
