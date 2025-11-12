package services

import (
	"errors"
	"testing"
	"time"

	"github.com/adfer-dev/analock-api/models"
	"github.com/stretchr/testify/assert"
)

// mockDiaryEntryStorage implements DiaryEntryStorageInterface
type mockDiaryEntryStorage struct {
	Entries      map[uint]*models.DiaryEntry
	UserEntries  map[uint][]*models.DiaryEntry
	GetErr       error
	GetByUIDErr  error
	GetByDateErr error
	CreateErr    error
	UpdateErr    error
}

func (m *mockDiaryEntryStorage) Get(id uint) (interface{}, error) {
	if m.GetErr != nil {
		return nil, m.GetErr
	}
	entry, ok := m.Entries[id]
	if !ok {
		return nil, errors.New("diary entry not found")
	}
	return entry, nil
}

func (m *mockDiaryEntryStorage) GetByUserId(userId uint) (interface{}, error) {
	if m.GetByUIDErr != nil {
		return nil, m.GetByUIDErr
	}
	entries, ok := m.UserEntries[userId]
	if !ok {
		return []*models.DiaryEntry{}, nil // Return empty slice if not found, consistent with some storage patterns
	}
	return entries, nil
}

func (m *mockDiaryEntryStorage) GetByUserIdAndDateInterval(userId uint, startDate int64, endDate int64) (interface{}, error) {
	if m.GetByDateErr != nil {
		return nil, m.GetByDateErr
	}
	userEntries, ok := m.UserEntries[userId]
	if !ok {
		return []*models.DiaryEntry{}, nil
	}
	var filteredEntries []*models.DiaryEntry
	for _, entry := range userEntries {
		if entry.Registration.RegistrationDate >= startDate && entry.Registration.RegistrationDate <= endDate {
			filteredEntries = append(filteredEntries, entry)
		}
	}
	return filteredEntries, nil
}

func (m *mockDiaryEntryStorage) Create(data interface{}) error {
	if m.CreateErr != nil {
		return m.CreateErr
	}
	entry, ok := data.(*models.DiaryEntry)
	if !ok {
		return errors.New("create: invalid type for DiaryEntry")
	}
	if entry.Id == 0 { // Simulate ID generation
		entry.Id = uint(len(m.Entries) + 1000) // Basic ID simulation
	}
	m.Entries[entry.Id] = entry
	if m.UserEntries == nil {
		m.UserEntries = make(map[uint][]*models.DiaryEntry)
	}
	m.UserEntries[entry.Registration.UserRefer] = append(m.UserEntries[entry.Registration.UserRefer], entry)
	return nil
}

func (m *mockDiaryEntryStorage) Update(data interface{}) error {
	if m.UpdateErr != nil {
		return m.UpdateErr
	}
	entry, ok := data.(*models.DiaryEntry)
	if !ok {
		return errors.New("update: invalid type for DiaryEntry")
	}
	_, exists := m.Entries[entry.Id]
	if !exists {
		return errors.New("update: diary entry not found")
	}
	m.Entries[entry.Id] = entry
	// Update in UserEntries as well
	userEntries, uExists := m.UserEntries[entry.Registration.UserRefer]
	if uExists {
		for i, ue := range userEntries {
			if ue.Id == entry.Id {
				m.UserEntries[entry.Registration.UserRefer][i] = entry
				break
			}
		}
	}
	return nil
}

var diaryEntryService DiaryEntryService = &DefaultDiaryEntryService{}

func TestGetDiaryEntryById(t *testing.T) {
	originalDiaryEntryStorage := diaryEntryStorage
	diaryEntryStorageMock := &mockDiaryEntryStorage{
		Entries: make(map[uint]*models.DiaryEntry),
	}
	diaryEntryStorage = diaryEntryStorageMock
	defer func() { diaryEntryStorage = originalDiaryEntryStorage }()

	testEntry := &models.DiaryEntry{Id: 1, Title: "Test Title", Content: "Test content"}
	diaryEntryStorageMock.Entries[testEntry.Id] = testEntry

	entry, err := diaryEntryService.GetDiaryEntryById(1)
	assert.NoError(t, err)
	assert.Equal(t, testEntry, entry)

	_, err = diaryEntryService.GetDiaryEntryById(2) // Non-existent
	assert.Error(t, err)

	diaryEntryStorageMock.GetErr = errors.New("forced Get error")
	_, err = diaryEntryService.GetDiaryEntryById(1)
	assert.Error(t, err)
	assert.EqualError(t, err, "forced Get error")
}

