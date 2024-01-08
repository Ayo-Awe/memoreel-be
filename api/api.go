package api

import (
	"net/http"

	"github.com/ayo-awe/memoreel-be/api/public"
	"github.com/ayo-awe/memoreel-be/api/types"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type applicationHandler struct {
	Router http.Handler
	Opts   types.APIOptions
}

func NewApplicationHandler(opts types.APIOptions) (*applicationHandler, error) {
	return &applicationHandler{Opts: opts}, nil
}

func (a *applicationHandler) BuildRoutes() *chi.Mux {
	router := chi.NewMux()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)

	publicHandler := public.PublicHandler{Opts: a.Opts}
	router.Mount("/api", publicHandler.BuildRoutes())

	a.Router = router

	return router
}
