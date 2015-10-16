package vm_pool

import (
	"database/sql"
	"os"
	"path/filepath"
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	common "github.com/maximilien/bosh-softlayer-cpi/common"
)

var (
	SQLITE_DB_FOLDER    = common.GetOSEnvVariable("SQLITE_DB_FOLDER", "/var/vcap/store/director/")
	SQLITE_DB_FILE      = common.GetOSEnvVariable("SQLITE_DB_FILE", "vm_pool.sqlite")
	SQLITE_DB_FILE_PATH = filepath.Join(SQLITE_DB_FOLDER, SQLITE_DB_FILE)
	DB_RETRY_INTERVAL = 3 * time.Second
	DB_RETRY_TIMES = 10
	updateVMPoolDBLogTag = "updateVMPoolDBLogTag"
)

func InitVMPoolDB() error {
	err := os.MkdirAll(SQLITE_DB_FOLDER, 0777)
	if err != nil {
		return bosherr.WrapError(err, "Failed to make director: "+SQLITE_DB_FOLDER)
	}

	db, err := OpenDB(SQLITE_DB_FILE_PATH)
	defer db.Close()
	if err != nil {
		return bosherr.WrapError(err, "Opening DB")
	}

	sqlStmt := `create table if not exists vms (id int not null primary key, name varchar(32), in_use varchar(32),
										  public_ip varchar(32), private_ip varchar(32), root_pwd varchar(32),
										  image_id varchar(64),
										  agent_id varchar(32),
										  timestamp timestamp)`
	err = exec(db, sqlStmt, DB_RETRY_TIMES, DB_RETRY_INTERVAL)
	if err != nil {
		return bosherr.WrapError(err, "Failed to execute sql statement: "+sqlStmt)
	}

	return nil
}

func OpenDB(dbPath string) (*sql.DB, error) {
	_, err := IsDirectory(dbPath)
	if err != nil {
		return nil, bosherr.WrapError(err, "Failed to open VM Pool DB, invalid DB path")
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, bosherr.WrapError(err, "Failed to open VM Pool DB")
	}

	return db, nil
}

func IsDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	return fileInfo.IsDir(), err
}
