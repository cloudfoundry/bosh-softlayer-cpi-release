package vm

import (
	"fmt"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"path/filepath"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	//sl "github.com/maximilien/softlayer-go/softlayer"
)

const (
	SQLITE_DB_FOLDER = "/var/vcap/store/director/"
	SQLITE_DB_FOLDER_CLI = "/usr/local/"
	SQLITE_DB_FILE = "vm_pool.sqlite"
)

var (
	SQLITE_DB_FILE_PATH = filepath.Join(SQLITE_DB_FOLDER, SQLITE_DB_FILE)
)

type VMInfo struct {
	id int64
	name string
	in_use bool
	image_id string
	agent_id string
}

func OpenDB() (*sql.DB, error) {

	db, err := sql.Open("sqlite3", SQLITE_DB_FILE_PATH)

	if err != nil {
		return nil, bosherr.WrapError(err, "Failed to open VM DB")
	}

	return db, nil
}

func InitVMDB(db *sql.DB) error {

	// Create vms table if it does not exist
	sqlStmt := `create table if not exists vms (id int not null primary key, name varchar(32), in_use varchar(32),
										  public_ip varchar(32), private_ip varchar(32), root_pwd varchar(32),
										  image_id varchar(64),
										  agent_id varchar(32),
										  timestamp timestamp)`
	err := exec(db, sqlStmt)
	if err != nil {
		return bosherr.WrapError(err, "Failed to execute sql statement: " + sqlStmt)
	}
	return err
}

func exec(db *sql.DB, sqlStmt string) error {

	tx, err := db.Begin()
	if err != nil {
		return bosherr.WrapError(err, "Failed to begin DB transcation")
	}
	defer db.Close()

	_, err = tx.Exec(sqlStmt)
	if err != nil {
		return bosherr.WrapError(err, "Failed to execute sql statement: " + sqlStmt)
	}

	tx.Commit()
	return nil
}

func (vmInfo *VMInfo) QueryVMInfobyAgentID(db *sql.DB) error {

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

	err = sqlStmt.QueryRow(vmInfo.agent_id).Scan(&vmInfo.id, &vmInfo.image_id)
	if err != nil {
		return bosherr.WrapError(err, "Failed to query VM info from vms table")
	}
	tx.Commit()

	return nil

}

func (vmInfo *VMInfo) QueryVMInfobyNamePrefix(db *sql.DB) error {

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
}

func (vmInfo *VMInfo) DeleteVMFromDB(db *sql.DB) error {

	sqlStmt := fmt.Sprintf("delete from vms where id=%d", vmInfo.id)
	err := exec(db, sqlStmt)
	if err != nil {
		return bosherr.WrapError(nil, "Failed to delete VM info from vms table")
	}

	return nil
}

func (vmInfo *VMInfo) InsertVMInfo(db *sql.DB) error {

	sqlStmt := fmt.Sprintf("insert into vms (id, name, in_use, image_id, agent_id, timestamp) values (%d, %s, %s, %s, CURRENT_TIMESTAMP)", vmInfo.id, vmInfo.name, vmInfo.in_use, vmInfo.image_id, vmInfo.agent_id)
	err := exec(db, sqlStmt)
	if err != nil {
		return bosherr.WrapError(nil, "Failed to insert VM info into vms table")
	}

	return nil
}

func (vmInfo *VMInfo) UpdateVMInfo(db *sql.DB) error {

	tx, err := db.Begin()
	if err != nil {
		return bosherr.WrapError(err, "Failed to begin DB transcation")
	}
	defer db.Close()

	if vmInfo.in_use == "f" || vmInfo.in_use == "t" {
		sqlStmt := fmt.Sprintf("update vms set in_use='%s', timestamp=CURRENT_TIMESTAMP) where id = %s", vmInfo.in_use, vmInfo.id)
		_, err = tx.Exec(sqlStmt)
		if err != nil {
			return bosherr.WrapError(err, "Failed to begin DB transcation")
		}
	}
	if vmInfo.image_id != "" {
		sqlStmt := fmt.Sprintf("update vms set image_id where id = %s", vmInfo.image_id, vmInfo.id)
		_, err = tx.Exec(sqlStmt)
		if err != nil {
			return bosherr.WrapError(err, "Failed to begin DB transcation")
		}
	}
	if vmInfo.agent_id != "" {
		sqlStmt := fmt.Sprintf("update vms set agent_id where id = %s", vmInfo.agent_id, vmInfo.id)
		_, err = tx.Exec(sqlStmt)
		if err != nil {
			return bosherr.WrapError(err, "Failed to begin DB transcation")
		}
	}
	tx.Commit()

	return nil
}
