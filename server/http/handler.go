package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/hellobchain/memmq/broker"
	"github.com/hellobchain/memmq/core/log"
	"github.com/hellobchain/wswlog/wlogging"
)

var logger = wlogging.MustGetFileLoggerWithoutName(log.LogConfig)
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type codeMsg struct {
	Code int    `json:"code"`
	Msg  string `json:"msg,omitempty"`
}

func newCodeMsg(code int, msg string) string {
	cm := &codeMsg{
		Code: code,
		Msg:  msg,
	}
	ret, _ := json.Marshal(cm)
	return string(ret)
}

func pub(w http.ResponseWriter, r *http.Request) {
	topic := r.URL.Query().Get("topic")
	logger.Infof("pub topic: %s", topic)
	if websocket.IsWebSocketUpgrade(r) {
		conn, err := upgrader.Upgrade(w, r, w.Header())
		if err != nil {
			logger.Errorf("upgrade error: %v", err)
			return
		}
		for {
			messageType, b, err := conn.ReadMessage()
			if messageType == -1 {
				return
			}
			if err != nil {
				logger.Warnf("read message error: %v", err)
				continue
			}
			logger.Infof("pub topic: %s, payload: %s", topic, string(b))
			broker.Publish(topic, b)
		}
	} else {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Errorf("pub error: %v", err)
			http.Error(w, newCodeMsg(200, err.Error()), http.StatusOK)
			return
		}
		r.Body.Close()
		logger.Infof("pub topic: %s, payload: %s", topic, string(b))
		if err := broker.Publish(topic, b); err != nil {
			logger.Errorf("pub error: %v", err)
			http.Error(w, newCodeMsg(200, err.Error()), http.StatusOK)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(newCodeMsg(200, "请求成功")))
	}
}

func sub(w http.ResponseWriter, r *http.Request) {
	var wr writer

	if websocket.IsWebSocketUpgrade(r) {
		conn, err := upgrader.Upgrade(w, r, w.Header())
		if err != nil {
			logger.Errorf("upgrade error: %v", err)
			return
		}
		// Drain the websocket so that we handle pings and connection close
		go func(c *websocket.Conn) {
			for {
				if _, _, err := c.NextReader(); err != nil {
					c.Close()
					logger.Errorf("websocket error: %v", err)
					break
				}
			}
		}(conn)
		wr = &wsWriter{conn}
	} else {
		wr = &httpWriter{w}
	}

	topic := r.URL.Query().Get("topic")
	logger.Info("Subscribing to topic: %s", topic)
	ch, err := broker.Subscribe(topic)
	if err != nil {
		logger.Errorf("Could not retrieve events: %v", err)
		http.Error(w, fmt.Sprintf("Could not retrieve events: %v", err), http.StatusInternalServerError)
		return
	}
	defer broker.Unsubscribe(topic, ch)

	for {
		select {
		case e := <-ch:
			logger.Info("Sending event: %s", string(e))
			if err = wr.Write(e); err != nil {
				logger.Errorf("Could not write event: %v", err)
				return
			}
		}
	}
}
