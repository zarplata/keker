package main

import (
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const helloMessage = "HELLO"

func readMessage(
	connection *websocket.Conn,
	isDisconnected chan bool,
	incomingMessages chan string,
) {
	for {
		messageType, message, err := connection.ReadMessage()
		if err != nil {
			switch err.(type) {
			case *websocket.CloseError:
				logger.Debugf(
					"handle CLOSE frame, client %s will be disconnected",
					connection.RemoteAddr().String(),
				)
				isDisconnected <- true

			default:
				logger.Errorf(
					"unexpected error type %s with message %s",
					reflect.TypeOf(err).String(),
					err.Error(),
				)
				isDisconnected <- true
			}
			return
		}

		switch messageType {
		case websocket.PingMessage, websocket.PongMessage:
			logger.Debugf("got PING-PONG frame, ignore it")
			continue

		case websocket.TextMessage:
			incomingMessages <- string(message)

		default:
			logger.Errorf(
				"unexpected message type %s ",
				messageType,
			)
		}
	}
}

func authenticateUser(
	incomingMessages chan string,
	isAuthenticated chan string,
	terminateAuthentication chan bool,
) {

	isAlreadyAuthenticated := false

	for {

		select {
		case <-terminateAuthentication:
			logger.Debug(
				"authentication process terminated",
			)
			return

		case incomingMessage := <-incomingMessages:
			logger.Debugf(
				"got new message from client: %s",
				incomingMessage,
			)

			if strings.HasPrefix(incomingMessage, helloMessage) {
				if isAlreadyAuthenticated {
					logger.Errorf(
						"session already authenticated",
					)
					continue
				}

				tokens := strings.Fields(incomingMessage)
				if len(tokens) != 2 {
					logger.Errorf(
						"unexpected HELLO message: %s",
						incomingMessage,
					)
					continue
				}

				authToken := strings.TrimSpace(tokens[1])
				logger.Debugf(
					"user with auth token %s successfully authenticated",
					authToken,
				)

				isAuthenticated <- authToken
				isAlreadyAuthenticated = true

				continue
			}
		}
	}
}

func keepAlive(
	connection *websocket.Conn,
	timeout time.Duration,
	terminateKeepAlive chan bool,
	sync chan bool,
) {
	lastResponse := time.Now()
	keepaliveStartTime := lastResponse

	connection.SetPongHandler(func(msg string) error {
		logger.Debugf(
			"handle PONG frame from %s with message %s",
			connection.RemoteAddr().String(),
			msg,
		)

		lastResponse = time.Now()
		return nil
	})

	for {
		select {
		case <-terminateKeepAlive:
			logger.Debugf(
				"keepalive process of %s terminated",
				connection.RemoteAddr().String(),
			)
			return

		default:
			connection.SetReadDeadline(time.Now().Add(timeout))
			connection.SetWriteDeadline(time.Now().Add(timeout))

			logger.Debugf(
				"sending PING frame to client %s",
				connection.RemoteAddr().String(),
			)

			sync <- true
			err := connection.WriteMessage(
				websocket.PingMessage,
				[]byte("ping"),
			)
			<-sync
			if err != nil {
				logger.Errorf(
					"unable to write PING frame to %s, reason %s",
					connection.RemoteAddr().String(),
					err.Error(),
				)

				return
			}

			time.Sleep(timeout / 2)

			if time.Now().Sub(lastResponse) > timeout {
				logger.Errorf(
					"ping timeout exceeded from %s, "+
						"keepalive session length %s",
					connection.RemoteAddr().String(),
					time.Now().Sub(keepaliveStartTime).String(),
				)

				return
			}
		}
	}
}

type sortedSessions struct {
	AuthTokens []string
	Sessions   []int
}

func newSortedSessions(
	rawSessions map[string]int,
) *sortedSessions {
	sessions := &sortedSessions{
		AuthTokens: make([]string, 0, len(rawSessions)),
		Sessions:   make([]int, 0, len(rawSessions)),
	}

	for authToken, sess := range rawSessions {
		sessions.AuthTokens = append(sessions.AuthTokens, authToken)
		sessions.Sessions = append(sessions.Sessions, sess)
	}
	return sessions
}

func (sessions *sortedSessions) Sort() {
	sort.Sort(sessions)
}

func (sessions *sortedSessions) Len() int {
	return len(sessions.Sessions)
}

func (sessions *sortedSessions) Less(i, j int) bool {
	return sessions.Sessions[i] < sessions.Sessions[j]
}

func (sessions *sortedSessions) Swap(i, j int) {
	sessions.Sessions[i], sessions.Sessions[j] =
		sessions.Sessions[j], sessions.Sessions[i]

	sessions.AuthTokens[i], sessions.AuthTokens[j] =
		sessions.AuthTokens[j], sessions.AuthTokens[i]
}
