// Copyright 2012, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package streamlog provides a non-blocking message broadcaster.
package streamlog

import (
	"io"
	"net/http"
	"net/url"
	"sync"

	log "github.com/golang/glog"
	"gopkg.in/sqle/vitess-go.v1/acl"
	"gopkg.in/sqle/vitess-go.v1/stats"
)

var (
	sendCount         = stats.NewCounters("StreamlogSend")
	deliveredCount    = stats.NewMultiCounters("StreamlogDelivered", []string{"Log", "Subscriber"})
	deliveryDropCount = stats.NewMultiCounters("StreamlogDeliveryDroppedMessages", []string{"Log", "Subscriber"})
)

// StreamLogger is a non-blocking broadcaster of messages.
// Subscribers can use channels or HTTP.
type StreamLogger struct {
	name       string
	size       int
	mu         sync.Mutex
	subscribed map[chan interface{}]string
}

// New returns a new StreamLogger that can stream events to subscribers.
// The size parameter defines the channel size for the subscribers.
func New(name string, size int) *StreamLogger {
	return &StreamLogger{
		name:       name,
		size:       size,
		subscribed: make(map[chan interface{}]string),
	}
}

// Send sends message to all the writers subscribed to logger. Calling
// Send does not block.
func (logger *StreamLogger) Send(message interface{}) {
	logger.mu.Lock()
	defer logger.mu.Unlock()

	for ch, name := range logger.subscribed {
		select {
		case ch <- message:
			deliveredCount.Add([]string{logger.name, name}, 1)
		default:
			deliveryDropCount.Add([]string{logger.name, name}, 1)
		}
	}
	sendCount.Add(logger.name, 1)
}

// Subscribe returns a channel which can be used to listen
// for messages.
func (logger *StreamLogger) Subscribe(name string) chan interface{} {
	logger.mu.Lock()
	defer logger.mu.Unlock()

	ch := make(chan interface{}, logger.size)
	logger.subscribed[ch] = name
	return ch
}

// Unsubscribe removes the channel from the subscription.
func (logger *StreamLogger) Unsubscribe(ch chan interface{}) {
	logger.mu.Lock()
	defer logger.mu.Unlock()

	delete(logger.subscribed, ch)
}

// Name returns the name of StreamLogger.
func (logger *StreamLogger) Name() string {
	return logger.name
}

// ServeLogs registers the URL on which messages will be broadcast.
// It is safe to register multiple URLs for the same StreamLogger.
func (logger *StreamLogger) ServeLogs(url string, messageFmt func(url.Values, interface{}) string) {
	http.HandleFunc(url, func(w http.ResponseWriter, r *http.Request) {
		if err := acl.CheckAccessHTTP(r, acl.DEBUGGING); err != nil {
			acl.SendError(w, err)
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		ch := logger.Subscribe("ServeLogs")
		defer logger.Unsubscribe(ch)

		// Notify client that we're set up. Helpful to distinguish low-traffic streams from connection issues.
		w.WriteHeader(http.StatusOK)
		w.(http.Flusher).Flush()

		for message := range ch {
			if _, err := io.WriteString(w, messageFmt(r.Form, message)); err != nil {
				return
			}
			w.(http.Flusher).Flush()
		}
	})
	log.Infof("Streaming logs from %s at %v.", logger.Name(), url)
}
