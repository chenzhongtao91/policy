package metadata

import (
	"database/sql"
	"fmt"
	"strconv"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	EVENT_DB_NAME = "metadata_pending_events.db"

	EVENT_UPDATE_DEVICE_STATUS = "UPDATEDEVICESTATUS"
	EVENT_ADD_HOST_DEVICE      = "ADDHOSTDEVICE"
	EVENT_DEL_HOST_DEVICE      = "DELHOSTDEVICE"
)

const (
	createEventTable = `
       CREATE TABLE IF NOT EXISTS pending_events (
            id INTEGER PRIMARY KEY AUTOINCREMENT, 
           evType CHAR(24) NOT NULL,   
           optime INT(11) NOT NULL, 
           evStr BLOB NOT NULL
        );
       `
)

type EventPrototype interface {
	Name() string
	Value() ([]byte, error)
	SetName(string)
	SetValue([]byte) error
}

type EventFunc func(event *EventPrototype) error

// Database is a graph database for storing entities and their relationships.
type Database struct {
	conn *sql.DB
	mux  sync.RWMutex
}

type PendingSet struct {
	mutex   sync.Mutex
	DB      *Database
	Channel chan string          // event type(bucket) : channel
	EvFuncs map[string]EventFunc // event type(bucket)  : func
}

var (
	PendingOps *PendingSet
)

func NewSqliteConn(dbname string) (*Database, error) {
	conn, err := sql.Open("sqlite3", dbname)
	if err != nil {
		return nil, err
	}
	return NewDatabase(conn)
}

func NewDatabase(conn *sql.DB) (*Database, error) {
	if conn == nil {
		return nil, fmt.Errorf("Database connection cannot be nil")
	}
	db := &Database{conn: conn}

	// Create root entities
	tx, err := conn.Begin()
	if err != nil {
		return nil, err
	}

	if _, err := tx.Exec(createEventTable); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return db, nil
}

// Close the underlying connection to the database.
func (db *Database) Close() error {
	return db.conn.Close()
}

func PendingSetSetup() *PendingSet {

	PendingOps = new(PendingSet)
	var err error
	PendingOps.DB, err = NewSqliteConn("metadata_pending_evnets.db")
	if err != nil {
		panic(err.Error())
	}

	return PendingOps
}

func (ps *PendingSet) Add(evType string, ev EventPrototype) error {
	ps.DB.mux.Lock()
	defer ps.DB.mux.Unlock()

	var sqlStmt string
	t := strconv.FormatInt(time.Now().Unix(), 10)
	v, err := ev.Value()
	if err != nil {
		return err
	}
	fmt.Sprintf(sqlStmt, "INSERT INTO pending_events(evType, optime, evStr) values(%s,%d,%s);", evType, t, v)

	_, err = ps.DB.conn.Exec(sqlStmt)
	if err != nil {
		return err
	}

	ps.Channel <- evType
	return nil
}

func (ps *PendingSet) Delete(id int) error {
	ps.DB.mux.Lock()
	defer ps.DB.mux.Unlock()

	var sqlStmt string
	fmt.Sprintf(sqlStmt, "DELETE FROM pending_events WHERE id = %d", id)

	_, err := ps.DB.conn.Exec(sqlStmt)
	if err != nil {
		return err
	}

	return nil
}

func (ps *PendingSet) MetadataUpdater() {
	fmt.Println("MetadataMaintainer start")

	for {
		evType := <-ps.Channel

		var sqlStmt string
		fmt.Sprint(sqlStmt, "SELECT * FROM pending_events WHERE evType = %s ORDER BY optime;", evType)
		ps.mutex.Lock()
		rows, err := ps.DB.conn.Query(sqlStmt)
		if err != nil {
			continue
		}

		for rows.Next() {
			var id int
			var et string
			var optime int
			var value string

			rows.Scan(&id, &et, &optime, &value)

			var err error
			switch evType {
			case EVENT_UPDATE_DEVICE_STATUS:
				err = ExecuteUpdateDeviceStatus([]byte(value))
			case EVENT_ADD_HOST_DEVICE:
				err = ExecuteAddHostDevices([]byte(value))
			case EVENT_DEL_HOST_DEVICE:
				err = ExecuteDelHostDevices([]byte(value))
			}

			if err == nil {
				ps.Delete(id)
				ps.Channel <- evType //TODO not safe  if etcd failover， add sleep
				break
			}
		}
		rows.Close()

	}

	fmt.Println("MetadataMaintainer stop.")
}
