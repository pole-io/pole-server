package sqldb

import (
	"database/sql"
	"errors"

	"github.com/pole-io/pole-server/apis/store"
)

func resolveAIBackendServiceID(
	tx *BaseTx, resourceNamespace, backendType, backendNamespace, backendName string,
) (string, error) {
	if backendType != "service" {
		return "", nil
	}
	if backendNamespace == "" || backendName == "" {
		return "", store.NewStatusError(
			store.EmptyParamsErr, "ai backend service missing namespace or name")
	}
	if resourceNamespace != backendNamespace {
		return "", store.NewStatusError(
			store.DataConflictErr, "ai backend service must be in the resource namespace")
	}

	var serviceID string
	err := tx.QueryRow(`SELECT id FROM service
		WHERE namespace = ? AND name = ? AND flag != 1 AND IFNULL(reference, '') = ''
		FOR UPDATE`, backendNamespace, backendName).Scan(&serviceID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", store.NewStatusError(
			store.NotFoundService, "ai backend service not found or is an alias")
	}
	if err != nil {
		return "", store.Error(err)
	}
	return serviceID, nil
}

func nullableAIBackendServiceID(serviceID string) any {
	if serviceID == "" {
		return nil
	}
	return serviceID
}
