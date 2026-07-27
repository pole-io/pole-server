package sqldb

import (
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"

	storeapi "github.com/pole-io/pole-server/apis/store"
)

func TestDeleteLogicalServiceLocksRootBeforeCheckingBindings(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()

	store := &logicalServiceStore{master: &BaseDB{DB: rawDB}}
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id FROM logical_service
		WHERE id = ? AND flag = 0 FOR UPDATE`)).
		WithArgs("logical-1").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("logical-1"))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM service_environment_binding
		WHERE logical_service_id = ?`)).
		WithArgs("logical-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec("UPDATE logical_service SET flag = 1").
		WithArgs("logical-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	require.NoError(t, store.DeleteLogicalService("logical-1"))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUnbindServiceEnvironmentRejectsMismatchedLogicalService(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()

	store := &logicalServiceStore{master: &BaseDB{DB: rawDB}}
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id FROM logical_service
		WHERE id = ? AND flag = 0 FOR UPDATE`)).
		WithArgs("logical-1").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("logical-1"))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT logical_service_id FROM service_environment_binding
		WHERE service_id = ? FOR UPDATE`)).
		WithArgs("service-1").
		WillReturnRows(sqlmock.NewRows([]string{"logical_service_id"}).AddRow("logical-2"))
	mock.ExpectRollback()

	err = store.UnbindServiceEnvironment("logical-1", "service-1", "revision-2")
	require.Equal(t, storeapi.DataConflictErr, storeapi.Code(err))
	require.NoError(t, mock.ExpectationsWereMet())
}
