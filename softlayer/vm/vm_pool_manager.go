package vm

import (
	"fmt"
	sql "database/sql"
	_ "github.com/mattn/go-sqlite3"
	"path/filepath"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"strings"
)

const (
	//SQLITE_DB_FOLDER = "/var/vcap/store/director/"
	SQLITE_DB_FOLDER = "/tmp/director/"
	SQLITE_DB_FOLDER_CLI = "/usr/local/"
	SQLITE_DB_FILE = "vm_pool.sqlite"
)

var (
	SQLITE_DB_FILE_PATH = filepath.Join(SQLITE_DB_FOLDER, SQLITE_DB_FILE)
)

type VMProperties struct {
	id int
	name string
	in_use string
	image_id string
	agent_id string
}

type VMInfoDB struct {
	vmProperties VMProperties
	dbConn *sql.DB
	logger boshlog.Logger
}


func NewVMInfoDB(id int, name string, in_use string, image_id string, agent_id string, logger boshlog.Logger) VMInfoDB {
	dbConn, _:= openDB()

	vmProperties := VMProperties{id, name, in_use, image_id, agent_id}
	return VMInfoDB{
		vmProperties: vmProperties,
	    dbConn: dbConn,
		logger: logger,
	}
}

func openDB() (*sql.DB, error) {

	db, err := sql.Open("sqlite3", SQLITE_DB_FILE_PATH)
	if err != nil {
		return nil, bosherr.WrapError(err, "Failed to open VM Pool DB")
	}

	return db, nil
}

func InitVMPoolDB() error {

	db, err := openDB()

	// Create vms table if it does not exist
	sqlStmt := `create table if not exists vms (id int not null primary key, name varchar(32), in_use varchar(32),
										  public_ip varchar(32), private_ip varchar(32), root_pwd varchar(32),
										  image_id varchar(64),
										  agent_id varchar(32),
										  timestamp timestamp)`
	err = exec(db, sqlStmt)
	if err != nil {
		return bosherr.WrapError(err, "Failed to execute sql statement: " + sqlStmt)
	}
	return nil
}

func exec(db *sql.DB, sqlStmt string) error {

	tx, err := db.Begin()
	if err != nil {
		return bosherr.WrapError(err, "Failed to begin DB transcation")
	}

	_, err = tx.Exec(sqlStmt)
	if err != nil {
		return bosherr.WrapError(err, "Failed to execute sql statement: " + sqlStmt)
	}

	tx.Commit()
	return nil
}

func (vmInfoDB *VMInfoDB) QueryVMInfobyAgentID() (error) {

	defer vmInfoDB.dbConn.Close()

	tx, err := vmInfoDB.dbConn.Begin()
	if err != nil {
		return bosherr.WrapError(err, "Failed to begin DB transcation")
	}

	sqlStmt, err := tx.Prepare("select id, image_id, agent_id from vms where agent_id=? and in_use='f'")
	if err != nil {
		return bosherr.WrapError(err, "Failed to prepare sql statement")
	}
	defer sqlStmt.Close()

	err = sqlStmt.QueryRow(vmInfoDB.vmProperties.agent_id).Scan(&vmInfoDB.vmProperties.id, &vmInfoDB.vmProperties.image_id, &vmInfoDB.vmProperties.agent_id)
	if err != nil && !strings.Contains(err.Error(), "no rows") {
		return bosherr.WrapError(err, "Failed to query VM info from vms table")
	}
	tx.Commit()

	return nil

}

/*func (vmInfo *VMInfo) queryVMInfobyNamePrefix(db *sql.DB) error {

	tx, err := db.Begin()
	if err != nil {
		return bosherr.WrapError(err, "Failed to begin DB transcation")
	}
	defer db.Close()

	sqlStmt, err := tx.Prepare("select id, image_id from vms where agent_id=? and in_use='f'")
	if err != nil {
		return bosherr.WrapError(err, "Failed to prepare sql statement")
	}
	defer sqlStmt.Close()

	err = sqlStmt.QueryRow(vmInfo.name).Scan(&vmInfo.id, &vmInfo.image_id)
	if err != nil {
		return bosherr.WrapError(err, "Failed to query VM info from vms table")
	}
	tx.Commit()

	return nil
}*/

func (vmInfoDB *VMInfoDB) DeleteVMFromVMDB() error {

	defer vmInfoDB.dbConn.Close()

	sqlStmt := fmt.Sprintf("delete from vms where id=%d", vmInfoDB.vmProperties.id)
	err := exec(vmInfoDB.dbConn, sqlStmt)
	if err != nil {
		return bosherr.WrapError(nil, "Failed to delete VM info from vms table")
	}

	return nil
}

func (vmInfoDB *VMInfoDB) InsertVMInfo() error {

	defer vmInfoDB.dbConn.Close()

	sqlStmt := fmt.Sprintf("insert into vms (id, name, in_use, image_id, agent_id, timestamp) values (%d, '%s', '%s', '%s', '%s', CURRENT_TIMESTAMP)", vmInfoDB.vmProperties.id, vmInfoDB.vmProperties.name, vmInfoDB.vmProperties.in_use, vmInfoDB.vmProperties.image_id, vmInfoDB.vmProperties.agent_id)
	err := exec(vmInfoDB.dbConn, sqlStmt)
	if err != nil {
		return bosherr.WrapError(err, "Failed to insert VM info into vms table")
	}

	return nil
}

func (vmInfoDB *VMInfoDB) UpdateVMInfo() error {

	defer vmInfoDB.dbConn.Close()

	tx, err := vmInfoDB.dbConn.Begin()
	if err != nil {
		return bosherr.WrapError(err, "Failed to begin DB transcation")
	}

	if vmInfoDB.vmProperties.in_use == "f" || vmInfoDB.vmProperties.in_use == "t" {
		sqlStmt := fmt.Sprintf("update vms set in_use='%s', timestamp=CURRENT_TIMESTAMP) where id = %s", vmInfoDB.vmProperties.in_use, vmInfoDB.vmProperties.id)
		_, err = tx.Exec(sqlStmt)
		if err != nil {
			return bosherr.WrapError(err, "Failed to begin DB transcation")
		}
	}
	if vmInfoDB.vmProperties.image_id != "" {
		sqlStmt := fmt.Sprintf("update vms set image_id where id = %s", vmInfoDB.vmProperties.image_id, vmInfoDB.vmProperties.id)
		_, err = tx.Exec(sqlStmt)
		if err != nil {
			return bosherr.WrapError(err, "Failed to begin DB transcation")
		}
	}
	if vmInfoDB.vmProperties.agent_id != "" {
		sqlStmt := fmt.Sprintf("update vms set agent_id where id = %s", vmInfoDB.vmProperties.agent_id, vmInfoDB.vmProperties.id)
		_, err = tx.Exec(sqlStmt)
		if err != nil {
			return bosherr.WrapError(err, "Failed to begin DB transcation")
		}
	}
	tx.Commit()

	return nil
}

/*func (vmInfoDB *VMInfoDB) ReleaseVMToPool(id int, agent_id string) error {

	vmInfo := &VMInfo{id, "", "f", "", agent_id}
	err := vmInfoDB.updateVMInfo()
	if err != nil {
		return bosherr.WrapError(err, "Failed to release VM to the pool")
	} else {
		return nil
	}
}*/


