package i3qws

import (
	"container/list"
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/sirupsen/logrus"
	"go.i3wm.org/i3/v4"
)

type (
	// I3qws is a synchronized window list with additional exported methods
	I3qws struct {
		m          sync.RWMutex
		l          *list.List
		markFormat string
	}
)

// DoSpy starts spy for i3wm windows events
func DoSpy(ctx context.Context, markFormat string) *I3qws {
	i3qws := &I3qws{
		l:          list.New(),
		markFormat: markFormat,
	}
	recv := i3.Subscribe(i3.WindowEventType, i3.ShutdownEventType)

	go func() {
		<-ctx.Done()
		if e := recv.Close(); e != nil {
			logrus.WithError(e).Warn("Closing i3 communication channel with an error")
		}
	}()

	go func() {
		for recv.Next() {
			i := recv.Event()

			switch ev := i.(type) {
			case *i3.WindowEvent:
				i3qws.onWindowEvent(ev)
			case *i3.ShutdownEvent:
				i3qws.onShutdownEvent(ev)
			}

		}
	}()
	return i3qws
}

func (q *I3qws) onWindowEvent(ev *i3.WindowEvent) {
	q.m.Lock()
	le := logrus.WithFields(logrus.Fields{
		"id":   ev.Container.ID,
		"name": ev.Container.Name,
		"type": ev.Container.Type,
	})
	le.Debugf("Window change: %s", ev.Change)

	shouldRemark := false
	switch ev.Change {
	case "focus":
		found := false
		for c := q.l.Front(); c != nil; c = c.Next() {
			n := c.Value.(*i3.Node)
			if n.ID == ev.Container.ID {
				q.l.MoveToFront(c)
				found = true
				break
			}
		}
		if !found {
			q.l.PushFront(&ev.Container)
		}
		shouldRemark = true
	case "title", "mark":
		found := false
		for c := q.l.Front(); c != nil; c = c.Next() {
			n := c.Value.(*i3.Node)
			if n.ID == ev.Container.ID {
				c.Value = &ev.Container
				found = true
				break
			}
		}
		if !found {
			le.Debug("Container for update was not found")
		}
		shouldRemark = false
	case "close":
		for c := q.l.Front(); c != nil; c = c.Next() {
			n := c.Value.(*i3.Node)
			if n.ID == ev.Container.ID {
				q.l.Remove(c)
				break
			}
		}
		shouldRemark = true
	}
	q.m.Unlock()
	if len(q.markFormat) > 0 && shouldRemark {
		q.remark()
	}
}

func (q *I3qws) onShutdownEvent(ev *i3.ShutdownEvent) {
	q.m.Lock()
	logrus.Warnf("Shutdown change: %s", ev.Change)
	q.l.Init()
	q.m.Unlock()
}

func (q *I3qws) remark() {
	q.m.RLock()
	defer q.m.RUnlock()
	i := 0
	for c := q.l.Front(); c != nil; c = c.Next() {
		n := c.Value.(*i3.Node)
		mark := fmt.Sprintf(q.markFormat, i)
		q.runCommand("[con_id=" + strconv.FormatInt(int64(n.ID), 10) + "] mark --add " + mark) // nolint: errcheck
		i++
	}
}

// DumpList dumps windows list
func (q *I3qws) DumpList() []*i3.Node {
	nn := make([]*i3.Node, 0, q.l.Len())
	q.m.RLock()
	defer q.m.RUnlock()
	for c := q.l.Front(); c != nil; c = c.Next() {
		nn = append(nn, c.Value.(*i3.Node))
	}
	return nn
}

// Focus will bring to front window with number `num`.
// `num` starts from 0 — the first window, focused, 1 — the second one, last time focused and so on
// if `num` is negative, it should count from the end
func (q *I3qws) Focus(num int) (*i3.Node, error) {
	q.m.RLock()
	defer q.m.RUnlock()
	if num < 0 && q.l.Len() > 0 {
		num = q.l.Len() + num
		if num < 0 {
			num = 0
		}
	}
	i := 0
	for c := q.l.Front(); c != nil; c = c.Next() {
		if i == num {
			n := c.Value.(*i3.Node)
			_, err := q.runCommand("[con_id=" + strconv.FormatInt(int64(n.ID), 10) + "] focus")
			return n, err
		}
		i++
	}
	return nil, nil
}

func (q *I3qws) runCommand(command string) ([]i3.CommandResult, error) {
	res, err := i3.RunCommand(command)
	if err != nil {
		logrus.WithField("i3Command", command).WithError(err).Error("Error while running command on i3wm")
	} else {
		logrus.WithField("i3Command", command).WithField("res", res).Trace("Successfully ran command on i3wm")
	}
	return res, err
}
