package mysql

import (
	"errors"
	"testing"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
)

func TestIsMissingTableError(t *testing.T) {
	require.True(t, isMissingTableError(&mysqlDriver.MySQLError{Number: 1146, Message: "table does not exist"}))
	require.False(t, isMissingTableError(&mysqlDriver.MySQLError{Number: 1064, Message: "syntax error"}))
	require.False(t, isMissingTableError(errors.New("other error")))
}
