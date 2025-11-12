package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/adfer-dev/analock-api/constants"
	"github.com/adfer-dev/analock-api/models"
	"github.com/adfer-dev/analock-api/services"
	"github.com/adfer-dev/analock-api/utils"
	"github.com/gorilla/mux"
)

var diaryEntryService services.DiaryEntryService = &services.DefaultDiaryEntryService{}

func InitDiaryEntryRoutes(router *mux.Router) {
	router.HandleFunc("/api/v1/diaryEntries/user/{id:[0-9]+}", utils.ParseToHandlerFunc(handleGetUserEntries)).Methods("GET")
	router.HandleFunc("/api/v1/diaryEntries", utils.ParseToHandlerFunc(handleCreateDiaryEntry)).Methods("POST")
	router.HandleFunc("/api/v1/diaryEntries/{id:[0-9]+}", utils.ParseToHandlerFunc(handleUpdateDiaryEntry)).Methods("PUT")
}

// @Summary		Get user diary entries
// @Description	Get all diary entries for a user, optionally filtered by date range
// @Tags			diary
// @Accept			json
// @Produce		json
// @Param			id			path		int	true	"User ID"
// @Param			startDate	query		int	false	"Start date timestamp"
// @Param			endDate		query		int	false	"End date timestamp"
// @Success		200			{array}		models.DiaryEntry
// @Failure		400			{object}	models.HttpError
// @Failure		500			{object}	models.HttpError
// @Security		BearerAuth
// @Router			/diaryEntries/user/{id} [get]
func handleGetUserEntries(res http.ResponseWriter, req *http.Request) error {
	userId, _ := strconv.Atoi(mux.Vars(req)["id"])

	startDateString := req.URL.Query().Get(constants.StartDateQueryParam)
	endDateString := req.URL.Query().Get(constants.EndDateQueryParam)

	if len(startDateString) == 0 || len(endDateString) == 0 {
		userDiaryEntries, err := services.GetCacheServiceInstance().CacheResource(
			func() (interface{}, error) { return diaryEntryService.GetUserEntries(uint(userId)) },
			constants.DiaryEntriesCacheResource,
			utils.BuildUserCacheKey(uint(userId)),
		)

		if err != nil {
			return utils.WriteJSON(res, 500, err.Error())
		}

		return utils.WriteJSON(res, 200, userDiaryEntries)
	}

	startDate, startDateErr := strconv.Atoi(startDateString)

	if startDateErr != nil {
		return utils.WriteJSON(res, 400, models.HttpError{Status: http.StatusBadRequest, Description: fmt.Sprintf(constants.QueryParamError, constants.StartDateQueryParam)})
	}
	endDate, endDateErr := strconv.Atoi(endDateString)

	if endDateErr != nil {
		return utils.WriteJSON(res, 400, models.HttpError{Status: http.StatusBadRequest, Description: fmt.Sprintf(constants.QueryParamError, constants.EndDateQueryParam)})
	}
	dateIntervalUserDiaryEntries, err := diaryEntryService.GetUserEntriesTimeRange(uint(userId), int64(startDate), int64(endDate))
	if err != nil {
		return utils.WriteJSON(res, 500, err.Error())
	}

	return utils.WriteJSON(res, 200, dateIntervalUserDiaryEntries)
}

// @Summary		Create diary entry
// @Description	Create a new diary entry for a user
// @Tags			diary
// @Accept			json
// @Produce		json
// @Param			body	body		services.SaveDiaryEntryBody	true	"Diary entry information"
// @Success		201		{object}	models.DiaryEntry
// @Failure		400		{object}	models.HttpError
// @Failure		500		{object}	models.HttpError
// @Security		BearerAuth
// @Router			/diaryEntries [post]
func handleCreateDiaryEntry(res http.ResponseWriter, req *http.Request) error {
	entryBody := services.SaveDiaryEntryBody{}

	validationErrs := utils.HandleValidation(req, &entryBody)

	if len(validationErrs) > 0 {
		return utils.WriteJSON(res, 400, validationErrs)
	}

	tokenClaims, claimsErr := utils.GetTokenClaimsFromRequest(req)

	if claimsErr != nil {
		utils.GetCustomLogger().Errorf(
			"Error getting claims on create game registration: %s",
			claimsErr.Error(),
		)
		return utils.WriteJSON(res, 500, constants.ErrorGeneric)
	}

	userId := uint(tokenClaims["sub"].(float64))

	savedEntry, saveEntryErr := diaryEntryService.SaveDiaryEntry(&entryBody, userId)
	services.GetCacheServiceInstance().EvictResourceItem(
		constants.DiaryEntriesCacheResource,
		utils.BuildUserCacheKey(userId),
	)

	if saveEntryErr != nil {
		return utils.WriteJSON(res, 500, saveEntryErr.Error())
	}

	return utils.WriteJSON(res, 201, savedEntry)
}

// @Summary		Update diary entry
// @Description	Update an existing diary entry
// @Tags			diary
// @Accept			json
// @Produce		json
// @Param			id		path		int								true	"Diary entry ID"
// @Param			body	body		services.UpdateDiaryEntryBody	true	"Updated diary entry information"
// @Success		200		{object}	models.DiaryEntry
// @Failure		400		{object}	models.HttpError
// @Failure		500		{object}	models.HttpError
// @Security		BearerAuth
// @Router			/diaryEntries/{id} [put]
func handleUpdateDiaryEntry(res http.ResponseWriter, req *http.Request) error {
	entryId, _ := strconv.Atoi(mux.Vars(req)["id"])
	updateEntryBody := services.UpdateDiaryEntryBody{}

	validationErrs := utils.HandleValidation(req, &updateEntryBody)

	if len(validationErrs) > 0 {
		return utils.WriteJSON(res, 400, validationErrs)
	}

	updatedEntry, updateEntryErr := diaryEntryService.UpdateDiaryEntry(uint(entryId), &updateEntryBody)

	services.GetCacheServiceInstance().EvictResourceItem(
		constants.DiaryEntriesCacheResource,
		utils.BuildUserCacheKey(updatedEntry.Registration.UserRefer),
	)

	if updateEntryErr != nil {
		return utils.WriteJSON(res, 500, updateEntryErr.Error())
	}

	return utils.WriteJSON(res, 200, updatedEntry)
}
