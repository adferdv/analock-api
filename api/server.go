package api

import (
	"fmt"
	"net/http"

	"os"

	"github.com/adfer-dev/analock-api/constants"
	"github.com/adfer-dev/analock-api/docs"
	"github.com/adfer-dev/analock-api/handlers"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	httpSwagger "github.com/swaggo/http-swagger"
)

type APIServer struct {
	Port   int
	router *mux.Router
}

func (server *APIServer) Run() error {
	server.router = mux.NewRouter()

	// Swagger documentation

	// set swagger host from environment
	environment := os.Getenv("API_ENVIRONMENT")
	if environment != "local" {
		docs.SwaggerInfo.Host = os.Getenv("API_PROD_URL_HOST")
	} else {
		docs.SwaggerInfo.Host = fmt.Sprintf("%s:%d", "localhost", server.Port)
	}

	server.router.PathPrefix(constants.ApiV1UrlRoot + "/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL(constants.ApiV1UrlRoot+"/swagger/doc.json"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	))

	// CORS config
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		MaxAge:           86400,
		Debug:            false,
	}).Handler(server.router)

	// Middlewares
	server.router.Use(AuthMiddleware, ValidatePathParams, UserOwnershipMiddleware)

	server.initRoutes()

	return http.ListenAndServe(fmt.Sprintf(":%d", server.Port), corsHandler)
}

func (server *APIServer) initRoutes() {
	handlers.InitUserRoutes(server.router)
	handlers.InitAuthRoutes(server.router)
	handlers.InitDiaryEntryRoutes(server.router)
	handlers.InitActivityRegistrationRoutes(server.router)
}
