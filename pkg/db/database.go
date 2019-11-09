package db

import (
	"database/sql"
	"log"
)

type Database interface {
	Connect() Database
	Transaction(func(*sql.Tx) error) error
	Get() *sql.DB
	Close()
}

func NewClient(driver, query string) Database {
	return &Client{
		driver: driver,
		query:  query,
	}
}

type Client struct {
	db     *sql.DB
	driver string
	query  string
}

func (d *Client) Connect() Database {
	var err error
	d.db, err = sql.Open(d.driver, d.query)
	if err != nil {
		log.Fatal("Failed connection to database ("+d.driver+":"+d.query+"): ", err)
	}
	return d
}

func (d *Client) Transaction(txFunc func(*sql.Tx) error) (err error) {
	tx, err := d.db.Begin()
	if err != nil {
		return
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()
	err = txFunc(tx)
	return err
}

func (d *Client) Get() *sql.DB {
	return d.db
}

func (d *Client) Close() {
	_ = d.db.Close()
}
