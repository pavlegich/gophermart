package handlers

import "net/http"

// HandleMain обрабатывает запрос главной страницы
func (c *Controller) HandleMain(w http.ResponseWriter, r *http.Request) {
	// ctx := r.Context()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello, user!"))
}
