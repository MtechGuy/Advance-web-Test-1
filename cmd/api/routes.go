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

	//Product part
	router.HandlerFunc(http.MethodGet, "/healthcheck", a.healthcheckHandler)
	router.HandlerFunc(http.MethodGet, "/product", a.listProductHandler)
	router.HandlerFunc(http.MethodPost, "/product", a.createProductHandler)
	router.HandlerFunc(http.MethodGet, "/product/:id", a.displayProductHandler)
	router.HandlerFunc(http.MethodPatch, "/product/:id", a.updateProductHandler)
	router.HandlerFunc(http.MethodDelete, "/product/:id", a.deleteProductHandler)

	// //Review part
	router.HandlerFunc(http.MethodGet, "/review", a.listReviewHandler)
	router.HandlerFunc(http.MethodPost, "/review", a.createReviewHandler)
	router.HandlerFunc(http.MethodGet, "/review/:id", a.displayReviewHandler)
	router.HandlerFunc(http.MethodPatch, "/review/:id", a.updateReviewHandler)
	router.HandlerFunc(http.MethodDelete, "/review/:id", a.deleteReviewHandler)

	router.HandlerFunc(http.MethodGet, "/product_review/:id", a.listProductReviewHandler)

	return a.recoverPanic(router)

}
