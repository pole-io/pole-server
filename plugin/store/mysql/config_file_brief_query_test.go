package sqldb

import (
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestQueryConfigFilesBriefReturnsMetadataWithoutTreatingControlFieldsAsColumns(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()

	db := &BaseDB{DB: rawDB}
	store := &configFileStore{master: db, slave: db}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM config_file WHERE flag = 0 ")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(uint32(1)))
	query := store.baseSelectConfigFileSql(true) + " WHERE flag = 0  ORDER BY id DESC LIMIT ?, ? "
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(uint32(0), uint32(10)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "namespace", "group", "comment", "format",
			"ctime", "create_by", "mtime", "modify_by", "metadata",
		}).AddRow("1", "app.yaml", "dev", "application", "summary", "yaml", 1, "admin", 2, "admin", "{}"))

	total, files, err := store.QueryConfigFiles(map[string]string{
		"brief":  "true",
		"offset": "0",
		"limit":  "10",
	}, 0, 10)
	require.NoError(t, err)
	require.Equal(t, uint32(1), total)
	require.Len(t, files, 1)
	require.Empty(t, files[0].Content)
	require.NoError(t, mock.ExpectationsWereMet())
}
