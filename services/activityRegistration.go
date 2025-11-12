package services

import (
	"github.com/adfer-dev/analock-api/models"
	"github.com/adfer-dev/analock-api/storage"
)

// BookActicityRegistrationService interface and implementation
type BookActivityRegistrationService interface {
	GetUserBookActivityRegistrations(userId uint) ([]*models.BookActivityRegistration, error)
	GetUserBookActivityRegistrationsTimeRange(userId uint, startTime int64, endTime int64) ([]*models.BookActivityRegistration, error)
	CreateBookActivityRegistration(addRegistrationBody *AddBookActivityRegistrationBody, userId uint) (*models.BookActivityRegistration, error)
}
type BookActivityRegistrationServiceImpl struct{}

// GameActicityRegistrationService interface and implementation
type GameActivityRegistrationService interface {
	GetUserGameActivityRegistrations(userId uint) ([]*models.GameActivityRegistration, error)
	GetUserGameActivityRegistrationsTimeRange(userId uint, startDate int64, endDate int64) ([]*models.GameActivityRegistration, error)
	CreateGameActivityRegistration(addRegistrationBody *AddGameActivityRegistrationBody, userId uint) (*models.GameActivityRegistration, error)
}
type GameActivityRegistrationServiceImpl struct{}

// Request bodies structs
type AddBookActivityRegistrationBody struct {
	InternetArchiveId string `json:"internetArchiveId" validate:"required"`
	RegistrationDate  int64  `json:"registrationDate" validate:"required"`
}

type AddGameActivityRegistrationBody struct {
	GameName         string `json:"gameName" validate:"required"`
	RegistrationDate int64  `json:"registrationDate" validate:"required"`
}

var bookActivityRegistrationStorage storage.BookActivityRegistrationStorageInterface = &storage.BookActivityRegistrationStorage{}
var gameActivityRegistrationStorage storage.GameActivityRegistrationStorageInterface = &storage.GameActivityRegistrationStorage{}
var activityRegistrationStorage storage.ActivityRegistrationStorageInterface = &storage.ActivityRegistrationStorage{}

func (bookActivityRegistrationService *BookActivityRegistrationServiceImpl) GetUserBookActivityRegistrations(userId uint) ([]*models.BookActivityRegistration, error) {
	dbUserRegistrations, err := bookActivityRegistrationStorage.GetByUserId(userId)

	if err != nil {
		return nil, err
	}

	return dbUserRegistrations.([]*models.BookActivityRegistration), nil
}

func (bookActivityRegistrationService *GameActivityRegistrationServiceImpl) GetUserGameActivityRegistrations(userId uint) ([]*models.GameActivityRegistration, error) {
	dbUserRegistrations, err := gameActivityRegistrationStorage.GetByUserId(userId)

	if err != nil {
		return nil, err
	}

	return dbUserRegistrations.([]*models.GameActivityRegistration), nil
}

func (bookActivityRegistrationService *BookActivityRegistrationServiceImpl) GetUserBookActivityRegistrationsTimeRange(userId uint, startTime int64, endTime int64) ([]*models.BookActivityRegistration, error) {
	dbUserRegistrations, err := bookActivityRegistrationStorage.GetByUserIdAndTimeRange(userId, startTime, endTime)

	if err != nil {
		return nil, err
	}

	return dbUserRegistrations.([]*models.BookActivityRegistration), nil
}

func (gameActivityRegistrationService *GameActivityRegistrationServiceImpl) GetUserGameActivityRegistrationsTimeRange(userId uint, startDate int64, endDate int64) ([]*models.GameActivityRegistration, error) {
	dbUserRegistrations, err := gameActivityRegistrationStorage.GetByUserIdAndInterval(userId, startDate, endDate)

	if err != nil {
		return nil, err
	}

	return dbUserRegistrations.([]*models.GameActivityRegistration), nil
}

func (bookActivityRegistrationService *BookActivityRegistrationServiceImpl) CreateBookActivityRegistration(addRegistrationBody *AddBookActivityRegistrationBody, userId uint) (*models.BookActivityRegistration, error) {
	dbActivityRegistration := &models.ActivityRegistration{
		RegistrationDate: addRegistrationBody.RegistrationDate,
		UserRefer:        userId,
	}
	createActivityRegistrationErr := activityRegistrationStorage.Create(dbActivityRegistration)

	if createActivityRegistrationErr != nil {
		return nil, createActivityRegistrationErr
	}

	dbBookActivityRegistration := &models.BookActivityRegistration{
		InternetArchiveIdentifier: addRegistrationBody.InternetArchiveId,
		Registration:              *dbActivityRegistration,
	}

	createBookActivityRegistrationErr := bookActivityRegistrationStorage.Create(dbBookActivityRegistration)

	if createBookActivityRegistrationErr != nil {
		return nil, createBookActivityRegistrationErr
	}

	return dbBookActivityRegistration, nil
}

func (gameActivityRegistrationService *GameActivityRegistrationServiceImpl) CreateGameActivityRegistration(addRegistrationBody *AddGameActivityRegistrationBody, userId uint) (*models.GameActivityRegistration, error) {

	dbActivityRegistration := &models.ActivityRegistration{
		RegistrationDate: addRegistrationBody.RegistrationDate,
		UserRefer:        userId,
	}
	createActivityRegistrationErr := activityRegistrationStorage.Create(dbActivityRegistration)

	if createActivityRegistrationErr != nil {
		return nil, createActivityRegistrationErr
	}

	dbGameActivityRegistration := &models.GameActivityRegistration{
		GameName:     addRegistrationBody.GameName,
		Registration: *dbActivityRegistration,
	}

	createGameActivityRegistrationErr := gameActivityRegistrationStorage.Create(dbGameActivityRegistration)

	if createGameActivityRegistrationErr != nil {
		return nil, createGameActivityRegistrationErr
	}

	return dbGameActivityRegistration, nil
}
