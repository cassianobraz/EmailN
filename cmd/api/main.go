package main

import (
	"net/http"

	"emailn/cmd/initializers"
	"emailn/internal/domain/campaign"
	"emailn/internal/endpoints"
	"emailn/internal/infrastructure/database"
	"emailn/internal/infrastructure/mail"

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
		SendMail:   mail.SendMail,
	}
	handler := endpoints.Handler{
		CampaignService: &campaignService,
	}

	r.Route("/campaigns", func(r chi.Router) {
		r.Use(endpoints.Auth)
		r.Post("/", endpoints.HandlerError(handler.CampaignPost))
		r.Get("/{id}", endpoints.HandlerError(handler.CampaignGetById))
		r.Delete("/delete/{id}", endpoints.HandlerError(handler.CampaignDelete))
		r.Patch("/start/{id}", endpoints.HandlerError(handler.CampaignStart))
	})

	http.ListenAndServe(":3000", r)
}
