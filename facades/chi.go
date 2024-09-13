package facades

import (
	"log"

	"github.com/goravel/framework/contracts/route"

	"github.com/goravel/chi"
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
