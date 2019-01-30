package main

import "sync"


type Sessions struct {
	routers map[string]*MediaRouter
	sync.Mutex
}


var sessions *Sessions = new(Sessions)


func (s *Sessions) Get(routerId string) *MediaRouter {
	s.Lock()
	defer s.Unlock()
	return  s.routers[routerId]
}


func (s *Sessions) Add(router *MediaRouter) {
	s.Lock()
	defer s.Unlock()
	s.routers[router.routerID] = router
}


func (s *Sessions) Remove(router *MediaRouter)  {
	s.Lock()
	defer s.Unlock()
	delete(s.routers,router.routerID)
}

