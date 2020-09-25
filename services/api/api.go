package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/romeufcrosa/where-to-eat/services/v1"
)

// Router ...
func Router() http.Handler {
	params := v1.Params{}

	router := httprouter.New()

	router.POST("/api/v1/restaurants", v1.FindWhereToEat(params))
	router.ServeFiles("/*filepath", http.Dir("./web"))

	return router
}
