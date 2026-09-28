package uss

import "net/http"

func (*Handler) MakeUSSReport(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusCreated)
}
