package sqldb

import (
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"

	apiservice "github.com/pole-io/specification/source/go/api/v1/service_manage"

	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
)

func TestCreateServiceContractUsesIdempotentUpsert(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()

	db := &BaseDB{DB: rawDB}
	storage := &serviceContractStore{master: db, slave: db}
	contract := &svctypes.ServiceContract{
		ID: "contract-1", Type: "protobuf", Namespace: "default", Service: "payments",
		Protocol: "grpc", Version: "v1", Revision: "revision-1", Content: "proto",
		ContentDigest: "digest", MetadataStr: "{}",
	}
	mock.ExpectExec("INSERT INTO service_contract.*ON DUPLICATE KEY UPDATE").
		WithArgs(
			"contract-1", "protobuf", "default", "payments", "grpc", "v1", "revision-1",
			"proto", "digest", "{}",
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	require.NoError(t, storage.CreateServiceContract(contract))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAddServiceContractInterfacesReplacesOnlySameSource(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()

	db := &BaseDB{DB: rawDB}
	storage := &serviceContractStore{master: db, slave: db}
	contract := &svctypes.EnrichServiceContract{
		ServiceContract: &svctypes.ServiceContract{ID: "contract-1", Revision: "revision-2"},
		InterfaceSource: apiservice.InterfaceDescriptor_Client,
		Interfaces: []*svctypes.InterfaceDescriptor{{
			ID: "interface-1", ContractID: "contract-1", Namespace: "default", Service: "payments",
			Protocol: "grpc", Version: "v1", Method: "Charge", Path: "payment.v1.PaymentService",
			Type: "protobuf", Revision: "interface-revision", Source: apiservice.InterfaceDescriptor_Client,
		}},
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE service_contract SET revision = ?, mtime = sysdate() WHERE id = ?")).
		WithArgs("revision-2", "contract-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM service_contract_detail WHERE contract_id = ? AND source IN (?, 0)")).
		WithArgs("contract-1", int(apiservice.InterfaceDescriptor_Client)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("REPLACE INTO service_contract_detail").
		WithArgs(
			"interface-1", "contract-1", "default", "payments", "grpc", "v1", "Charge",
			"payment.v1.PaymentService", "protobuf", "", "", "interface-revision", 0,
			int(apiservice.InterfaceDescriptor_Client),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	require.NoError(t, storage.AddServiceContractInterfaces(contract))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetServiceContractExcludesSoftDeletedInterfaces(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()

	db := &BaseDB{DB: rawDB}
	storage := &serviceContractStore{master: db, slave: db}
	mock.ExpectBegin()
	mock.ExpectQuery("FROM service_contract WHERE flag = 0 AND id = \\?").
		WithArgs("contract-1").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "type", "namespace", "service", "protocol", "version", "revision", "flag",
			"content", "content_digest", "metadata", "ctime", "mtime",
		}).AddRow(
			"contract-1", "protobuf", "default", "payments", "grpc", "v1", "revision-1", 0,
			"proto", "digest", "{}", int64(100), int64(200),
		))
	mock.ExpectQuery("WHERE contract_id = \\? AND flag = 0").
		WithArgs("contract-1").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "contract_id", "namespace", "service", "protocol", "version", "type", "method",
			"path", "content", "content_digest", "revision", "ctime", "mtime", "source",
		}))
	mock.ExpectRollback()

	contract, err := storage.GetServiceContract("contract-1")

	require.NoError(t, err)
	require.NotNil(t, contract)
	require.Empty(t, contract.Interfaces)
	require.NoError(t, mock.ExpectationsWereMet())
}
