package vm_pool

import (
	"fmt"
	"strings"
	"time"

	"database/sql"
	"database/sql/driver"

	sqlite3 "github.com/mattn/go-sqlite3"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type DB interface {
	Begin() (*sql.Tx, error)
	Close() error
	Driver() driver.Driver
	Exec(query string, args ...interface{}) (sql.Result, error)
	Ping() error
	Prepare(query string) (*sql.Stmt, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	SetMaxIdleConns(n int)
	SetMaxOpenConns(n int)
}

type vmProperties struct {
	Id      int
	Name    string
	InUse   string
	ImageId string
	AgentId string
}

type VMInfoDB struct {
	db     DB
	logger boshlog.Logger

	VmProperties vmProperties
}

func NewVMInfoDB(id int, name string, in_use string, image_id string, agent_id string, logger boshlog.Logger, db DB) VMInfoDB {
	return VMInfoDB{
		VmProperties: vmProperties{
			Id:      id,
			Name:    name,
			InUse:   in_use,
			ImageId: image_id,
			AgentId: agent_id},
		db:     db,
		logger: logger,
	}
}

func (vmInfoDB *VMInfoDB) CloseDB() error {
	err := vmInfoDB.db.Close()
	if err != nil {
		return bosherr.WrapError(err, "Failed to close VM Pool DB connection")
	}
	return nil
}

func (vmInfoDB *VMInfoDB) QueryVMInfobyAgentID(retryTimes int, retryInterval time.Duration) error {
	locked := true
	retry_counter := 0
	var err error
	for locked && retry_counter <= retryTimes {
		retry_counter +=1
		tx, err := vmInfoDB.db.Begin()
		if err != nil {
			sqliteErr := err.(sqlite3.Error)
			if sqliteErr.Code == sqlite3.ErrBusy || sqliteErr.Code == sqlite3.ErrLocked {
				vmInfoDB.logger.Info(updateVMPoolDBLogTag, fmt.Sprintf("%s, retrying %d times...", sqliteErr.Error(), retry_counter))
				continue
			} else {
				return bosherr.WrapError(sqliteErr.Error(), "Failed to begin DB transcation")
			}
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
			sqliteErr := err.(sqlite3.Error)
			if sqliteErr.Code == sqlite3.ErrBusy || sqliteErr.Code ==  sqlite3.ErrLocked {
				vmInfoDB.logger.Info(updateVMPoolDBLogTag, fmt.Sprintf("%s, retrying %d times...", sqliteErr.Error(), retry_counter))
				continue
			} else {
				return bosherr.WrapError(sqliteErr.Error(), "Failed to prepare sql statement")
			}
		}

		err = sqlStmt.QueryRow(vmInfoDB.VmProperties.AgentId).Scan(&vmInfoDB.VmProperties.Id, &vmInfoDB.VmProperties.ImageId, &vmInfoDB.VmProperties.AgentId)
		if err != nil && !strings.Contains(err.Error(), "no rows") {
			sqliteErr := err.(sqlite3.Error)
			if sqliteErr.Code == sqlite3.ErrBusy || sqliteErr.Code == 6 {
				vmInfoDB.logger.Info(updateVMPoolDBLogTag, fmt.Sprintf("%s, retrying %d times...", sqliteErr.Error(), retry_counter))
				continue
			} else {
				return bosherr.WrapError(sqliteErr.Error(), "Failed to query VM info from vms table")
			}

		}
		tx.Commit()
		locked = false
	}

	if locked == true {
		sqliteErr := err.(sqlite3.Error)
		return bosherr.WrapError(sqliteErr.Error(), "Failed to query VM Pool DB")
	} else {
		return nil
	}

}

func (vmInfoDB *VMInfoDB) QueryVMInfobyID(retryTimes int, retryInterval time.Duration) error {
	locked := true
	retry_counter := 0
	var err error
	for locked && retry_counter <= retryTimes {
		retry_counter +=1
		tx, err := vmInfoDB.db.Begin()
		if err != nil {
			sqliteErr := err.(sqlite3.Error)
			if sqliteErr.Code == sqlite3.ErrBusy || sqliteErr.Code == sqlite3.ErrLocked {
				vmInfoDB.logger.Info(updateVMPoolDBLogTag, fmt.Sprintf("%s, retrying %d times...", sqliteErr.Error(), retry_counter))
				continue
			} else {
				return bosherr.WrapError(sqliteErr.Error(), "Failed to begin DB transcation")
			}
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
			sqliteErr := err.(sqlite3.Error)
			if sqliteErr.Code == sqlite3.ErrBusy || sqliteErr.Code == sqlite3.ErrLocked {
				vmInfoDB.logger.Info(updateVMPoolDBLogTag, fmt.Sprintf("%s, retrying %d times...", sqliteErr.Error(), retry_counter))
				continue
			} else {
				return bosherr.WrapError(sqliteErr.Error(), "Failed to prepare sql statement")
			}
		}
		defer sqlStmt.Close()

		err = sqlStmt.QueryRow(vmInfoDB.VmProperties.Id).Scan(&vmInfoDB.VmProperties.Id, &vmInfoDB.VmProperties.InUse, &vmInfoDB.VmProperties.ImageId, &vmInfoDB.VmProperties.AgentId)
		if err != nil && !strings.Contains(err.Error(), "no rows") {
			sqliteErr := err.(sqlite3.Error)
			if sqliteErr.Code == sqlite3.ErrBusy || sqliteErr.Code == sqlite3.ErrLocked {
				vmInfoDB.logger.Info(updateVMPoolDBLogTag, fmt.Sprintf("%s, retrying %d times...", sqliteErr.Error(), retry_counter))
				continue
			} else {
				return bosherr.WrapError(sqliteErr.Error(), "Failed to query VM info from vms table")
			}

		}
		tx.Commit()
		locked = false
	}

	if locked == true {
		sqliteErr := err.(sqlite3.Error)
		return bosherr.WrapError(sqliteErr.Error(), "Failed to execute Query VM info")
	} else {
		return nil
	}

}

func (vmInfoDB *VMInfoDB) DeleteVMFromVMDB(retryTimes int, retryInterval time.Duration) error {
	sqlStmt := fmt.Sprintf("delete from vms where id=%d", vmInfoDB.VmProperties.Id)
	err := exec(vmInfoDB.db, sqlStmt, DB_RETRY_TIMES, DB_RETRY_INTERVAL)
	if err != nil {
		return bosherr.WrapError(err, "Failed to delete VM info from vms table")
	}
	return nil
}

func (vmInfoDB *VMInfoDB) InsertVMInfo(retryTimes int, retryInterval time.Duration) error {
	sqlStmt := fmt.Sprintf("insert into vms (id, name, in_use, image_id, agent_id, timestamp) values (%d, '%s', '%s', '%s', '%s', CURRENT_TIMESTAMP)", vmInfoDB.VmProperties.Id, vmInfoDB.VmProperties.Name, vmInfoDB.VmProperties.InUse, vmInfoDB.VmProperties.ImageId, vmInfoDB.VmProperties.AgentId)
	err := exec(vmInfoDB.db, sqlStmt, DB_RETRY_TIMES, DB_RETRY_INTERVAL)
	if err != nil {
		return bosherr.WrapError(err, "Failed to insert VM info into vms table")
	}

	return nil
}

func (vmInfoDB *VMInfoDB) UpdateVMInfoByID(retryTimes int, retryInterval time.Duration) error {
	locked := true
	retry_counter := 0
	var err error
	for locked && retry_counter <= retryTimes {
		retry_counter +=1
		tx, err := vmInfoDB.db.Begin()
		if err != nil {
			sqliteErr := err.(sqlite3.Error)
			if sqliteErr.Code == sqlite3.ErrBusy || sqliteErr.Code == sqlite3.ErrLocked {
				vmInfoDB.logger.Info(updateVMPoolDBLogTag, fmt.Sprintf("%s, retrying %d times...", sqliteErr.Error(), retry_counter))
				continue
			} else {
				return bosherr.WrapError(sqliteErr.Error(), "Failed to begin DB transcation")
			}
		}

		if vmInfoDB.VmProperties.InUse == "f" || vmInfoDB.VmProperties.InUse == "t" {
			sqlStmt := fmt.Sprintf("update vms set in_use='%s', timestamp=CURRENT_TIMESTAMP where id = %d", vmInfoDB.VmProperties.InUse, vmInfoDB.VmProperties.Id)
			_, err = tx.Exec(sqlStmt)
			if err != nil {
				sqliteErr := err.(sqlite3.Error)
				if sqliteErr.Code == sqlite3.ErrBusy || sqliteErr.Code == sqlite3.ErrLocked {
					vmInfoDB.logger.Info(updateVMPoolDBLogTag, fmt.Sprintf("%s, retrying %d times...", sqliteErr.Error(), retry_counter))
					continue
				} else {
					return bosherr.WrapError(sqliteErr.Error(), "Failed to update in_use column in vms")
				}
			}
		}

		if vmInfoDB.VmProperties.ImageId != "" {
			sqlStmt := fmt.Sprintf("update vms set image_id='%s' where id = %d", vmInfoDB.VmProperties.ImageId, vmInfoDB.VmProperties.Id)
			_, err = tx.Exec(sqlStmt)
			if err != nil {
				sqliteErr := err.(sqlite3.Error)
				if sqliteErr.Code == sqlite3.ErrBusy || sqliteErr.Code == sqlite3.ErrLocked {
					vmInfoDB.logger.Info(updateVMPoolDBLogTag, fmt.Sprintf("%s, retrying %d times...", sqliteErr.Error(), retry_counter))
					continue
				} else {
					return bosherr.WrapError(sqliteErr.Error(), "Failed to update in_use column in vms")
				}
			}
		}

		if vmInfoDB.VmProperties.AgentId != "" {
			sqlStmt := fmt.Sprintf("update vms set agent_id='%s' where id = %d", vmInfoDB.VmProperties.AgentId, vmInfoDB.VmProperties.Id)
			_, err = tx.Exec(sqlStmt)
			if err != nil {
				sqliteErr := err.(sqlite3.Error)
				if sqliteErr.Code == sqlite3.ErrBusy || sqliteErr.Code == sqlite3.ErrLocked {
					vmInfoDB.logger.Info(updateVMPoolDBLogTag, fmt.Sprintf("%s, retrying %d times...", sqliteErr.Error(), retry_counter))
					continue
				} else {
					return bosherr.WrapError(sqliteErr.Error(), "Failed to update in_use column in vms")
				}
			}
		}
		tx.Commit()
		locked = false
	}

	if locked == true {
		sqliteErr := err.(sqlite3.Error)
		return bosherr.WrapError(sqliteErr.Error(), "Failed to update VM Pool DB")
	} else {
		return nil
	}
}

// Private methods

func exec(db DB, sqlStmt string, retryTimes int, retryInterval time.Duration) error {
	locked := true
	retry_counter := 0
	var err error
	for locked && retry_counter <= retryTimes {
		retry_counter +=1
		tx, err := db.Begin()
		if err != nil {
			sqliteErr := err.(sqlite3.Error)
			if sqliteErr.Code == sqlite3.ErrBusy || sqliteErr.Code == sqlite3.ErrLocked {
				boshlog.Logger.Info(updateVMPoolDBLogTag, fmt.Sprintf("%s, retrying %d times...", sqliteErr.Error(), retry_counter))
				continue
			} else {
				return bosherr.WrapError(sqliteErr.Error(), "Failed to begin DB transcation")
			}
		}

		_, err = tx.Exec(sqlStmt)
		if err != nil {
			sqliteErr := err.(sqlite3.Error)
			if sqliteErr.Code == sqlite3.ErrBusy || sqliteErr.Code == sqlite3.ErrLocked {
				boshlog.Logger.Info(updateVMPoolDBLogTag, fmt.Sprintf("%s, retrying %d times...", sqliteErr.Error(), retry_counter))
				continue
			} else {
				return bosherr.WrapError(sqliteErr.Error(), "Failed to execute sql statement: "+sqlStmt)
			}
		}

		tx.Commit()
		locked = false
	}

	if locked == true {
		sqliteErr := err.(sqlite3.Error)
		return bosherr.WrapError(sqliteErr.Error(), "Failed to execute sql statement: "+sqlStmt)
	} else {
		return nil
	}

}
