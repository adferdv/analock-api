package services

import (
	"errors"
	"testing"
	"time"

	"github.com/adfer-dev/analock-api/models"
	"github.com/stretchr/testify/assert"
)

type mockBookActivityRegistrationStorage struct {
	Registrations map[uint][]*models.BookActivityRegistration
	Err           error
}

func (m *mockBookActivityRegistrationStorage) GetByUserId(userId uint) (interface{}, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	regs, ok := m.Registrations[userId]
	if !ok {
		return []*models.BookActivityRegistration{}, nil
	}
	return regs, nil
}

func (m *mockBookActivityRegistrationStorage) GetByUserIdAndTimeRange(userId uint, startTime int64, endTime int64) (interface{}, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	userRegs, ok := m.Registrations[userId]
	if !ok {
		return []*models.BookActivityRegistration{}, nil
	}

	var filteredRegs []*models.BookActivityRegistration
	for _, reg := range userRegs {
		if reg.Registration.RegistrationDate >= startTime && reg.Registration.RegistrationDate <= endTime {
			filteredRegs = append(filteredRegs, reg)
		}
	}
	return filteredRegs, nil
}

func (m *mockBookActivityRegistrationStorage) Create(data interface{}) error {
	if m.Err != nil {
		return m.Err
	}

	reg, ok := data.(*models.BookActivityRegistration)
	if !ok {
		return &models.DbCouldNotParseItemError{DbItem: &models.BookActivityRegistration{}}
	}
	m.Registrations[reg.Registration.UserRefer] = append(m.Registrations[reg.Registration.UserRefer], reg)
	return nil
}

type mockGameActivityRegistrationStorage struct {
	Registrations map[uint][]*models.GameActivityRegistration
	Err           error
}

func (m *mockGameActivityRegistrationStorage) GetByUserId(userId uint) (interface{}, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	regs, ok := m.Registrations[userId]
	if !ok {
		return []*models.GameActivityRegistration{}, nil
	}
	return regs, nil
}

func (m *mockGameActivityRegistrationStorage) GetByUserIdAndInterval(userId uint, startDate int64, endDate int64) (interface{}, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	userRegs, ok := m.Registrations[userId]
	if !ok {
		return []*models.GameActivityRegistration{}, nil
	}

	var filteredRegs []*models.GameActivityRegistration
	for _, reg := range userRegs {
		if reg.Registration.RegistrationDate >= startDate && reg.Registration.RegistrationDate <= endDate {
			filteredRegs = append(filteredRegs, reg)
		}
	}
	return filteredRegs, nil
}

func (m *mockGameActivityRegistrationStorage) Create(data interface{}) error {
	if m.Err != nil {
		return m.Err
	}
	reg, ok := data.(*models.GameActivityRegistration)
	if !ok {
		return nil
	}
	m.Registrations[reg.Registration.UserRefer] = append(m.Registrations[reg.Registration.UserRefer], reg)
	return nil
}

type mockActivityRegistrationStorage struct {
	CreatedActivity *models.ActivityRegistration
	UpdatedActivity *models.ActivityRegistration
	DeletedId       uint
	Err             error
	UpdateErr       error
	DeleteErr       error
}

func (m *mockActivityRegistrationStorage) Create(data interface{}) error {
	if m.Err != nil {
		return m.Err
	}
	actReg, ok := data.(*models.ActivityRegistration)
	if !ok {
		return errors.New("Create: invalid type for ActivityRegistration")
	}
	m.CreatedActivity = actReg
	return nil
}

func (m *mockActivityRegistrationStorage) Update(data interface{}) error {
	if m.UpdateErr != nil {
		return m.UpdateErr
	}
	actReg, ok := data.(*models.ActivityRegistration)
	if !ok {
		return errors.New("Update: invalid type for ActivityRegistration")
	}
	m.UpdatedActivity = actReg
	return nil
}

func (m *mockActivityRegistrationStorage) Delete(id uint) error {
	if m.DeleteErr != nil {
		return m.DeleteErr
	}
	m.DeletedId = id
	return nil
}

var bookRegistrationService BookActivityRegistrationService = &BookActivityRegistrationServiceImpl{}
var gameRegistrationService GameActivityRegistrationService = &GameActivityRegistrationServiceImpl{}

