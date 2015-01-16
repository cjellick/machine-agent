package locks

type Unlocker interface {
	Unlock()
}

func Lock(key string) Unlocker {
	lockRequest := newLockRequest(key, LOCK)
	lockRequests <- lockRequest
	success := <-lockRequest.success
	if success {
		return newUnlocker(key)
	}

	return nil
}

type operation int

const (
	LOCK operation = iota
	UNLOCK
)

type lockRequest struct {
	key     string
	op      operation
	success chan bool
}

var lockRequests chan *lockRequest

func locker() {
	// note: bool value is meaningless. This is a set
	lockedItems := make(map[string]bool)
	for {
		lockReq := <-lockRequests
		switch lockReq.op {
		case LOCK:
			if _, locked := lockedItems[lockReq.key]; locked {
				// already locked
				lockReq.success <- false
			} else {
				lockedItems[lockReq.key] = true
				lockReq.success <- true
			}
		case UNLOCK:
			delete(lockedItems, lockReq.key)
			lockReq.success <- true
		}
	}
}

func init() {
	lockRequests = make(chan *lockRequest)
	go locker()

}

func newLockRequest(key string, op operation) *lockRequest {
	return &lockRequest{
		key:     key,
		op:      op,
		success: make(chan bool, 1),
	}
}

type unlockerImpl struct {
	key string
}

func (u *unlockerImpl) Unlock() {
	lockRequests <- newLockRequest(u.key, UNLOCK)
}

func newUnlocker(key string) *unlockerImpl {
	return &unlockerImpl{key: key}
}
