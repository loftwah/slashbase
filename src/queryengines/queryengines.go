package queryengines

import (
	"errors"

	"slashbase.com/backend/src/models"
	"slashbase.com/backend/src/queryengines/pgqueryengine"
	"slashbase.com/backend/src/queryengines/pgqueryengine/pgxutils"
	"slashbase.com/backend/src/utils"
)

var postgresQueryEngine *pgqueryengine.PostgresQueryEngine

func InitQueryEngines() {
	postgresQueryEngine = pgqueryengine.InitPostgresQueryEngine()
}

func RunQuery(user *models.User, dbConn *models.DBConnection, query string, userRole string) (map[string]interface{}, error) {
	queryType := pgxutils.GetPSQLQueryType(query)
	isAllowed := false
	if queryType == pgxutils.QUERY_READ && utils.ContainsString([]string{models.ROLE_ANALYST, models.ROLE_ADMIN, models.ROLE_DEVELOPER}, userRole) {
		isAllowed = true
	} else if queryType == pgxutils.QUERY_WRITE && utils.ContainsString([]string{models.ROLE_ADMIN, models.ROLE_DEVELOPER}, userRole) {
		isAllowed = true
	} else if queryType == pgxutils.QUERY_ALTER && utils.ContainsString([]string{models.ROLE_ADMIN, models.ROLE_DEVELOPER}, userRole) {
		isAllowed = true
	}
	if !isAllowed {
		return nil, errors.New("not allowed")
	}
	return postgresQueryEngine.RunQuery(user, dbConn, query)
}

func GetDataModels(user *models.User, dbConn *models.DBConnection) ([]*DBDataModel, error) {
	data, err := postgresQueryEngine.GetDataModels(user, dbConn)
	if err != nil {
		return nil, err
	}
	dataModels := []*DBDataModel{}
	for _, table := range data {
		view := BuildDBDataModel(dbConn, table)
		if view != nil {
			dataModels = append(dataModels, view)
		}
	}
	return dataModels, nil
}

func GetSingleDataModel(user *models.User, dbConn *models.DBConnection, schemaName string, name string) (*DBDataModel, error) {
	fieldsData, err := postgresQueryEngine.GetSingleDataModelFields(user, dbConn, schemaName, name)
	if err != nil {
		return nil, err
	}
	allFields := []DBDataModelField{}
	for _, field := range fieldsData {
		fieldView := BuildDBDataModelField(dbConn, field)
		if fieldView != nil {
			allFields = append(allFields, *fieldView)
		}
	}

	dataModels := DBDataModel{
		SchemaName: schemaName,
		Name:       name,
		Fields:     allFields,
	}
	return &dataModels, nil
}

func GetData(user *models.User, dbConn *models.DBConnection, schemaName string, name string, limit int, offset int64, fetchCount bool, filter []string, sort []string) (map[string]interface{}, error) {
	return postgresQueryEngine.GetData(user, dbConn, schemaName, name, limit, offset, fetchCount, filter, sort)
}

func UpdateSingleData(user *models.User, dbConn *models.DBConnection, schemaName string, name string, ctid string, columnName, value string) (map[string]interface{}, error) {
	return postgresQueryEngine.UpdateSingleData(user, dbConn, schemaName, name, ctid, columnName, value)
}

func AddData(user *models.User, dbConn *models.DBConnection, schemaName string, name string, data map[string]interface{}) (map[string]interface{}, error) {
	return postgresQueryEngine.AddData(user, dbConn, schemaName, name, data)
}

func DeleteData(user *models.User, dbConn *models.DBConnection, schemaName string, name string, ctids []string) (map[string]interface{}, error) {
	return postgresQueryEngine.DeleteData(user, dbConn, schemaName, name, ctids)
}

func RemoveUnusedConnections() {
	postgresQueryEngine.RemoveUnusedConnections()
}