func TestGetUserBookActivityRegistrations(t *testing.T) {
	originalBookStorage := bookActivityRegistrationStorage
	mockStorage := &mockBookActivityRegistrationStorage{
		Registrations: make(map[uint][]*models.BookActivityRegistration),
	}
	bookActivityRegistrationStorage = mockStorage // Temporally replace global storage with mock

	defer func() { bookActivityRegistrationStorage = originalBookStorage }() // Restore original

	// Setup test data
	userId := uint(1)
	expectedRegs := []*models.BookActivityRegistration{
		{InternetArchiveIdentifier: "ia_id1", Registration: models.ActivityRegistration{Id: 1, RegistrationDate: time.Now().Unix(), UserRefer: userId}},
	}
	mockStorage.Registrations[userId] = expectedRegs

	regs, err := bookRegistrationService.GetUserBookActivityRegistrations(userId)

	assert.NoError(t, err)
	assert.Equal(t, expectedRegs, regs)

	// Test case for error
	mockStorage.Err = assert.AnError // Simulate an error
	_, err = bookRegistrationService.GetUserBookActivityRegistrations(userId)
	assert.Error(t, err)
	mockStorage.Err = nil

	// Test case for no registrations
	_, err = bookRegistrationService.GetUserBookActivityRegistrations(2) // Different user ID
	assert.NoError(t, err)
}

func TestGetUserGameActivityRegistrations(t *testing.T) {
	originalGameStorage := gameActivityRegistrationStorage
	mockStorage := &mockGameActivityRegistrationStorage{
		Registrations: make(map[uint][]*models.GameActivityRegistration),
	}
	gameActivityRegistrationStorage = mockStorage

	defer func() { gameActivityRegistrationStorage = originalGameStorage }()

	userId := uint(1)
	expectedRegs := []*models.GameActivityRegistration{
		{GameName: "game1", Registration: models.ActivityRegistration{Id: 1, RegistrationDate: time.Now().Unix(), UserRefer: userId}},
	}
	mockStorage.Registrations[userId] = expectedRegs

	regs, err := gameRegistrationService.GetUserGameActivityRegistrations(userId)

	assert.NoError(t, err)
	assert.Equal(t, expectedRegs, regs)

	// Test case for error
	mockStorage.Err = assert.AnError
	_, err = gameRegistrationService.GetUserGameActivityRegistrations(userId)
	assert.Error(t, err)
	mockStorage.Err = nil
}

func TestGetUserBookActivityRegistrationsTimeRange(t *testing.T) {
	originalBookStorage := bookActivityRegistrationStorage
	mockStorage := &mockBookActivityRegistrationStorage{
		Registrations: make(map[uint][]*models.BookActivityRegistration),
	}
	bookActivityRegistrationStorage = mockStorage
	defer func() { bookActivityRegistrationStorage = originalBookStorage }()

	userId := uint(1)
	now := time.Now().Unix()
	reg1 := &models.BookActivityRegistration{InternetArchiveIdentifier: "ia_id1", Registration: models.ActivityRegistration{Id: 1, RegistrationDate: now - 100, UserRefer: userId}}
	reg2 := &models.BookActivityRegistration{InternetArchiveIdentifier: "ia_id2", Registration: models.ActivityRegistration{Id: 2, RegistrationDate: now, UserRefer: userId}}
	reg3 := &models.BookActivityRegistration{InternetArchiveIdentifier: "ia_id3", Registration: models.ActivityRegistration{Id: 3, RegistrationDate: now + 100, UserRefer: userId}}
	mockStorage.Registrations[userId] = []*models.BookActivityRegistration{reg1, reg2, reg3}

	// Test case: Get registrations within a specific time range
	regs, err := bookRegistrationService.GetUserBookActivityRegistrationsTimeRange(userId, now-50, now+50)
	assert.NoError(t, err)
	assert.Equal(t, []*models.BookActivityRegistration{reg2}, regs)

	// Test case: Error from storage
	mockStorage.Err = assert.AnError
	_, err = bookRegistrationService.GetUserBookActivityRegistrationsTimeRange(userId, now-50, now+50)
	assert.Error(t, err)
	mockStorage.Err = nil
}

func TestGetUserGameActivityRegistrationsTimeRange(t *testing.T) {
	originalGameStorage := gameActivityRegistrationStorage
	mockStorage := &mockGameActivityRegistrationStorage{
		Registrations: make(map[uint][]*models.GameActivityRegistration),
	}
	gameActivityRegistrationStorage = mockStorage
	defer func() { gameActivityRegistrationStorage = originalGameStorage }()

	userId := uint(1)
	now := time.Now().Unix()
	reg1 := &models.GameActivityRegistration{GameName: "game1", Registration: models.ActivityRegistration{Id: 1, RegistrationDate: now - 100, UserRefer: userId}}
	reg2 := &models.GameActivityRegistration{GameName: "game2", Registration: models.ActivityRegistration{Id: 2, RegistrationDate: now, UserRefer: userId}}
	reg3 := &models.GameActivityRegistration{GameName: "game3", Registration: models.ActivityRegistration{Id: 3, RegistrationDate: now + 100, UserRefer: userId}}
	mockStorage.Registrations[userId] = []*models.GameActivityRegistration{reg1, reg2, reg3}

	regs, err := gameRegistrationService.GetUserGameActivityRegistrationsTimeRange(userId, now-50, now+50)
	assert.NoError(t, err)
	assert.Equal(t, []*models.GameActivityRegistration{reg2}, regs)

	// Test case: Error from storage
	mockStorage.Err = assert.AnError
	_, err = gameRegistrationService.GetUserGameActivityRegistrationsTimeRange(userId, now-50, now+50)
	assert.Error(t, err)
	mockStorage.Err = nil
}

