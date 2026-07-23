package sqldb

import (
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"

	conftypes "github.com/pole-io/pole-server/apis/pkg/types/config"
	"github.com/pole-io/pole-server/apis/pkg/types/rules"
)

func TestCreateConfigFileGrayReleaseKeepsExistingGrayActive(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()

	db := &BaseDB{DB: rawDB}
	store := &configFileReleaseStore{}
	release := configFileReleaseForStoreTest("gray-a", conftypes.ReleaseTypeGray)

	mock.ExpectBegin()
	tx, err := db.Begin()
	require.NoError(t, err)
	storeTx := NewSqlDBTx(tx)

	mock.ExpectExec(regexp.QuoteMeta("SELECT id FROM config_file WHERE namespace = ? AND `group` = ? AND name = ? FOR UPDATE")).
		WithArgs(release.Namespace, release.Group, release.Name).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM config_file_release WHERE namespace = ? AND `group` = ? AND file_name = ? AND name = ? AND flag = 1")).
		WithArgs(release.Namespace, release.Group, release.FileName, release.Name).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT IFNULL(MAX(`version`), 0) FROM config_file_release WHERE namespace = ? AND  `group` = ? AND file_name = ?")).
		WithArgs(release.Namespace, release.Group, release.FileName).
		WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow(uint64(3)))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO config_file_release(name, namespace, `group`, file_name, content , comment, md5,  version, ctime, create_by , mtime, modify_by, active, tags, description, release_type)  VALUES (?, ?, ?, ?, ? , ?, ?, ?, sysdate(), ? , sysdate(), ?, 1, ?, ?, ?)")).
		WithArgs(release.Name, release.Namespace, release.Group, release.FileName, release.Content, release.Comment,
			release.Md5, uint64(4), release.CreateBy, release.ModifyBy, "{}", release.ReleaseDescription, release.ReleaseType).
		WillReturnResult(sqlmock.NewResult(1, 1))

	require.NoError(t, store.CreateConfigFileReleaseTx(storeTx, release))
	require.NoError(t, mock.ExpectationsWereMet())
}

func configFileReleaseForStoreTest(name string, releaseType rules.ReleaseType) *conftypes.ConfigFileRelease {
	return &conftypes.ConfigFileRelease{
		SimpleConfigFileRelease: &conftypes.SimpleConfigFileRelease{
			ConfigFileReleaseKey: &conftypes.ConfigFileReleaseKey{
				Name:        name,
				Namespace:   "default",
				Group:       "group-a",
				FileName:    "app.yaml",
				ReleaseType: releaseType,
			},
			Comment:            "comment",
			Md5:                "md5",
			Metadata:           map[string]string{},
			CreateBy:           "tester",
			ModifyBy:           "tester",
			ReleaseDescription: "desc",
		},
		Content: "content",
	}
}