func TestGetUserEntries(t *testing.T) {
	originalDiaryEntryStorage := diaryEntryStorage
	diaryEntryStorageMock := &mockDiaryEntryStorage{
		UserEntries: make(map[uint][]*models.DiaryEntry),
	}
	diaryEntryStorage = diaryEntryStorageMock
	defer func() { diaryEntryStorage = originalDiaryEntryStorage }()

	userId := uint(1)
	expectedEntries := []*models.DiaryEntry{
		{Id: 1, Title: "Entry 1", Registration: models.ActivityRegistration{UserRefer: userId}},
		{Id: 2, Title: "Entry 2", Registration: models.ActivityRegistration{UserRefer: userId}},
	}
	diaryEntryStorageMock.UserEntries[userId] = expectedEntries

	entries, err := diaryEntryService.GetUserEntries(userId)
	assert.NoError(t, err)
	assert.Equal(t, expectedEntries, entries)

	otherEntries, err := diaryEntryService.GetUserEntries(2) // Non-existent user
	assert.NoError(t, err)
	assert.Empty(t, otherEntries)

	diaryEntryStorageMock.GetByUIDErr = errors.New("forced GetByUIDErr error")
	_, err = diaryEntryService.GetUserEntries(userId)
	assert.Error(t, err)
	assert.EqualError(t, err, "forced GetByUIDErr error")
}

func TestGetUserEntriesTimeRange(t *testing.T) {
	originalDiaryEntryStorage := diaryEntryStorage
	diaryEntryStorageMock := &mockDiaryEntryStorage{
		UserEntries: make(map[uint][]*models.DiaryEntry),
	}
	diaryEntryStorage = diaryEntryStorageMock
	defer func() { diaryEntryStorage = originalDiaryEntryStorage }()

	userId := uint(1)
	now := time.Now().Unix()
	entry1 := &models.DiaryEntry{Id: 1, Registration: models.ActivityRegistration{UserRefer: userId, RegistrationDate: now - 100}}
	entry2 := &models.DiaryEntry{Id: 2, Registration: models.ActivityRegistration{UserRefer: userId, RegistrationDate: now}}
	entry3 := &models.DiaryEntry{Id: 3, Registration: models.ActivityRegistration{UserRefer: userId, RegistrationDate: now + 100}}
	diaryEntryStorageMock.UserEntries[userId] = []*models.DiaryEntry{entry1, entry2, entry3}

	filtered, err := diaryEntryService.GetUserEntriesTimeRange(userId, now-50, now+50)
	assert.NoError(t, err)
	assert.Equal(t, []*models.DiaryEntry{entry2}, filtered)

	diaryEntryStorageMock.GetByDateErr = errors.New("forced GetByDateErr error")
	_, err = diaryEntryService.GetUserEntriesTimeRange(userId, now-50, now+50)
	assert.Error(t, err)
	assert.EqualError(t, err, "forced GetByDateErr error")
}

func TestSaveDiaryEntry(t *testing.T) {
	originalDiaryEntryStorage := diaryEntryStorage
	originalActivityRegistrationStorage := activityRegistrationStorage // ARS for ActivityRegistrationStorage

	diaryEntryStorageMock := &mockDiaryEntryStorage{
		Entries:     make(map[uint]*models.DiaryEntry),
		UserEntries: make(map[uint][]*models.DiaryEntry),
	}
	// Assuming mockActivityRegistrationStorage is available from activityRegistration_test.go
	activityRegistrationStorageMock := &mockActivityRegistrationStorage{}

	diaryEntryStorage = diaryEntryStorageMock
	activityRegistrationStorage = activityRegistrationStorageMock
	defer func() {
		diaryEntryStorage = originalDiaryEntryStorage
		activityRegistrationStorage = originalActivityRegistrationStorage
	}()

	saveBody := &SaveDiaryEntryBody{
		Title:       "New Diary Entry",
		Content:     "Diary content here.",
		PublishDate: time.Now().Unix(),
	}

	// --- Test successful save ---
	userId := uint(1)
	createdEntry, err := diaryEntryService.SaveDiaryEntry(saveBody, userId)

	assert.NoError(t, err)
	assert.NotNil(t, createdEntry)
	assert.Equal(t, saveBody.Title, createdEntry.Title)
	assert.Equal(t, saveBody.Content, createdEntry.Content)
	assert.Equal(t, saveBody.PublishDate, createdEntry.Registration.RegistrationDate)
	assert.Equal(t, userId, createdEntry.Registration.UserRefer)
	assert.NotNil(t, activityRegistrationStorageMock.CreatedActivity) // Check activity was passed to mock ARS
	assert.Equal(t, saveBody.PublishDate, activityRegistrationStorageMock.CreatedActivity.RegistrationDate)

	// --- Test error from activityRegistrationStorage.Create ---
	activityRegistrationStorageMock.Err = errors.New("ARS create failed")
	_, err = diaryEntryService.SaveDiaryEntry(saveBody, userId)
	assert.Error(t, err)
	assert.EqualError(t, err, "ARS create failed")
	activityRegistrationStorageMock.Err = nil // Reset error

	// --- Test error from diaryEntryStorage.Create ---
	diaryEntryStorageMock.CreateErr = errors.New("DES create failed")
	_, err = diaryEntryService.SaveDiaryEntry(saveBody, userId)
	assert.Error(t, err)
	assert.EqualError(t, err, "DES create failed")
	diaryEntryStorageMock.CreateErr = nil // Reset error
}

