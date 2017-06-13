package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type service struct {
	gin                  *gin.Engine
	wsUpgrder            *websocket.Upgrader
	currentSubscriptions subscriptions
	timeout              time.Duration
	sync                 chan bool
	cache                *Cache
}

func newService(
	currentSubscriptions subscriptions,
	readBufferSize int,
	writeBufferSize int,
	timeout time.Duration,
	debugMode string,
	cache *Cache,
) *service {

	if debugMode == "DEBUG" {
		gin.SetMode(gin.ReleaseMode)
	}

	upgrader := &websocket.Upgrader{
		ReadBufferSize:  readBufferSize,
		WriteBufferSize: writeBufferSize,
		CheckOrigin:     func(request *http.Request) bool { return true },
	}

	return &service{
		gin.Default(),
		upgrader,
		currentSubscriptions,
		timeout,
		make(chan bool, 1),
		cache,
	}

}

func (s *service) run(
	listen string,
) {

	v1 := s.gin.Group("v1")
	{
		v1.GET("/subscribe", s.handleSubscription)
		v1.PUT("/publish", s.handlePublish)
		v1.GET("/stats", s.handleStats)
		v1.GET("/stats/sessions", s.handleStatsSessions)
		v1.GET("/health", s.handleHealth)
	}

	//	ginpprof.Wrap(s.gin)
	logger.Debugf("run service on %s", listen)
	s.gin.Run(listen)
}

func (s *service) handleSubscription(context *gin.Context) {

	logger.Debugf(
		"handling new request from %s",
		context.Request.RemoteAddr,
	)

	authInfo := make(chan string)
	isDisconnected := make(chan bool, 1)
	terminateAuthentication := make(chan bool, 1)
	terminateKeepAlive := make(chan bool, 1)
	incomingMessages := make(chan string)

	connection, err := s.wsUpgrder.Upgrade(
		context.Writer, context.Request, nil,
	)

	if err != nil {
		logger.Errorf(
			"something happing wrong during create websocket connection, "+
				"reason %s",
			err.Error(),
		)

		return
	}

	defer func() {
		err := connection.Close()

		if err != nil {
			logger.Errorf(
				"unable to close websocket connection from %s, reason: %s",
				connection.RemoteAddr().String(),
				err.Error(),
			)

			return
		}

		logger.Infof(
			"websocket connection successfully closed for client %s",
			connection.RemoteAddr().String(),
		)
		return
	}()

	connection.SetReadDeadline(time.Now().Add(s.timeout))
	connection.SetWriteDeadline(time.Now().Add(s.timeout))

	go readMessage(connection, isDisconnected, incomingMessages)
	go authenticateUser(
		incomingMessages,
		authInfo,
		terminateAuthentication,
	)

	var (
		authToken     string = ""
		sessionNumber int    = 0
		messages      chan []byte
	)

	for {
		select {
		case authToken = <-authInfo:
			logger.Debugf(
				"auth token received: %s",
				authToken,
			)

			go keepAlive(connection, s.timeout, terminateKeepAlive, s.sync)

			s.sync <- true
			sessionNumber, messages = s.currentSubscriptions.subscribeUser(
				authToken,
			)
			<-s.sync

			logger.Infof(
				"user %s was subscribed with ID %d",
				connection.RemoteAddr().String(),
				sessionNumber,
			)

			if len(s.cache.messages[authToken]) != 0 {
				logger.Debugf(
					"pop messages from cache for %s",
					authToken,
				)

				go func() {
					for _, cacheObject := range s.cache.pop(authToken) {
						messages <- []byte(cacheObject.message)
					}
				}()
			}

		case <-isDisconnected:
			s.sync <- true

			logger.Infof(
				"client %s with token %s with session %d will be disconnected",
				context.Request.RemoteAddr,
				authToken,
				sessionNumber,
			)

			terminateAuthentication <- true
			terminateKeepAlive <- true
			logger.Infof(
				"terminate message was successfully sended into keepalive" +
					" and authentication processes",
			)

			activeSubscribers := s.currentSubscriptions.getActiveSubscribers(
				authToken,
			)

			switch activeSubscribers {
			case 0:
				logger.Debug("there are no subscribers, exiting...")

			case 1:
				logger.Debugf(
					"there is only one subscription for user %s, "+
						"unsubscribe user",
					authToken,
				)

				s.currentSubscriptions.unsubscribeUser(authToken)

			default:
				logger.Debugf(
					"there are several sessions for user %s, "+
						"trying deactivate session %d",
					authToken,
					sessionNumber,
				)

				s.currentSubscriptions.setInactive(authToken, sessionNumber)
			}

			<-s.sync
			return

		case message := <-messages:
			s.sync <- true
			logger.Debugf("got message %s", message)

			err := connection.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				logger.Errorf(
					"%s: can`t write message to websocket, reason %s",
					reflect.TypeOf(err).String(),
					err.Error(),
				)
				<-s.sync

				return
			}
			<-s.sync
		}
	}
}

func (s *service) handlePublish(context *gin.Context) {
	authToken := strings.TrimSpace(
		context.Request.Header.Get("Token"),
	)
	message, err := ioutil.ReadAll(context.Request.Body)

	if err != nil {
		logger.Errorf(
			"could not read response body, reason: %s",
			err.Error(),
		)
		context.JSON(makeInternalServerError(err.Error(), nil))
		return
	}

	if _, exists := s.currentSubscriptions[authToken]; exists {

		s.currentSubscriptions.publishMessage(authToken, message)

		context.JSON(makeStatusOk("message published", string(message)))
		return
	}

	s.cache.put(authToken, string(message))

	context.JSON(makeNotFoundError(
		fmt.Sprintf("recipient %s not found", authToken),
		string(message),
	))
	return
}

func (s *service) handleStats(context *gin.Context) {
	stat := ServiceStats{
		SubscribersCount: len(s.currentSubscriptions),
	}

	context.JSON(makeStatusOk(
		context.Request.URL.String(), stat,
	))

	return
}

func (s *service) handleStatsSessions(context *gin.Context) {

	sessionsCount := make(map[string]int)
	for authToken, sessions := range s.currentSubscriptions {
		sessionsCount[authToken] = len(sessions)

	}
	sortedCurrentSessions := newSortedSessions(sessionsCount)
	sortedCurrentSessions.Sort()

	stat := ActiveSessions{
		SubscribersSessionsCount: sessionsCount,
	}

	context.JSON(makeStatusOk(
		context.Request.URL.String(), stat,
	))

	return
}
func (s *service) handleHealth(context *gin.Context) {

	context.JSON(makeStatusOk(
		context.Request.URL.String(), "OK",
	))

	return
}
