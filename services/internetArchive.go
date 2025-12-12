package services

import (
	"fmt"
	"net/http"
	"time"

	"github.com/adfer-dev/analock-api/models"
	"github.com/adfer-dev/analock-api/utils"
)

type InternetArchiveService interface {
	SearchBooks(collection string, language string, subject string, rows string) (*models.InternetArchiveSearchResponse, error)
	GetBookMetadata(bookId string) (*models.InternetArchiveMetadataResponse, error)
	DownloadBook(bookId string, fileName string) (*http.Response, error)
}

type InternetArchiveServiceImpl struct{}

// Performs an HTTP request to Internet Archive API to get books that match the given criteria.
//
// It returns the number of books given in the rows param.
func (iaService *InternetArchiveServiceImpl) SearchBooks(collection string, language string, subject string, rows string) (*models.InternetArchiveSearchResponse, error) {
	url := fmt.Sprintf(
		"https://archive.org/advancedsearch.php?q=collection:%s+AND+language:%s+AND+subject:%s+AND+mediatype:texts&fl=title,creator,identifier&sort[]=downloads+desc&sort[]=avg_rating+desc&rows=%s&page=1&output=json",
		collection,
		language,
		subject,
		rows,
	)

	res, err := PerformRequest[models.InternetArchiveSearchResponse](http.MethodGet, url, nil)

	if err != nil {
		return nil, err
	}

	return res, nil
}

// Performs an HTTP request to Internet Archive API to get the metadata of the book that matches given identifier.
func (iaService *InternetArchiveServiceImpl) GetBookMetadata(bookId string) (*models.InternetArchiveMetadataResponse, error) {
	url := fmt.Sprintf(
		"https://archive.org/metadata/%s", bookId)

	res, err := PerformRequest[models.InternetArchiveMetadataResponse](http.MethodGet, url, nil)

	if err != nil {
		return nil, err
	}

	return res, nil
}

// Performs an HTTP request to Internet Archive API to download a book's file.
// The file that is returned depends on the given book identifier and file name.
func (iaService *InternetArchiveServiceImpl) DownloadBook(bookId string, fileName string) (*http.Response, error) {
	url := fmt.Sprintf(
		"https://archive.org/download/%s/%s", bookId, fileName)

	request, buildReqErr := http.NewRequest(http.MethodGet, url, nil)

	if buildReqErr != nil {
		return nil, buildReqErr
	}

	httpClient := utils.GetCustomHttpClient(10 * time.Minute)
	response, requestErr := httpClient.Do(request)

	if requestErr != nil {
		return nil, requestErr
	}

	return response, nil
}
