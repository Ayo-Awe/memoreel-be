package public

import (
	"net/http"

	"github.com/ayo-awe/memoreel-be/api/types"
	"github.com/go-chi/chi/v5"
)

type PublicHandler struct {
	Router http.Handler
	Opts   types.APIOptions
}

func (p *PublicHandler) BuildRoutes() http.Handler {
	router := chi.NewRouter()
	v1Router := chi.NewRouter()

	v1Router.Route("/me", func(meRouter chi.Router) {
		meRouter.Get("/", func(w http.ResponseWriter, r *http.Request) {})
		meRouter.Patch("/", func(w http.ResponseWriter, r *http.Request) {})
	})

	v1Router.Route("/reels", func(reelRouter chi.Router) {
		reelRouter.Get("/", func(w http.ResponseWriter, r *http.Request) {})
		reelRouter.Post("/", func(w http.ResponseWriter, r *http.Request) {})
		reelRouter.Post("/confirm", func(w http.ResponseWriter, r *http.Request) {})
		reelRouter.Route("/{reelID}", func(reelSubRouter chi.Router) {
			reelSubRouter.Get("/", func(w http.ResponseWriter, r *http.Request) {})
			reelSubRouter.Put("/", func(w http.ResponseWriter, r *http.Request) {})
			reelSubRouter.Delete("/", func(w http.ResponseWriter, r *http.Request) {})
			reelSubRouter.Route("/recipients", func(recipientRouter chi.Router) {
				recipientRouter.Post("/", func(w http.ResponseWriter, r *http.Request) {})
				recipientRouter.Delete("/{recipientID}", func(w http.ResponseWriter, r *http.Request) {})
			})
		})

	})

	v1Router.Route("/videos", func(videoRouter chi.Router) {
		videoRouter.Post("/", func(w http.ResponseWriter, r *http.Request) {})
	})

	router.Mount("/v1", v1Router)

	p.Router = router
	return router
}
