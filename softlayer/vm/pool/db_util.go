package vm_pool

import (
	"database/sql"
	"os"
	"path/filepath"
	"time"

	common "github.com/cloudfoundry/bosh-softlayer-cpi/common"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	_ "github.com/mattn/go-sqlite3"
)

var (
	SQLITE_DB_FOLDER    string
	SQLITE_DB_FILE      string
	SQLITE_DB_FILE_PATH string
	DB_RETRY_INTERVAL   = 5 * time.Second
	DB_RETRY_TIMEOUT    = 300 * time.Second
)

func InitVMPoolDB(retryTimeout time.Duration, retryInterval time.Duration, logger boshlog.Logger) error {

	SQLITE_DB_FOLDER = common.GetOSEnvVariable("SQLITE_DB_FOLDER", "/var/vcap/store/director/")
	SQLITE_DB_FILE = common.GetOSEnvVariable("SQLITE_DB_FILE", "vm_pool.sqlite")
	SQLITE_DB_FILE_PATH = filepath.Join(SQLITE_DB_FOLDER, SQLITE_DB_FILE)

	err := os.MkdirAll(SQLITE_DB_FOLDER, os.ModePerm)
	if err != nil {
		return bosherr.WrapError(err, "Failed to make directory: "+SQLITE_DB_FOLDER)
	}

	db, err := OpenDB(SQLITE_DB_FILE_PATH)
	defer db.Close()
	if err != nil {
		return bosherr.WrapError(err, "Opening DB")
	}

	sqlStmt := `CREATE TABLE IF NOT EXISTS vms (id int not null primary key, name varchar(32), in_use varchar(32),
										  public_ip varchar(32), private_ip varchar(32), root_pwd varchar(32),
										  image_id varchar(64),
										  agent_id varchar(32),
										  timestamp timestamp)`
	err = exec(db, sqlStmt, retryTimeout, retryInterval, logger)
	if err != nil {
		return bosherr.WrapError(err, "Failed to execute sql statement: "+sqlStmt)
	}

	return nil
}

func OpenDB(dbPath string) (*sql.DB, error) {

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, bosherr.WrapError(err, "Failed to open VM Pool DB")
	}

	return db, nil
}
