package txthrottler

import (
	"github.com/golang/mock/gomock"
	"gopkg.in/sqle/vitess-go.v1/vt/topo"
)

// NewMockServer wraps a MockImpl in a Server object.
// It returns the wrapped object so that the caller can set expectations on it.
func NewMockServer(ctrl *gomock.Controller) (topo.Server, *MockImpl) {
	mockImpl := NewMockImpl(ctrl)
	return topo.Server{Impl: mockImpl}, mockImpl
}