func TestUpdateDiaryEntry(t *testing.T) {
	originalDiaryEntryStorage := diaryEntryStorage
	originalActivityRegistrationStorage := activityRegistrationStorage

	diaryEntryStorageMock := &mockDiaryEntryStorage{
		Entries:     make(map[uint]*models.DiaryEntry),
		UserEntries: make(map[uint][]*models.DiaryEntry),
	}
	activityRegistrationStorageMock := &mockActivityRegistrationStorage{}

	diaryEntryStorage = diaryEntryStorageMock
	activityRegistrationStorage = activityRegistrationStorageMock
	defer func() {
		diaryEntryStorage = originalDiaryEntryStorage
		activityRegistrationStorage = originalActivityRegistrationStorage
	}()

	userId := uint(10)
	originalTime := time.Now().Unix() - 1000
	activityReg := models.ActivityRegistration{Id: 100, RegistrationDate: originalTime, UserRefer: userId}
	storedEntry := &models.DiaryEntry{
		Id:           1,
		Title:        "Original Title",
		Content:      "Original Content",
		Registration: activityReg,
	}
	diaryEntryStorageMock.Entries[storedEntry.Id] = storedEntry

	updateBody := &UpdateDiaryEntryBody{
		Title:       "Updated Title",
		Content:     "Updated Content",
		PublishDate: time.Now().Unix(),
	}

	// Test successful update
	updatedEntry, err := diaryEntryService.UpdateDiaryEntry(storedEntry.Id, updateBody)
	assert.NoError(t, err)
	assert.NotNil(t, updatedEntry)
	assert.Equal(t, updateBody.Title, updatedEntry.Title)
	assert.Equal(t, updateBody.Content, updatedEntry.Content)
	assert.Equal(t, updateBody.PublishDate, updatedEntry.Registration.RegistrationDate)
	assert.Equal(t, userId, updatedEntry.Registration.UserRefer)
	assert.Equal(t, activityReg.Id, updatedEntry.Registration.Id)
	assert.NotNil(t, activityRegistrationStorageMock.UpdatedActivity)
	assert.Equal(t, updateBody.PublishDate, activityRegistrationStorageMock.UpdatedActivity.RegistrationDate)

	// Test error from GetDiaryEntryById
	diaryEntryStorageMock.GetErr = errors.New("get failed for update")
	_, err = diaryEntryService.UpdateDiaryEntry(storedEntry.Id, updateBody)
	assert.Error(t, err)
	assert.EqualError(t, err, "get failed for update")
	diaryEntryStorageMock.GetErr = nil

	// Test error from activityRegistrationStorage.Update
	activityRegistrationStorageMock.UpdateErr = errors.New("ARS update failed")
	_, err = diaryEntryService.UpdateDiaryEntry(storedEntry.Id, updateBody)
	assert.Error(t, err)
	assert.EqualError(t, err, "ARS update failed")
	activityRegistrationStorageMock.UpdateErr = nil

	// Test error from diaryEntryStorage.Update
	diaryEntryStorageMock.UpdateErr = errors.New("DES update failed")
	_, err = diaryEntryService.UpdateDiaryEntry(storedEntry.Id, updateBody)
	assert.Error(t, err)
	assert.EqualError(t, err, "DES update failed")
	diaryEntryStorageMock.UpdateErr = nil
}

func TestDeleteDiaryEntry(t *testing.T) {
	originalDiaryEntryStorage := diaryEntryStorage
	originalActivityRegistrationStorage := activityRegistrationStorage

	diaryEntryStorageMock := &mockDiaryEntryStorage{
		Entries: make(map[uint]*models.DiaryEntry),
	}
	activityRegistrationStorageMock := &mockActivityRegistrationStorage{}

	diaryEntryStorage = diaryEntryStorageMock
	activityRegistrationStorage = activityRegistrationStorageMock
	defer func() {
		diaryEntryStorage = originalDiaryEntryStorage
		activityRegistrationStorage = originalActivityRegistrationStorage
	}()

	activityRegId := uint(200)
	entryToDelete := &models.DiaryEntry{Id: 2, Registration: models.ActivityRegistration{Id: activityRegId}}
	diaryEntryStorageMock.Entries[entryToDelete.Id] = entryToDelete

	// Test successful delete
	err := diaryEntryService.DeleteDiaryEntry(entryToDelete.Id)
	assert.NoError(t, err)
	assert.Equal(t, activityRegId, activityRegistrationStorageMock.DeletedId)

	// Test error from GetDiaryEntryById
	diaryEntryStorageMock.GetErr = errors.New("get failed for delete")
	err = diaryEntryService.DeleteDiaryEntry(entryToDelete.Id)
	assert.Error(t, err)
	assert.EqualError(t, err, "get failed for delete")
	diaryEntryStorageMock.GetErr = nil

	// Test error from activityRegistrationStorage.Delete
	// Need to ensure the entry is found again by GetDiaryEntryById for this sub-test
	diaryEntryStorageMock.Entries[entryToDelete.Id] = entryToDelete
	activityRegistrationStorageMock.DeleteErr = errors.New("ARS delete failed")
	err = diaryEntryService.DeleteDiaryEntry(entryToDelete.Id)
	assert.Error(t, err)
	assert.EqualError(t, err, "ARS delete failed")
	activityRegistrationStorageMock.DeleteErr = nil
}
