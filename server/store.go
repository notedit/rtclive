package server

import (
	"sync"

	"github.com/notedit/rtclive/router"
)

type Routers struct {
	routers map[string]*router.MediaRouter
	sync.Mutex
}

var routers *Routers = &Routers{
	routers: map[string]*router.MediaRouter{},
}

func (r *Routers) GetRouter(routerId string) *router.MediaRouter {
	routers.Lock()
	defer routers.Unlock()
	return routers.routers[routerId]
}

func (r *Routers) AddRouter(router *router.MediaRouter) {
	routers.Lock()
	defer routers.Unlock()
	routers.routers[router.GetID()] = router
}

func (r *Routers) RemoveRouter(router *router.MediaRouter) {
	routers.Lock()
	defer routers.Unlock()
	delete(routers.routers, router.GetID())
}
