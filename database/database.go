package database

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/adfer-dev/analock-api/utils"
	_ "github.com/tursodatabase/go-libsql"
)

const (
	createUsersTableQuery = "CREATE TABLE IF NOT EXISTS `user` (`id` integer, `email` text, 'username' text, `role` integer" +
		", PRIMARY KEY (`id`), UNIQUE (`email`));"
	createTokensTableQuery = "CREATE TABLE IF NOT EXISTS `token` (`id` integer, `value` text, `kind` integer, `user_id` text," +
		" PRIMARY KEY (`id`)," +
		" UNIQUE (`user_id`, `kind`)," +
		" CONSTRAINT `fk_users_tokens` FOREIGN KEY (`user_id`)" +
		" REFERENCES `user` (`id`) ON DELETE CASCADE ON UPDATE CASCADE);"
	createExternalLoginTableQuery = "CREATE TABLE IF NOT EXISTS `external_login`" +
		" (`id` integer, `provider` integer, `provider_client_id` text, `provider_client_token` text, `user_id` integer," +
		" PRIMARY KEY (`id`), UNIQUE (`provider_client_id`)," +
		" CONSTRAINT `fk_users_external_login` FOREIGN KEY (`user_id`)" +
		" REFERENCES `user` (`id`) ON DELETE CASCADE ON UPDATE CASCADE);"
	createDiaryEntryTableQuery = "CREATE TABLE IF NOT EXISTS `diary_entry` (`id` integer, `title` text, `content` text, `registration_id` integer," +
		" PRIMARY KEY (`id`)," +
		" UNIQUE (`id`, `registration_id`)," +
		" CONSTRAINT `fk_activity_registration_diary_entry` FOREIGN KEY (`registration_id`)" +
		" REFERENCES `activity_registration` (`id`) ON DELETE CASCADE ON UPDATE CASCADE);"
	createActivityRegistrationTableQuery = "CREATE TABLE IF NOT EXISTS `activity_registration` (" +
		"`id` integer PRIMARY KEY, " +
		"`registration_date` integer, " +
		"`user_id` integer, " +
		"CONSTRAINT `fk_activity_registration_user` FOREIGN KEY (`user_id`) " +
		"REFERENCES `user` (`id`) ON DELETE CASCADE ON UPDATE CASCADE);"
	createActivityRegistrationBookTableQuery = "CREATE TABLE IF NOT EXISTS `activity_registration_book` (" +
		"`id` integer PRIMARY KEY, " +
		"`registration_id` integer, " +
		"`internet_archive_id` text, " +
		"CONSTRAINT `fk_activity_registration` FOREIGN KEY (`registration_id`) " +
		"REFERENCES `activity_registration` (`id`) ON DELETE CASCADE ON UPDATE CASCADE);"
	createActivityRegistrationGameTableQuery = "CREATE TABLE IF NOT EXISTS `activity_registration_game` (" +
		"`id` integer PRIMARY KEY, " +
		"`registration_id` integer, " +
		"`game_name` text, " +
		"CONSTRAINT `fk_activity_registration` FOREIGN KEY (`registration_id`) " +
		"REFERENCES `activity_registration` (`id`) ON DELETE CASCADE ON UPDATE CASCADE);"
)

type Database struct {
	dbConnection *sql.DB
}

var connectionInstance *Database

func GetDatabaseInstance() *Database {

	if connectionInstance == nil {
		db, dbErr := sql.Open("libsql", fmt.Sprintf("%s?authToken=%s", os.Getenv("TURSO_DB_URL"), os.Getenv("TURSO_DB_TOKEN")))

		if dbErr != nil {
			utils.GetCustomLogger().Error(dbErr.Error())
		}

		if connErr := db.Ping(); connErr != nil {
			utils.GetCustomLogger().Error(connErr.Error())
		} else {
			utils.GetCustomLogger().Info("Connected to database")
		}

		connectionInstance = &Database{dbConnection: db}
		initDatabase()
	}

	return connectionInstance
}

func (conn *Database) GetConnection() *sql.DB {
	return connectionInstance.dbConnection
}

func initDatabase() {
	var createTableQueryMap map[string]string = make(map[string]string)

	createTableQueryMap["user"] = createUsersTableQuery
	createTableQueryMap["token"] = createTokensTableQuery
	createTableQueryMap["external_login"] = createExternalLoginTableQuery
	createTableQueryMap["diary_entry"] = createDiaryEntryTableQuery
	createTableQueryMap["activity_registration"] = createActivityRegistrationTableQuery
	createTableQueryMap["activity_registration_book"] = createActivityRegistrationBookTableQuery
	createTableQueryMap["activity_registration_game"] = createActivityRegistrationGameTableQuery

	for tableName, query := range createTableQueryMap {
		_, createTableErr := connectionInstance.GetConnection().Exec(query)
		if createTableErr != nil {
			utils.GetCustomLogger().Error(fmt.Sprintf("Error when creating table %s: %s", tableName, createTableErr.Error()))
		}
	}
}
