// Copyright 2014, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

// Imports and register the 'etcd2' topo.Server and its Explorer.

import (
	"gopkg.in/sqle/vitess-go.v1/vt/servenv"
	"gopkg.in/sqle/vitess-go.v1/vt/topo/etcd2topo"
	"gopkg.in/sqle/vitess-go.v1/vt/vtctld"
)

func init() {
	// Wait until flags are parsed, so we can check which topo server is in use.
	servenv.OnRun(func() {
		if s, ok := ts.Impl.(*etcd2topo.Server); ok {
			vtctld.HandleExplorer("etcd2", vtctld.NewBackendExplorer(s))
		}
	})
}
