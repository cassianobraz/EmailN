package main

import (
	"net/http"

	"emailn/cmd/initializers"
	"emailn/internal/domain/campaign"
	"emailn/internal/endpoints"
	"emailn/internal/infrastructure/database"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type myHandler struct{}

func (m myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("myHandler"))
}

func init() {
	initializers.LoadEnvVariables()
}

func main() {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.ClientIPFromRemoteAddr)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	db := database.NewDB()
	campaignService := campaign.ServiceImp{
		Repository: &database.CampaignRepository{Db: db},
	}
	handler := endpoints.Handler{
		CampaignService: &campaignService,
	}

	r.Route("/campaigns", func(r chi.Router) {
		r.Use(endpoints.Auth)
		r.Post("/", endpoints.HandlerError(handler.CampaignPost))
		r.Get("/{id}", endpoints.HandlerError(handler.CampaignGetById))
		r.Patch("/cancel/{id}", endpoints.HandlerError(handler.CampaignCancelPatch))
		r.Delete("/delete/{id}", endpoints.HandlerError(handler.CampaignDelete))
	})

	http.ListenAndServe(":3000", r)
}
