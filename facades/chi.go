package facades

import (
	"log"

	"github.com/go-chi/chi/v5"
	"github.com/goravel/framework/contracts/route"
)

func Route(driver string) route.Route {
	instance, err := chi.App.MakeWith(chi.RouteBinding, map[string]any{
		"driver": driver,
	})
	if err != nil {
		log.Fatalln(err)
		return nil
	}

	return instance.(*chi.Route)
}
