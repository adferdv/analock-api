package storage

import (
	"database/sql"
	"fmt"

	"github.com/adfer-dev/analock-api/database"
	"github.com/adfer-dev/analock-api/models"
	"github.com/adfer-dev/analock-api/utils"
)

const (
	getUserQuery            = "SELECT * FROM user where id = ?;"
	getUserByUserEmailQuery = "SELECT * FROM user where email = ?;"
	insertUserQuery         = "INSERT INTO user (email, username, role) VALUES (?, ?, ?);"
	updateUserQuery         = "UPDATE user SET username = ?, role = ? WHERE id = ?;"
	deleteUserQuery         = "DELETE FROM user WHERE id = ?;"
)

// UserStorageInterface defines storage operations for users.
type UserStorageInterface interface {
	Get(id uint) (interface{}, error)
	GetByEmail(email string) (interface{}, error)
	Create(data interface{}) error
	Update(data interface{}) error
	Delete(id uint) error
}

type UserStorage struct{}

var userNotFoundError = &models.DbNotFoundError{DbItem: &models.User{}}
var failedToParseUserError = &models.DbCouldNotParseItemError{DbItem: &models.User{}}

func (userStorage *UserStorage) Get(id uint) (interface{}, error) {
	result, err := database.GetDatabaseInstance().GetConnection().Query(getUserQuery, id)

	if err != nil {
		return nil, err
	}

	defer result.Close()

	if !result.Next() {
		return nil, userNotFoundError
	}

	scannedUser, scanErr := userStorage.Scan(result)

	if scanErr != nil {
		return nil, scanErr
	}

	user, ok := scannedUser.(*models.User)

	if !ok {
		return nil, failedToParseUserError
	}

	return user, nil
}

func (userStorage *UserStorage) GetByEmail(email string) (interface{}, error) {
	result, err := database.GetDatabaseInstance().GetConnection().Query(getUserByUserEmailQuery, email)

	if err != nil {
		return nil, err
	}

	defer result.Close()

	if !result.Next() {
		return nil, userNotFoundError
	}

	scannedUser, scanErr := userStorage.Scan(result)

	if scanErr != nil {
		return nil, scanErr
	}

	user, ok := scannedUser.(*models.User)

	if !ok {
		return nil, failedToParseUserError
	}

	return user, nil
}

func (userStorage *UserStorage) Create(user interface{}) error {
	dbUser, ok := user.(*models.User)
	userAlreadyExistsError := &models.DbItemAlreadyExistsError{DbItem: &models.User{}}

	if !ok {
		return failedToParseUserError
	}
	user, getUserErr := userStorage.Get(dbUser.Id)
	_, isNotFoundError := getUserErr.(*models.DbNotFoundError)

	if user != nil && !isNotFoundError {
		return userAlreadyExistsError
	}

	result, err := database.GetDatabaseInstance().GetConnection().Exec(insertUserQuery, dbUser.Email, dbUser.UserName, dbUser.Role)
	if err != nil {
		utils.GetCustomLogger().Error(fmt.Sprintf("error when saving user: %s", err.Error()))
		return err
	}

	userId, idErr := result.LastInsertId()
	if idErr != nil {
		return idErr
	}

	dbUser.Id = uint(userId)

	return nil
}

func (userStorage *UserStorage) Update(user interface{}) error {
	dbUser, ok := user.(*models.User)

	if !ok {
		return failedToParseUserError
	}

	result, err := database.GetDatabaseInstance().GetConnection().Exec(updateUserQuery, dbUser.UserName, dbUser.Role, dbUser.Id)

	if err != nil {
		return err
	}

	affectedRows, errAffectedRows := result.RowsAffected()

	if errAffectedRows != nil {
		return errAffectedRows
	}

	if affectedRows == 0 {
		return userNotFoundError
	}

	return nil
}

func (userStorage *UserStorage) Delete(id uint) error {

	result, err := database.GetDatabaseInstance().GetConnection().Exec(deleteUserQuery, id)

	if err != nil {
		return err
	}

	affectedRows, errAffectedRows := result.RowsAffected()

	if errAffectedRows != nil {
		return errAffectedRows
	}

	if affectedRows == 0 {
		return userNotFoundError
	}

	return nil
}

func (userStorage *UserStorage) Scan(rows *sql.Rows) (interface{}, error) {
	var user models.User

	scanErr := rows.Scan(&user.Id, &user.Email, &user.UserName, &user.Role)

	return &user, scanErr
}
