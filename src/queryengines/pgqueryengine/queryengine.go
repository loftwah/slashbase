package pgqueryengine

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"slashbase.com/backend/src/daos"
	"slashbase.com/backend/src/models"
	"slashbase.com/backend/src/models/sbsql"
	"slashbase.com/backend/src/queryengines/pgqueryengine/pgxutils"
	"slashbase.com/backend/src/sshtunnel"
)

var dbQueryLogDao daos.DBQueryLogDao

type PostgresQueryEngine struct {
	openConnections map[string]pgxConnPoolInstance
}

func InitPostgresQueryEngine() *PostgresQueryEngine {
	return &PostgresQueryEngine{
		openConnections: map[string]pgxConnPoolInstance{},
	}
}

func (pgqe *PostgresQueryEngine) RunQuery(user *models.User, dbConn *models.DBConnection, query string) (map[string]interface{}, error) {
	port, _ := strconv.Atoi(string(dbConn.DBPort))
	if dbConn.UseSSH != models.DBUSESSH_NONE {
		remoteHost := string(dbConn.DBHost)
		if remoteHost == "" {
			remoteHost = "localhost"
		}
		sshTun := sshtunnel.GetSSHTunnel(dbConn.ID, dbConn.UseSSH,
			string(dbConn.SSHHost), remoteHost, port, string(dbConn.SSHUser),
			string(dbConn.SSHPassword), string(dbConn.SSHKeyFile),
		)
		dbConn.DBHost = "localhost"
		dbConn.DBPort = sbsql.CryptedData(fmt.Sprintf("%d", sshTun.GetLocalEndpoint().Port))
	}
	port, _ = strconv.Atoi(string(dbConn.DBPort))
	conn, err := pgqe.getConnection(dbConn.ID, string(dbConn.DBHost), uint16(port), string(dbConn.DBName), string(dbConn.DBUser), string(dbConn.DBPassword))
	if err != nil {
		return nil, err
	}
	queryLog := models.NewQueryLog(user.ID, dbConn.ID, query)

	queryType := pgxutils.GetPSQLQueryType(query)
	if queryType == pgxutils.QUERY_READ {
		rows, err := conn.Query(context.Background(), query)
		if err != nil {
			return nil, err
		}
		columns, rowsData := pgxutils.PgSqlRowsToJson(rows)
		go dbQueryLogDao.CreateDBQueryLog(queryLog)
		return map[string]interface{}{
			"columns": columns,
			"rows":    rowsData,
		}, nil
	}
	cmdTag, err := conn.Exec(context.Background(), query)
	if err != nil {
		return nil, err
	}
	go dbQueryLogDao.CreateDBQueryLog(queryLog)
	return map[string]interface{}{
		"message": cmdTag.String(),
	}, nil
}

func (pgqe *PostgresQueryEngine) GetDataModels(user *models.User, dbConn *models.DBConnection) ([]map[string]interface{}, error) {
	data, err := pgqe.RunQuery(user, dbConn, "SELECT * FROM pg_catalog.pg_tables WHERE schemaname != 'pg_catalog' AND schemaname != 'information_schema';")
	if err != nil {
		return nil, err
	}
	rdata := data["rows"].([]map[string]interface{})
	return rdata, nil
}

func (pgqe *PostgresQueryEngine) GetSingleDataModelFields(user *models.User, dbConn *models.DBConnection, schema string, name string) ([]map[string]interface{}, error) {
	query := fmt.Sprintf(`
		SELECT *
		FROM
		(SELECT *
		FROM information_schema.columns
		WHERE table_schema = '%s' AND table_name = '%s') AS t1
		LEFT JOIN
		(SELECT conname,
				contype,
				unnest(conkey) AS conkey,
				pg_get_constraintdef(oid, TRUE) AS pretty_source
		FROM pg_constraint
		WHERE conrelid = '"%s"."%s"'::regclass AND contype = 'p') AS t2 ON t1.ordinal_position = t2.conkey order by t1.ordinal_position;`,
		schema, name, schema, name)
	data, err := pgqe.RunQuery(user, dbConn, query)
	if err != nil {
		return nil, err
	}
	rdata := data["rows"].([]map[string]interface{})
	return rdata, err
}

