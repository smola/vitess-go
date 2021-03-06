// Copyright 2015, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"gopkg.in/sqle/vitess-go.v1/vt/servenv"
	"gopkg.in/sqle/vitess-go.v1/vt/vtctl/grpcvtctlserver"
)

func init() {
	servenv.RegisterGRPCFlags()
	servenv.OnRun(func() {
		if servenv.GRPCCheckServiceMap("vtctl") {
			grpcvtctlserver.StartServer(servenv.GRPCServer, ts)
		}
	})
}
