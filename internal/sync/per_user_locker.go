package sync

import (
	"fmt"
	"sync"
	"time"
)

type PerUserLocker interface {
	TryLock(userID int64, request interface{}) bool
	Lock(userID int64, request interface{})
	Unlock(userID int64)
	Len() int
	FrozenRequests(threshold time.Duration) map[int64]interface{}
}

type user struct {
	mutex       *sync.Mutex
	lastRequest interface{}
	requestsInQ int
	startedAt   time.Time
}

func NewPerUserLocker() PerUserLocker {
	return &locker{
		users: make(map[int64]*user),
	}
}

type locker struct {
	users   map[int64]*user
	mapLock sync.Mutex
}

func (l *locker) TryLock(userID int64, request interface{}) bool {
	l.mapLock.Lock()
	defer l.mapLock.Unlock()
	u := l.getUser(userID)
	if u.mutex.TryLock() {
		u.requestsInQ++
		u.lastRequest = request
		u.startedAt = time.Now()
		return true
	}
	return false
}

func (l *locker) Lock(userID int64, request interface{}) {
	l.mapLock.Lock()
	u := l.getUser(userID)
	u.requestsInQ++
	l.mapLock.Unlock()
	u.mutex.Lock()
	u.lastRequest = request
	u.startedAt = time.Now()
}

func (l *locker) Unlock(userID int64) {
	l.mapLock.Lock()
	defer l.mapLock.Unlock()
	user, ok := l.users[userID]
	if !ok {
		panic(fmt.Sprintf("Unlocking non-existing mutex, userID=%v", userID))
	}
	user.mutex.Unlock()
	user.requestsInQ--
	if user.requestsInQ == 0 {
		delete(l.users, userID)
	}
}

func (l *locker) Len() int {
	l.mapLock.Lock()
	defer l.mapLock.Unlock()
	return len(l.users)
}

func (l *locker) FrozenRequests(threshold time.Duration) map[int64]interface{} {
	l.mapLock.Lock()
	defer l.mapLock.Unlock()
	res := make(map[int64]interface{})
	for userID, u := range l.users {
		if time.Since(u.startedAt) > threshold {
			res[userID] = u.lastRequest
		}
	}
	return res
}

// getUser returns user by userID, if user does not exist, it creates new one
// the function is not thread safe, it should be called from thread safe function
func (l *locker) getUser(userID int64) *user {
	u, ok := l.users[userID]
	if !ok {
		u = &user{
			mutex: &sync.Mutex{},
		}
		l.users[userID] = u
	}
	return u
}
