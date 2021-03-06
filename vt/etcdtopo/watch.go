package etcdtopo

import (
	"fmt"
	"sync"

	"github.com/coreos/go-etcd/etcd"
	"golang.org/x/net/context"

	"gopkg.in/sqle/vitess-go.v1/vt/topo"
)

func newWatchData(valueType dataType, node *etcd.Node) *topo.WatchData {
	bytes, err := rawDataFromNodeValue(valueType, node.Value)
	if err != nil {
		return &topo.WatchData{Err: err}
	}

	return &topo.WatchData{
		Contents: bytes,
		Version:  EtcdVersion(node.ModifiedIndex),
	}
}

// Watch is part of the topo.Backend interface
func (s *Server) Watch(ctx context.Context, cellName, filePath string) (*topo.WatchData, <-chan *topo.WatchData, topo.CancelFunc) {
	cell, err := s.getCell(cellName)
	if err != nil {
		return &topo.WatchData{Err: fmt.Errorf("Watch cannot get cell: %v", err)}, nil, nil
	}

	// Special paths where we need to be backward compatible.
	var valueType dataType
	valueType, filePath = oldTypeAndFilePath(filePath)

	// Get the initial version of the file
	initial, err := cell.Get(filePath, false /* sort */, false /* recursive */)
	if err != nil {
		// generic error
		return &topo.WatchData{Err: convertError(err)}, nil, nil
	}
	if initial.Node == nil {
		// node doesn't exist
		return &topo.WatchData{Err: topo.ErrNoNode}, nil, nil
	}
	wd := newWatchData(valueType, initial.Node)
	if wd.Err != nil {
		return wd, nil, nil
	}

	// mu protects the stop channel. We need to make sure the 'cancel'
	// func can be called multiple times, and that we don't close 'stop'
	// more than once.
	mu := sync.Mutex{}
	stop := make(chan bool)
	cancel := func() {
		mu.Lock()
		defer mu.Unlock()
		if stop != nil {
			close(stop)
			stop = nil
		}
	}

	notifications := make(chan *topo.WatchData, 10)

	// This watch go routine will stop if the 'stop' channel is closed.
	// Otherwise it will try to watch everything in a loop, and send events
	// to the 'watch' channel.
	// In any case, the Watch call will close the 'watch' channel.
	// Note we pass in the 'stop' channel as a parameter because
	// the go routine can take some time to start, and if someone
	// calls 'cancel' before the go routine starts, stop will be nil (note
	// this happens in the topo unit test, that cals cancel() right
	// after setting the watch).
	watchChannel := make(chan *etcd.Response)
	watchError := make(chan error)
	go func(stop chan bool) {
		versionToWatch := initial.Node.ModifiedIndex + 1
		_, err := cell.Client.Watch(filePath, versionToWatch, false /* recursive */, watchChannel, stop)
		// Watch will only return a non-nil error, otherwise
		// it keeps on watching. Send the error down.
		watchError <- err
		close(watchError)
	}(stop)

	// This go routine is the main event handling routine:
	// - it will stop if 'stop' is closed.
	// - if it receives a notification from the watch, it will forward it
	// to the notifications channel.
	go func() {
		defer close(notifications)

		for resp := range watchChannel {
			if resp.Action == "delete" || resp.Action == "compareAndDelete" {
				// Node doesn't exist any more, we can
				// stop watching. Swallow the watchError.
				mu.Lock()
				if stop == nil {
					// Watch was already interrupted
					mu.Unlock()
					return
				}
				close(stop)
				stop = nil
				mu.Unlock()
				<-watchError
				notifications <- &topo.WatchData{Err: topo.ErrNoNode}
				return
			}

			wd := newWatchData(valueType, resp.Node)
			notifications <- wd
			if wd.Err != nil {
				// Error packing / unpacking data,
				// stop the watch. Swallow the watchError.
				mu.Lock()
				if stop == nil {
					// Watch was already interrupted
					mu.Unlock()
					return
				}
				close(stop)
				stop = nil
				mu.Unlock()
				<-watchError
				notifications <- &topo.WatchData{Err: wd.Err}
				return
			}
		}

		// Watch terminated, because of an error. Recover the error,
		// and translate the interruption error.
		err := <-watchError
		if err == etcd.ErrWatchStoppedByUser {
			err = topo.ErrInterrupted
		}
		notifications <- &topo.WatchData{Err: err}
	}()

	return wd, notifications, cancel
}