func TestCreateBookActivityRegistration(t *testing.T) {
	originalBookStorage := bookActivityRegistrationStorage
	originalActivityStorage := activityRegistrationStorage

	mockBookStore := &mockBookActivityRegistrationStorage{
		Registrations: make(map[uint][]*models.BookActivityRegistration),
	}
	mockActivityStore := &mockActivityRegistrationStorage{}

	bookActivityRegistrationStorage = mockBookStore
	activityRegistrationStorage = mockActivityStore

	defer func() {
		bookActivityRegistrationStorage = originalBookStorage
		activityRegistrationStorage = originalActivityStorage
	}()

	addRegBody := &AddBookActivityRegistrationBody{
		InternetArchiveId: "test_ia_id",
		RegistrationDate:  time.Now().Unix(),
	}

	userRefer := uint(1)
	createdReg, err := bookRegistrationService.CreateBookActivityRegistration(addRegBody, userRefer)

	assert.NoError(t, err)
	assert.NotNil(t, createdReg)
	assert.Equal(t, addRegBody.InternetArchiveId, createdReg.InternetArchiveIdentifier)
	assert.Equal(t, addRegBody.RegistrationDate, createdReg.Registration.RegistrationDate)
	assert.Equal(t, userRefer, createdReg.Registration.UserRefer)

	// Assert that the generic activity registration was also "created"
	assert.NotNil(t, mockActivityStore.CreatedActivity)
	assert.Equal(t, addRegBody.RegistrationDate, mockActivityStore.CreatedActivity.RegistrationDate)
	assert.Equal(t, userRefer, mockActivityStore.CreatedActivity.UserRefer)

	// Test case: Error during activity registration creation
	mockActivityStore.Err = assert.AnError
	_, err = bookRegistrationService.CreateBookActivityRegistration(addRegBody, userRefer)
	assert.Error(t, err)
	mockActivityStore.Err = nil // Reset error

	// Test case: Error during book activity registration creation
	mockBookStore.Err = assert.AnError
	_, err = bookRegistrationService.CreateBookActivityRegistration(addRegBody, userRefer)
	assert.Error(t, err)
	mockBookStore.Err = nil // Reset error
}

func TestCreateGameActivityRegistration(t *testing.T) {
	originalGameStorage := gameActivityRegistrationStorage
	originalActivityStorage := activityRegistrationStorage

	mockGameStore := &mockGameActivityRegistrationStorage{
		Registrations: make(map[uint][]*models.GameActivityRegistration),
	}
	mockActivityStore := &mockActivityRegistrationStorage{}

	gameActivityRegistrationStorage = mockGameStore
	activityRegistrationStorage = mockActivityStore

	defer func() {
		gameActivityRegistrationStorage = originalGameStorage
		activityRegistrationStorage = originalActivityStorage
	}()

	addRegBody := &AddGameActivityRegistrationBody{
		GameName:         "test_game",
		RegistrationDate: time.Now().Unix(),
	}
	userRefer := uint(1)

	createdReg, err := gameRegistrationService.CreateGameActivityRegistration(addRegBody, userRefer)

	assert.NoError(t, err)
	assert.NotNil(t, createdReg)
	assert.Equal(t, addRegBody.GameName, createdReg.GameName)
	assert.Equal(t, addRegBody.RegistrationDate, createdReg.Registration.RegistrationDate)
	assert.Equal(t, userRefer, createdReg.Registration.UserRefer)

	assert.NotNil(t, mockActivityStore.CreatedActivity)
	assert.Equal(t, addRegBody.RegistrationDate, mockActivityStore.CreatedActivity.RegistrationDate)
	assert.Equal(t, userRefer, mockActivityStore.CreatedActivity.UserRefer)

	// Test case: Error during activity registration creation
	mockActivityStore.Err = assert.AnError
	_, err = gameRegistrationService.CreateGameActivityRegistration(addRegBody, userRefer)
	assert.Error(t, err)
	mockActivityStore.Err = nil

	// Test case: Error during game activity registration creation
	mockGameStore.Err = assert.AnError
	_, err = gameRegistrationService.CreateGameActivityRegistration(addRegBody, userRefer)
	assert.Error(t, err)
	mockGameStore.Err = nil
}
