package sqldb

import (
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetOrCreateServiceIdentityByTokenBackfillsLegacyService(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()

	store := &serviceStore{master: &BaseDB{DB: rawDB}}
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`select id, name, namespace from service
		where token = ? and flag = 0 and (reference is null or reference = '')
		limit 2 for update`)).
		WithArgs("service-token").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "namespace"}).
			AddRow("service-id", "orders", "default"))
	mock.ExpectExec(regexp.QuoteMeta(`insert ignore into service_identity
		(service_id, subject, revision, ctime, mtime)
		values (?, ?, ?, sysdate(), sysdate())`)).
		WithArgs("service-id", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(regexp.QuoteMeta(`select subject, revision from service_identity where service_id = ?`)).
		WithArgs("service-id").
		WillReturnRows(sqlmock.NewRows([]string{"subject", "revision"}).
			AddRow("pole://service/stable-id", "identity-revision"))
	mock.ExpectCommit()

	svc, err := store.GetOrCreateServiceIdentityByToken("service-token")

	require.NoError(t, err)
	require.NotNil(t, svc)
	require.NotNil(t, svc.Identity)
	assert.Equal(t, "orders", svc.Name)
	assert.Equal(t, "default", svc.Namespace)
	assert.Equal(t, "pole://service/stable-id", svc.Identity.Subject)
	assert.Equal(t, "identity-revision", svc.Identity.Revision)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetOrCreateServiceIdentityByTokenRejectsAmbiguousToken(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()

	store := &serviceStore{master: &BaseDB{DB: rawDB}}
	mock.ExpectBegin()
	mock.ExpectQuery("select id, name, namespace from service").
		WithArgs("duplicate-token").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "namespace"}).
			AddRow("service-a", "orders", "default").
			AddRow("service-b", "payments", "default"))
	mock.ExpectRollback()

	svc, err := store.GetOrCreateServiceIdentityByToken("duplicate-token")

	assert.Nil(t, svc)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "multiple services")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCleanServiceRemovesIdentityBeforeHardDelete(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()

	mock.ExpectBegin()
	tx, err := rawDB.Begin()
	require.NoError(t, err)
	baseTx := &BaseTx{Tx: tx}

	mock.ExpectExec("delete service_identity from service_identity").
		WithArgs("orders", "default").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("delete from service where name").
		WithArgs("orders", "default").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectRollback()

	require.NoError(t, cleanService(baseTx, "orders", "default"))
	require.NoError(t, baseTx.Rollback())
	require.NoError(t, mock.ExpectationsWereMet())
}
