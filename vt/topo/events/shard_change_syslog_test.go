package events

import (
	"log/syslog"
	"testing"

	topodatapb "gopkg.in/sqle/vitess-go.v1/vt/proto/topodata"
)

func TestShardChangeSyslog(t *testing.T) {
	wantSev, wantMsg := syslog.LOG_INFO, "keyspace-123/shard-123 [shard] status value: master_alias:<cell:\"test\" uid:123 > "
	sc := &ShardChange{
		KeyspaceName: "keyspace-123",
		ShardName:    "shard-123",
		Shard: &topodatapb.Shard{
			MasterAlias: &topodatapb.TabletAlias{
				Cell: "test",
				Uid:  123,
			},
		},
		Status: "status",
	}
	gotSev, gotMsg := sc.Syslog()

	if gotSev != wantSev {
		t.Errorf("wrong severity: got %v, want %v", gotSev, wantSev)
	}
	if gotMsg != wantMsg {
		t.Errorf("wrong message: got %v, want %v", gotMsg, wantMsg)
	}
}
