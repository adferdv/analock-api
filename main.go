package main

import (
	"fmt"
	"log"

	"github.com/adfer-dev/analock-api/api"
	"github.com/adfer-dev/analock-api/utils"
	"github.com/joho/godotenv"
)

//	@title			Analock API
//	@version		1.0
//	@description	This is the API server for the Analock application.
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@BasePath					/api/v1
//	@schemes					http https
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and JWT token.
//	@security					[{ "BearerAuth": [] }]

func main() {
	err := godotenv.Load()

	if err != nil {
		log.Fatal("No env file is present")
	}

	server := api.APIServer{Port: 3000}

	utils.GetCustomLogger().Info(fmt.Sprintf("Server listening at port %d...\n", server.Port))
	utils.GetCustomLogger().Error(server.Run().Error())
}
