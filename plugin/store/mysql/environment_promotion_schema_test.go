package sqldb

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestEnsureEnvironmentPromotionSchema(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS environment_promotion_topology ").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS environment_promotion_topology_revision ").
		WillReturnResult(sqlmock.NewResult(0, 0))

	require.NoError(t, ensureEnvironmentPromotionSchema(&BaseDB{DB: rawDB}))
	require.NoError(t, mock.ExpectationsWereMet())
}
