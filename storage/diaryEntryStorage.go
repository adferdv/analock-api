package storage

import (
	"database/sql"

	"github.com/adfer-dev/analock-api/database"
	"github.com/adfer-dev/analock-api/models"
)

const (
	getDiaryEntryByIdentifierQuery   = "SELECT de.id, de.title, de.content, ar.id, ar.registration_date, ar.user_id FROM diary_entry de INNER JOIN activity_registration ar ON (de.registration_id = ar.id) WHERE de.id = ?;"
	getUserDiaryEntriesQuery         = "SELECT de.id, de.title, de.content, ar.id, ar.registration_date, ar.user_id FROM diary_entry de INNER JOIN activity_registration ar ON (de.registration_id = ar.id) WHERE ar.user_id = ?;"
	getIntervalUserDiaryEntriesQuery = "SELECT de.id, de.title, de.content, ar.id, ar.registration_date, ar.user_id FROM diary_entry de INNER JOIN activity_registration ar ON (de.registration_id = ar.id) WHERE ar.user_id = ? AND ar.registration_date >= ? AND ar.registration_date <= ?;"
	insertDiaryEntryQuery            = "INSERT INTO diary_entry (title, content, registration_id) VALUES (?, ?, ?);"
	updateDiaryEntryQuery            = "UPDATE diary_entry SET title = ?, content = ? WHERE id = ?;"
	deleteDiaryEntryQuery            = "DELETE FROM diary_entry WHERE id = ?;"
)

type DiaryEntryStorageInterface interface {
	Get(id uint) (interface{}, error)
	GetByUserId(userId uint) (interface{}, error)
	GetByUserIdAndDateInterval(userId uint, startDate int64, endDate int64) (interface{}, error)
	Create(data interface{}) error
	Update(data interface{}) error
}

type DiaryEntryStorage struct{}

var diaryEntryNotFoundError = &models.DbNotFoundError{DbItem: &models.DiaryEntry{}}
var failedToParseDiaryEntryError = &models.DbCouldNotParseItemError{DbItem: &models.DiaryEntry{}}

func (diaryEntryStorage *DiaryEntryStorage) Get(id uint) (interface{}, error) {
	result, err := database.GetDatabaseInstance().GetConnection().Query(getDiaryEntryByIdentifierQuery, id)

	if err != nil {
		return nil, err
	}

	defer result.Close()

	if !result.Next() {
		return nil, diaryEntryNotFoundError
	}

	scannedDiaryEntry, scanErr := diaryEntryStorage.Scan(result)

	if scanErr != nil {
		return nil, scanErr
	}

	diaryEntry, ok := scannedDiaryEntry.(models.DiaryEntry)

	if !ok {
		return nil, failedToParseDiaryEntryError
	}
	return &diaryEntry, nil
}

func (diaryEntryStorage *DiaryEntryStorage) GetByUserId(userId uint) (interface{}, error) {
	userDiaryEntries := []*models.DiaryEntry{}
	result, err := database.GetDatabaseInstance().GetConnection().Query(getUserDiaryEntriesQuery, userId)

	if err != nil {
		return nil, err
	}

	defer result.Close()

	for result.Next() {
		scannedDiaryEntry, scanErr := diaryEntryStorage.Scan(result)

		if scanErr != nil {
			return nil, scanErr
		}
		diaryEntry, ok := scannedDiaryEntry.(models.DiaryEntry)

		if !ok {
			return nil, failedToParseDiaryEntryError
		}

		userDiaryEntries = append(userDiaryEntries, &diaryEntry)
	}

	return userDiaryEntries, nil
}

func (diaryEntryStorage *DiaryEntryStorage) GetByUserIdAndDateInterval(userId uint, startDate int64, endDate int64) (interface{}, error) {
	userDiaryEntries := []*models.DiaryEntry{}
	result, err := database.GetDatabaseInstance().GetConnection().Query(getIntervalUserDiaryEntriesQuery, userId, startDate, endDate)

	if err != nil {
		return nil, err
	}

	defer result.Close()

	for result.Next() {
		scannedDiaryEntry, scanErr := diaryEntryStorage.Scan(result)

		if scanErr != nil {
			return nil, scanErr
		}
		diaryEntry, ok := scannedDiaryEntry.(models.DiaryEntry)

		if !ok {
			return nil, failedToParseDiaryEntryError
		}

		userDiaryEntries = append(userDiaryEntries, &diaryEntry)
	}

	return userDiaryEntries, nil
}

func (diaryEntryStorage *DiaryEntryStorage) Create(diaryEntry interface{}) error {
	dbDiaryEntry, ok := diaryEntry.(*models.DiaryEntry)

	if !ok {
		return failedToParseDiaryEntryError
	}

	result, err := database.GetDatabaseInstance().GetConnection().Exec(insertDiaryEntryQuery,
		dbDiaryEntry.Title,
		dbDiaryEntry.Content,
		dbDiaryEntry.Registration.Id)

	if err != nil {
		return err
	}

	diaryEntryId, idErr := result.LastInsertId()
	if idErr != nil {
		return idErr
	}

	dbDiaryEntry.Id = uint(diaryEntryId)

	return nil
}

func (diaryEntryStorage *DiaryEntryStorage) Update(diaryEntry interface{}) error {
	dbDiaryEntry, ok := diaryEntry.(*models.DiaryEntry)

	if !ok {
		return failedToParseDiaryEntryError
	}

	result, err := database.GetDatabaseInstance().GetConnection().Exec(updateDiaryEntryQuery,
		dbDiaryEntry.Title,
		dbDiaryEntry.Content,
		dbDiaryEntry.Id)

	if err != nil {
		return err
	}

	affectedRows, errAffectedRows := result.RowsAffected()

	if errAffectedRows != nil {
		return errAffectedRows
	}

	if affectedRows == 0 {
		return diaryEntryNotFoundError
	}

	return nil
}

func (diaryEntryStorage *DiaryEntryStorage) Delete(id uint) error {
	result, err := database.GetDatabaseInstance().GetConnection().Exec(deleteDiaryEntryQuery, id)

	if err != nil {
		return err
	}

	affectedRows, errAffectedRows := result.RowsAffected()

	if errAffectedRows != nil {
		return errAffectedRows
	}

	if affectedRows == 0 {
		return diaryEntryNotFoundError
	}

	return nil
}

func (diaryEntryStorage *DiaryEntryStorage) Scan(rows *sql.Rows) (interface{}, error) {
	var diaryEntry models.DiaryEntry

	scanErr := rows.Scan(&diaryEntry.Id, &diaryEntry.Title, &diaryEntry.Content, &diaryEntry.Registration.Id,
		&diaryEntry.Registration.RegistrationDate, &diaryEntry.Registration.UserRefer)

	return diaryEntry, scanErr
}
