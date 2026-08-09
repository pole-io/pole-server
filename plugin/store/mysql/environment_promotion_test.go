package sqldb

import (
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"

	"github.com/pole-io/pole-server/apis/pkg/types"
	"github.com/pole-io/pole-server/apis/store"
)

func TestSaveEnvironmentPromotionTopologyCreatesFirstDraftRevision(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()
	repository := &environmentPromotionStore{master: &BaseDB{DB: rawDB}}
	topology := &types.EnvironmentPromotionTopology{
		Edges: []types.EnvironmentPromotionEdge{{ID: "dev-tst", Source: "dev", Target: "tst"}},
	}
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT draft_revision FROM environment_promotion_topology WHERE id = 1 FOR UPDATE`)).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectExec("INSERT INTO environment_promotion_topology").
		WithArgs(sqlmock.AnyArg(), "").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	require.NoError(t, repository.SaveEnvironmentPromotionTopology(topology, 0))
	require.Equal(t, uint64(1), topology.DraftRevision)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSaveEnvironmentPromotionTopologyRejectsStaleDraftRevision(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()
	repository := &environmentPromotionStore{master: &BaseDB{DB: rawDB}}
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT draft_revision FROM environment_promotion_topology WHERE id = 1 FOR UPDATE`)).
		WillReturnRows(sqlmock.NewRows([]string{"draft_revision"}).AddRow(3))
	mock.ExpectRollback()

	err = repository.SaveEnvironmentPromotionTopology(&types.EnvironmentPromotionTopology{}, 2)

	require.ErrorIs(t, err, store.ErrEnvironmentPromotionRevisionConflict)
	require.NoError(t, mock.ExpectationsWereMet())
}
