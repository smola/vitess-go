// Copyright 2015, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schemamanager

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	log "github.com/golang/glog"
	"golang.org/x/net/context"
)

// UIController handles schema events.
type UIController struct {
	sqls     []string
	keyspace string
	writer   http.ResponseWriter
}

// NewUIController creates a UIController instance
func NewUIController(
	sqlStr string, keyspace string, writer http.ResponseWriter) *UIController {
	controller := &UIController{
		sqls:     make([]string, 0, 32),
		keyspace: keyspace,
		writer:   writer,
	}
	for _, sql := range strings.Split(sqlStr, ";") {
		s := strings.TrimSpace(sql)
		if s != "" {
			controller.sqls = append(controller.sqls, s)
		}
	}

	return controller
}

// Open is a no-op.
func (controller *UIController) Open(ctx context.Context) error {
	return nil
}

// Read reads schema changes
func (controller *UIController) Read(ctx context.Context) ([]string, error) {
	return controller.sqls, nil
}

// Close is a no-op.
func (controller *UIController) Close() {
}

// Keyspace returns keyspace to apply schema.
func (controller *UIController) Keyspace() string {
	return controller.keyspace
}

// OnReadSuccess is no-op
func (controller *UIController) OnReadSuccess(ctx context.Context) error {
	controller.writer.Write(
		[]byte(fmt.Sprintf("OnReadSuccess, sqls: %v\n", controller.sqls)))
	return nil
}

// OnReadFail is no-op
func (controller *UIController) OnReadFail(ctx context.Context, err error) error {
	controller.writer.Write(
		[]byte(fmt.Sprintf("OnReadFail, error: %v\n", err)))
	return err
}

// OnValidationSuccess is no-op
func (controller *UIController) OnValidationSuccess(ctx context.Context) error {
	controller.writer.Write(
		[]byte(fmt.Sprintf("OnValidationSuccess, sqls: %v\n", controller.sqls)))
	return nil
}

// OnValidationFail is no-op
func (controller *UIController) OnValidationFail(ctx context.Context, err error) error {
	controller.writer.Write(
		[]byte(fmt.Sprintf("OnValidationFail, error: %v\n", err)))
	return err
}

// OnExecutorComplete is no-op
func (controller *UIController) OnExecutorComplete(ctx context.Context, result *ExecuteResult) error {
	data, err := json.Marshal(result)
	if err != nil {
		log.Errorf("Failed to serialize ExecuteResult: %v", err)
		return err
	}
	controller.writer.Write([]byte(fmt.Sprintf("Executor succeeds: %s", string(data))))
	return nil
}

var _ Controller = (*UIController)(nil)
