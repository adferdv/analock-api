package constants

const StartDateQueryParam = "start_date"
const EndDateQueryParam = "end_date"
const QueryParamError = "the query parameter %s is not provided or its format is not correct."
const ErrorUnauthorizedOperation = "you have no permissions over the resource you are trying to access to"
const ErrorGeneric = "something went wrong, please try again"
const ApiV1UrlRoot = "/api/v1"
const ApiUrlDiaryEntries = "/diaryEntries"
const ApiUrlUserDiaryEntries = "/diaryEntries/user"
const ApiUrlBookRegistrations = "/activityRegistrations/books"
const ApiUrlGameRegistrations = "/activityRegistrations/games"
const ApiGoogleTokenValidationUrl = "https://www.googleapis.com/oauth2/v3/tokeninfo"
const DiaryEntriesCacheResource = "diaryEntries"
const BookActivityRegistrationsCacheResource = "bookActivityRegistrations"
const GameActivityRegistrationsCacheResource = "gameActivityRegistrations"

// TEST CONSTANTS
const TestAccessTokenValue = "mock_access_jwt_from_manager_v_agnostic"
const TestRefreshTokenValue = "mock_refresh_jwt_from_manager_v_agnostic"
