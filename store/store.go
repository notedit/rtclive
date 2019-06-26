package store

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
