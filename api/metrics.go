package api

import "net/http"


func MetricsAPI(w http.ResponseWriter, r *http.Request) {
	H.ServeHTTP(w, r)
}
