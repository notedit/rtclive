package store

import (
	"sync"

	"github.com/notedit/rtclive/router"
	"github.com/olahol/melody"
)

type Routers struct {
	routers map[string]*router.MediaRouter
	sync.Mutex
}

var routers *Routers = &Routers{
	routers: map[string]*router.MediaRouter{},
}

func GetRouter(routerId string) *router.MediaRouter {
	routers.Lock()
	defer routers.Unlock()
	return routers.routers[routerId]
}

func AddRouter(router *router.MediaRouter) {
	routers.Lock()
	defer routers.Unlock()
	routers.routers[router.GetID()] = router
}

func RemoveRouter(router *router.MediaRouter) {
	routers.Lock()
	defer routers.Unlock()
	delete(routers.routers, router.GetID())
}

type SessionInfo struct {
	StreamID     string
	SubscriberID string
}

type Sessions struct {
	sessions map[*melody.Session]*SessionInfo
	sync.Mutex
}

var sessions *Sessions = &Sessions{
	sessions: map[*melody.Session]*SessionInfo{},
}

func AddSession(session *melody.Session) {
	sessions.Lock()
	defer sessions.Unlock()
	sessions.sessions[session] = new(SessionInfo)
}

func GetSession(session *melody.Session) *SessionInfo {
	sessions.Lock()
	defer sessions.Unlock()
	return sessions.sessions[session]
}

func RemoveSession(session *melody.Session) {
	sessions.Lock()
	defer sessions.Unlock()
	delete(sessions.sessions, session)
}
