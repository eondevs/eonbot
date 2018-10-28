package control

import (
	"github.com/sirupsen/logrus"
	"sync"
)

type Controller interface {
	// state control
	Stater

	// workflow control
	Starter
	Stopper

	Restart(start StartInfo, stop StopInfo, cause cause)
}

type Control struct {
	state struct {
		mu  sync.RWMutex
		val StateInfoer
		cb  func(s StateInfoer)
	}
	start struct {
		mu sync.Mutex
		ch chan StartInfo

		pen pendingAction
	}

	stop struct {
		mu sync.Mutex
		ch chan StopInfo

		pen pendingAction
	}
}

func New(cb func(s StateInfoer)) *Control {
	ctrl := &Control{}
	ctrl.state.val = newStoppedState(CauseInit)
	ctrl.state.cb = cb
	ctrl.start.ch = make(chan StartInfo)
	ctrl.stop.ch = make(chan StopInfo)
	return ctrl
}

func (c *Control) State() (info StateInfoer) {
	c.state.mu.RLock()
	info = c.state.val
	c.state.mu.RUnlock()
	return info
}

func (c *Control) updateState(s StateInfoer) {
	c.state.mu.Lock()
	c.state.val = s
	c.state.mu.Unlock()
	go func() {
		if c.state.cb != nil {
			c.state.cb(c.State())
		}
	}()
}

func (c *Control) WaitStart() <-chan StartInfo {
	return c.start.ch
}

func (c *Control) Start(info StartInfo, cause cause) startResult {
	return c.startExec(info, cause, nil)
}

func (c *Control) startExec(info StartInfo, cause cause, waitCh chan<- bool) startResult {
	waitOver := func() {
		if waitCh != nil {
			waitCh <- true
		}
	}

	c.start.mu.Lock() // we need to ensure that only one call of this function executes at a time
	if c.State().IsRunning() {
		c.start.mu.Unlock()
		waitOver()
		return StartAlreadyActive
	}

	if c.start.pen.isPending() {
		c.start.mu.Unlock()
		waitOver()
		return StartAlreadyInitialized
	}

	logrus.StandardLogger().Debug("start command execution initialized")

	go func() {
		info.Init()
		c.start.pen.activate()
		c.start.mu.Unlock()
		logrus.StandardLogger().Debug("start command sent")
		c.start.ch <- info
		if <-info.Commons.Confirm {
			c.updateState(newStartedState(cause))
		}
		c.start.pen.deactivate()
		waitOver()
	}()

	return StartActivated
}

func (c *Control) IsStartPending() bool {
	return c.start.pen.isPending()
}

func (c *Control) WaitStop() <-chan StopInfo {
	return c.stop.ch
}

func (c *Control) Stop(info StopInfo, cause cause) stopResult {
	return c.stopExec(info, cause, nil)
}

func (c *Control) stopExec(info StopInfo, cause cause, waitCh chan<- bool) stopResult {
	waitOver := func() {
		if waitCh != nil {
			waitCh <- true
		}
	}

	c.stop.mu.Lock() // we need to ensure that only one call of this function executes at a time
	if !c.State().IsRunning() {
		c.stop.mu.Unlock()
		waitOver()
		return StopAlreadyActive
	}

	if c.stop.pen.isPending() {
		c.stop.mu.Unlock()
		waitOver()
		return StopAlreadyInitialized
	}

	logrus.StandardLogger().Debug("stop command execution initialized")

	go func() {
		info.Init()
		c.stop.pen.activate()
		c.stop.mu.Unlock()
		logrus.StandardLogger().Debug("stop command sent")
		c.stop.ch <- info
		if <-info.Commons.Confirm {
			c.updateState(newStoppedState(cause))
		}
		c.stop.pen.deactivate()
		waitOver()
	}()

	return StopActivated
}

func (c *Control) IsStopPending() bool {
	return c.stop.pen.isPending()
}

func (c *Control) Restart(start StartInfo, stop StopInfo, cause cause) {
	go func() {
		waitCh := make(chan bool)
		go c.stopExec(stop, cause, waitCh)
		<-waitCh
		go c.startExec(start, cause, waitCh)
		<-waitCh
	}()
}

type pendingAction struct {
	sync.RWMutex
	pending bool
}

func (p *pendingAction) isPending() (res bool) {
	p.RLock()
	res = p.pending
	p.RUnlock()
	return res
}

func (p *pendingAction) activate() {
	p.Lock()
	p.pending = true
	p.Unlock()
}

func (p *pendingAction) deactivate() {
	p.Lock()
	p.pending = false
	p.Unlock()
}
