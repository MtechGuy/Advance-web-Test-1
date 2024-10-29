// Filename: cmd/api/routes.go
package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (a *applicationDependencies) routes() http.Handler {

	router := httprouter.New()

	router.NotFound = http.HandlerFunc(a.notFoundResponse)

	router.MethodNotAllowed = http.HandlerFunc(a.methodNotAllowedResponse)
	//Product
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", a.healthcheckHandler)
	router.HandlerFunc(http.MethodGet, "/v1/comments", a.listProductHandler)
	router.HandlerFunc(http.MethodPost, "/v1/comments", a.createProductHandler)
	router.HandlerFunc(http.MethodGet, "/v1/comments/:id", a.displayProductHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/comments/:id", a.updateProductHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/comments/:id", a.deleteProductHandler)

	//Review part
	router.HandlerFunc(http.MethodGet, "/v1/comments", a.listReviewHandler)
	// router.HandlerFunc(http.MethodGet, "/v1/comments", a.listProductReviewHandler)
	router.HandlerFunc(http.MethodPost, "/v1/comments", a.createReviewHandler)
	router.HandlerFunc(http.MethodGet, "/v1/comments/:id", a.displayReviewHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/comments/:id", a.updateReviewHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/comments/:id", a.deleteReviewHandler)

	return a.recoverPanic(router)

}
