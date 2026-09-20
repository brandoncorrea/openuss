package router

import "net/http"

type Registrar interface {
	RegisterRoutes(*http.ServeMux)
}

// New returns a handler covering every route its registrars serve.
func New(registrars ...Registrar) http.Handler {
	mux := http.NewServeMux()
	for _, registrar := range registrars {
		registrar.RegisterRoutes(mux)
	}
	return mux
}
