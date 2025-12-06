package handlers

import (
	"fmt"
	"io"
	"net/http"

	"github.com/adfer-dev/analock-api/constants"
	"github.com/adfer-dev/analock-api/services"
	"github.com/adfer-dev/analock-api/utils"
	"github.com/gorilla/mux"
)

var internetArchiveService services.InternetArchiveService = &services.InternetArchiveServiceImpl{}

func InitInternetArchiveRoutes(router *mux.Router) {
	router.HandleFunc("/api/v1/internetArchive/books/search", utils.ParseToHandlerFunc(handleSearchInternetArchiveBooks)).Methods("GET")
	router.HandleFunc("/api/v1/internetArchive/books/{bookId}/metadata", utils.ParseToHandlerFunc(handleGetInternetArchiveBookMetadata)).Methods("GET")
	router.HandleFunc("/api/v1/internetArchive/books/{bookId}/download", utils.ParseToHandlerFunc(handleBookDownload)).Methods("GET")
}

// @Summary		Gets all Internet Archive books that match given params
// @Description	Gets all Internet Archive books that match given params
// @Tags			internet archive
// @Produce		json
// @Param			collection	query		string	true	"The collection"
// @Param			language		query		string	true	"The language"
// @Param			subject		query		string	true	"The subject"
// @Param			rows		query		int	true	"Row limit"
// @Success		200			{object}		models.InternetArchiveSearchResponse
// @Failure		400			{object}	models.HttpError
// @Failure		500			{object}	models.HttpError
// @Security		BearerAuth
// @Router			/internetArchive/books/search [get]
func handleSearchInternetArchiveBooks(res http.ResponseWriter, req *http.Request) error {
	collection := req.URL.Query().Get("collection")
	language := req.URL.Query().Get("language")
	subject := req.URL.Query().Get("subject")
	rows := req.URL.Query().Get("rows")

	if len(collection) == 0 || len(language) == 0 || len(subject) == 0 || len(rows) == 0 {
		return utils.WriteError(
			res,
			400,
			constants.ErrorRequiredParams,
		)
	}

	books, err := services.GetCacheServiceInstance().CacheResource(
		func() (interface{}, error) {
			return internetArchiveService.SearchBooks(collection, language, subject, rows)
		},
		constants.InternetArchiveBookSearchCacheResource,
		fmt.Sprintf("collection%s-language%s-subject%s-rows%s", collection, language, subject, rows),
	)

	if err != nil {
		utils.GetCustomLogger().Errorf(
			"Search book request failed: %s\n",
			err.Error(),
		)
		return utils.WriteError(
			res,
			500,
			"could not retrieve internet archive books.",
		)
	}

	return utils.WriteJSON(res, 200, &books)
}

// @Summary		Get IA book metadata
// @Description	Gets the metadata of the book that matches given identifier
// @Tags			internet archive
// @Produce			json
// @Param			bookId			path		string	true	"The IA book's identifier"
// @Success		200			{object}		models.InternetArchiveMetadataResponse
// @Failure		400			{object}	models.HttpError
// @Failure		500			{object}	models.HttpError
// @Security		BearerAuth
// @Router			/internetArchive/books/{bookId}/metadata [get]
func handleGetInternetArchiveBookMetadata(res http.ResponseWriter, req *http.Request) error {
	bookId, exists := mux.Vars(req)["bookId"]

	if !exists {
		return utils.WriteError(
			res,
			400,
			constants.ErrorRequiredParams,
		)
	}

	metadata, err := services.GetCacheServiceInstance().CacheResource(
		func() (interface{}, error) {
			return internetArchiveService.GetBookMetadata(bookId)
		},
		constants.InternetArchiveBookMetadataCacheResource,
		fmt.Sprintf("book-%s", bookId),
	)

	if err != nil {
		utils.GetCustomLogger().Errorf(
			"Metadata book request failed: %s\n",
			err.Error(),
		)
		return utils.WriteError(
			res,
			500,
			"could not retrieve internet archive book metadata.",
		)
	}

	return utils.WriteJSON(res, 200, metadata)
}

// @Summary		Downloads given book
// @Description	Downloads Internet Archive book with given identifier and file name.
// @Tags			internet archive
// @Produce			application/epub+zip
// @Param			bookId			path		string	true	"The IA book's identifier."
// @Param			file	query		string	true	"The name of the file to be downloaded from IA API."
// @Success		200			{file}		"Returns book's EPUB file"
// @Failure		400			{object}	models.HttpError
// @Failure		500			{object}	models.HttpError
// @Security		BearerAuth
// @Router			/internetArchive/books/{bookId}/download [get]
func handleBookDownload(res http.ResponseWriter, req *http.Request) error {
	bookId, exists := mux.Vars(req)["bookId"]
	file := req.URL.Query().Get("file")

	if !exists || len(file) == 0 {
		return utils.WriteError(
			res,
			400,
			constants.ErrorRequiredParams,
		)
	}

	response, downloadErr := internetArchiveService.DownloadBook(bookId, file)
	if downloadErr != nil {
		utils.GetCustomLogger().Errorf(
			"Download book request failed: %s\n",
			downloadErr.Error(),
		)
		return utils.WriteError(
			res,
			500,
			"could not download book.",
		)
	}

	defer response.Body.Close()

	res.Header().Set("Content-Type", "application/epub+zip")
	res.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.epub\"", bookId))
	res.Header().Set("Content-Length", fmt.Sprintf("%d", response.ContentLength))
	_, writeErr := io.Copy(res, response.Body)

	return writeErr
}