func (pgqe *PostgresQueryEngine) GetData(user *models.User, dbConn *models.DBConnection, schema string, name string, limit int, offset int64, fetchCount bool, filter []string, sort []string) (map[string]interface{}, error) {
	sortQuery := ""
	if len(sort) == 2 {
		sortQuery = fmt.Sprintf(` ORDER BY %s %s`, sort[0], sort[1])
	}
	query := fmt.Sprintf(`SELECT ctid, * FROM "%s"."%s"%s LIMIT %d OFFSET %d;`, schema, name, sortQuery, limit, offset)
	countQuery := fmt.Sprintf(`SELECT count(*) FROM "%s"."%s";`, schema, name)
	if len(filter) > 1 {
		filter2 := ""
		if len(filter) == 3 {
			filter2 = " '" + filter[2] + "'"
		}
		query = fmt.Sprintf(`SELECT ctid, * FROM "%s"."%s" WHERE "%s" %s%s%s LIMIT %d OFFSET %d;`,
			schema,
			name,
			filter[0],
			filter[1],
			filter2,
			sortQuery,
			limit,
			offset)
		countQuery = fmt.Sprintf(`SELECT count(*) FROM "%s"."%s" WHERE "%s" %s%s;`,
			schema,
			name,
			filter[0],
			filter[1],
			filter2)
	}
	data, err := pgqe.RunQuery(user, dbConn, query)
	if err != nil {
		return nil, err
	}
	if fetchCount {
		countData, err := pgqe.RunQuery(user, dbConn, countQuery)
		if err != nil {
			return nil, err
		}
		data["count"] = countData["rows"].([]map[string]interface{})[0]["count"]
	}
	return data, err
}

func (pgqe *PostgresQueryEngine) UpdateSingleData(user *models.User, dbConn *models.DBConnection, schema string, name string, ctid string, columnName string, value string) (map[string]interface{}, error) {
	query := fmt.Sprintf(`UPDATE "%s"."%s" SET "%s" = '%s' WHERE ctid = '%s' RETURNING ctid;`, schema, name, columnName, value, ctid)
	data, err := pgqe.RunQuery(user, dbConn, query)
	if err != nil {
		return nil, err
	}
	ctID := data["rows"].([]map[string]interface{})[0]["ctid"]
	data = map[string]interface{}{
		"ctid": ctID,
	}
	return data, err
}

func (pgqe *PostgresQueryEngine) AddData(user *models.User, dbConn *models.DBConnection, schema string, name string, data map[string]interface{}) (map[string]interface{}, error) {
	keys := []string{}
	values := []string{}
	for key, value := range data {
		keys = append(keys, key)
		val := value.(string)
		values = append(values, val)
	}
	keysStr := strings.Join(keys, ", ")
	valuesStr := strings.Join(values, "','")
	query := fmt.Sprintf(`INSERT INTO "%s"."%s"(%s) VALUES('%s') RETURNING ctid;`, schema, name, keysStr, valuesStr)
	rData, err := pgqe.RunQuery(user, dbConn, query)
	if err != nil {
		return nil, err
	}
	ctID := rData["rows"].([]map[string]interface{})[0]["ctid"]
	rData = map[string]interface{}{
		"ctid": ctID,
	}
	return rData, err
}

func (pgqe *PostgresQueryEngine) DeleteData(user *models.User, dbConn *models.DBConnection, schema string, name string, ctids []string) (map[string]interface{}, error) {
	ctidsStr := strings.Join(ctids, "', '")
	query := fmt.Sprintf(`DELETE FROM "%s"."%s" WHERE ctid IN ('%s');`, schema, name, ctidsStr)
	return pgqe.RunQuery(user, dbConn, query)
}
