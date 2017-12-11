package whitelist

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"os"
)

/*
+---------+-------------+------+-----+---------+-------+
| Field   | Type        | Null | Key | Default | Extra |
+---------+-------------+------+-----+---------+-------+
| address | varchar(20) | NO   | PRI | NULL    |       |
| species | varchar(20) | NO   |     | NULL    |       |
+---------+-------------+------+-----+---------+-------+
*/

var (
	mysqldb    *sql.DB = nil
	tablename  string  = "whitelist"
	mysqlUName string  = "root"
	mysqlUPass string  = "root1234"
	dbName     string  = "nw_test"
)

func InitMyDB(addr, uname, upass, dbname, tbname string) {
	if mysqldb != nil {
		return
	}

	if len(addr) == 0 || len(uname) == 0 || len(upass) == 0 || len(dbname) == 0 {
		Logger.Error("params is null")
		os.Exit(1)
	}

	tablename = tbname
	mysqlUName = uname
	dbName = dbname

	//"root:root1234@tcp(127.0.0.1:3306)/nw_test?charset=utf8"
	connstr := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8", uname, upass, addr, dbname)
	Logger.Info("initmydb: test connect string :", connstr)

	db, err := sql.Open("mysql", connstr)
	if err != nil {
		Logger.Errorf("init mysql connect pool fail err: %T", err)
		os.Exit(1)
	}
	mysqldb = db

	mysqldb.SetMaxOpenConns(200)
	mysqldb.SetMaxIdleConns(100)
	err = mysqldb.Ping()
	if err != nil {
		Logger.Error("init mysql driver fail: ", err)
		os.Exit(1)
	}
}

func Inster(address, species string) error {

	Logger.Info("Inter start")
	stmt, err := mysqldb.Prepare("insert into whitelist(address,species)values(?,?)")
	defer stmt.Close()
	if err != nil {
		Logger.Error("inster: prepare fail")
		return err
	}

	_, err = stmt.Exec(address, species)
	if err != nil {
		Logger.Errorf("inster: inster data fail address: %S, species: %S ", address, species)
		return err
	}

	Logger.Info("Inter end")

	return nil

}
func Modify(address, species string) error {
	Logger.Info("Modify start")
	stmt, err := mysqldb.Prepare("update whitelist set species = ? where address = ?")
	defer stmt.Close()
	if err != nil {
		Logger.Error("modify: prepare fail")
		return err
	}

	_, err = stmt.Exec(species, address)
	if err != nil {
		Logger.Errorf("modify: modify fail address: %s, species: %s", address, species)
		return err
	}
	Logger.Info("Modify end")

	return nil
}
