package vm_pool

import (
	"fmt"
	"strings"

	sql "database/sql"
	_ "github.com/mattn/go-sqlite3"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type vmProperties struct {
	Id      int
	Name    string
	InUse   string
	ImageId string
	AgentId string
}

type VMInfoDB struct {
	dbConn *sql.DB
	logger boshlog.Logger

	VmProperties vmProperties
}

func NewVMInfoDB(id int, name string, in_use string, image_id string, agent_id string, logger boshlog.Logger, dbConn *sql.DB) VMInfoDB {
	return VMInfoDB{
		VmProperties: vmProperties{
			Id:      id,
			Name:    name,
			InUse:   in_use,
			ImageId: image_id,
			AgentId: agent_id},
		dbConn: dbConn,
		logger: logger,
	}
}

func (vmInfoDB *VMInfoDB) CloseDB() error {
	err := vmInfoDB.dbConn.Close()
	if err != nil {
		return bosherr.WrapError(err, "Failed to close VM Pool DB connection")
	}
	return nil
}

func (vmInfoDB *VMInfoDB) QueryVMInfobyAgentID() error {
	tx, err := vmInfoDB.dbConn.Begin()
	if err != nil {
		return bosherr.WrapError(err, "Failed to begin DB transcation")
	}

	var prepareStmt string
	if vmInfoDB.VmProperties.InUse == "t" {
		prepareStmt = "select id, image_id, agent_id from vms where in_use='t' and agent_id=?"
	} else if vmInfoDB.VmProperties.InUse == "f" {
		prepareStmt = "select id, image_id, agent_id from vms where in_use='f' and agent_id=?"
	} else {
		prepareStmt = "select id, image_id, agent_id from vms where agent_id==?"
	}

	sqlStmt, err := tx.Prepare(prepareStmt)
	defer sqlStmt.Close()
	if err != nil {
		return bosherr.WrapError(err, "Failed to prepare sql statement")
	}

	err = sqlStmt.QueryRow(vmInfoDB.VmProperties.AgentId).Scan(&vmInfoDB.VmProperties.Id, &vmInfoDB.VmProperties.ImageId, &vmInfoDB.VmProperties.AgentId)
	if err != nil && !strings.Contains(err.Error(), "no rows") {
		return bosherr.WrapError(err, "Failed to query VM info from vms table")
	}
	tx.Commit()

	return nil
}

func (vmInfoDB *VMInfoDB) QueryVMInfobyID() error {
	tx, err := vmInfoDB.dbConn.Begin()
	if err != nil {
		return bosherr.WrapError(err, "Failed to begin DB transcation")
	}

	var prepareStmt string
	if vmInfoDB.VmProperties.InUse == "t" {
		prepareStmt = "select id, in_use, image_id, agent_id from vms where id=? and in_use='t'"
	} else if vmInfoDB.VmProperties.InUse == "f" {
		prepareStmt = "select id, in_use, image_id, agent_id from vms where id=? and in_use='f'"
	} else {
		prepareStmt = "select id, in_use, image_id, agent_id from vms where id=?"
	}

	sqlStmt, err := tx.Prepare(prepareStmt)
	if err != nil {
		return bosherr.WrapError(err, "Failed to prepare sql statement")
	}
	defer sqlStmt.Close()

	err = sqlStmt.QueryRow(vmInfoDB.VmProperties.Id).Scan(&vmInfoDB.VmProperties.Id, &vmInfoDB.VmProperties.InUse, &vmInfoDB.VmProperties.ImageId, &vmInfoDB.VmProperties.AgentId)
	if err != nil && !strings.Contains(err.Error(), "no rows") {
		return bosherr.WrapError(err, "Failed to query VM info from vms table")
	}
	tx.Commit()

	return nil
}

func (vmInfoDB *VMInfoDB) DeleteVMFromVMDB() error {
	sqlStmt := fmt.Sprintf("delete from vms where id=%d", vmInfoDB.VmProperties.Id)
	err := exec(vmInfoDB.dbConn, sqlStmt)
	if err != nil {
		return bosherr.WrapError(nil, "Failed to delete VM info from vms table")
	}

	return nil
}

func (vmInfoDB *VMInfoDB) InsertVMInfo() error {
	sqlStmt := fmt.Sprintf("insert into vms (id, name, in_use, image_id, agent_id, timestamp) values (%d, '%s', '%s', '%s', '%s', CURRENT_TIMESTAMP)", vmInfoDB.VmProperties.Id, vmInfoDB.VmProperties.Name, vmInfoDB.VmProperties.InUse, vmInfoDB.VmProperties.ImageId, vmInfoDB.VmProperties.AgentId)
	err := exec(vmInfoDB.dbConn, sqlStmt)
	if err != nil {
		return bosherr.WrapError(err, "Failed to insert VM info into vms table")
	}

	return nil
}

func (vmInfoDB *VMInfoDB) UpdateVMInfoByID() error {
	tx, err := vmInfoDB.dbConn.Begin()
	if err != nil {
		return bosherr.WrapError(err, "Failed to begin DB transcation")
	}

	if vmInfoDB.VmProperties.InUse == "f" || vmInfoDB.VmProperties.InUse == "t" {
		sqlStmt := fmt.Sprintf("update vms set in_use='%s', timestamp=CURRENT_TIMESTAMP where id = %d", vmInfoDB.VmProperties.InUse, vmInfoDB.VmProperties.Id)
		_, err = tx.Exec(sqlStmt)
		if err != nil {
			return bosherr.WrapError(err, "Failed to update in_use column in vms")
		}
	}
	if vmInfoDB.VmProperties.ImageId != "" {
		sqlStmt := fmt.Sprintf("update vms set image_id='%s' where id = %d", vmInfoDB.VmProperties.ImageId, vmInfoDB.VmProperties.Id)
		_, err = tx.Exec(sqlStmt)
		if err != nil {
			return bosherr.WrapError(err, "Failed to update image_id column in vms")
		}
	}
	if vmInfoDB.VmProperties.AgentId != "" {
		sqlStmt := fmt.Sprintf("update vms set agent_id='%s' where id = %d", vmInfoDB.VmProperties.AgentId, vmInfoDB.VmProperties.Id)
		_, err = tx.Exec(sqlStmt)
		if err != nil {
			return bosherr.WrapError(err, "Failed to update agent_id column in vms")
		}
	}
	tx.Commit()

	return nil
}

// Private methods

func exec(db *sql.DB, sqlStmt string) error {
	tx, err := db.Begin()
	if err != nil {
		return bosherr.WrapError(err, "Failed to begin DB transcation")
	}

	_, err = tx.Exec(sqlStmt)
	if err != nil {
		return bosherr.WrapError(err, "Failed to execute sql statement: "+sqlStmt)
	}

	tx.Commit()
	return nil
}
