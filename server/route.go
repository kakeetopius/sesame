package server

import (
	"net/http"

	"github.com/gorilla/mux"
)

type route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type routes []route

func (s *Server) addRoutes() {
	routesForApp := routes{
		route{
			"Index Page",
			"GET",
			"/",
			s.Index,
		},
		route{
			"Add New IP",
			"POST",
			"/ip",
			s.IPAdd,
		},
		route{
			"Show all IPs present",
			"GET",
			"/ip",
			s.ShowAllIPs,
		},
		route{
			"Show if particular IP is present",
			"GET",
			"/ip/{ip}",
			s.IPShow,
		},
		route{
			"Delete IP",
			"DELETE",
			"/ip/{ip}",
			s.IPDelete,
		},
	}

	router := mux.NewRouter()
	subrouter := router.PathPrefix("/m{[0-9]+}").Subrouter().StrictSlash(true)
	subrouter.Use(loggingMiddleware, s.validateMiddleware)

	for _, route := range routesForApp {
		subrouter.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(route.HandlerFunc)
	}

	s.Router = router
}
