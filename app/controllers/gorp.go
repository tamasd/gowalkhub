package controllers

import (
	"database/sql"
	"fmt"

	"github.com/coopernurse/gorp"
	_ "github.com/go-sql-driver/mysql"
	"github.com/revel/revel"
	"github.com/tamasd/gowalkhub/app/models"
)

var DBM *gorp.DbMap

func InitDB() {
	db, dberr := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4",
		revel.Config.StringDefault("db.user", ""),
		revel.Config.StringDefault("db.pass", ""),
		revel.Config.StringDefault("db.host", "localhost:3306"),
		revel.Config.StringDefault("db.name", ""),
	))
	if dberr != nil {
		revel.ERROR.Print(dberr)
		return
	}

	db.Exec("SET @@global.innodb_large_prefix = 1;")

	dialect := &gorp.MySQLDialect{
		Engine:   "InnoDB",
		Encoding: "utf8mb4",
	}

	DBM = &gorp.DbMap{
		Db:      db,
		Dialect: dialect,
	}

	revel.INFO.Println("Database initialized")

	userTable := DBM.AddTable(models.User{})
	userTable.SetKeys(false, "Id")
	userTable.ColMap("Id").SetMaxSize(36)
	userTable.ColMap("Email").SetMaxSize(100).SetUnique(true)

	walkthroughTable := DBM.AddTable(models.Walkthrough{})
	walkthroughTable.SetKeys(false, "UUID")
	walkthroughTable.ColMap("UUID").SetMaxSize(36)
	walkthroughTable.ColMap("Description").SetMaxSize(65536 / 4)

	stepTable := DBM.AddTable(models.Step{})
	stepTable.SetKeys(false, "UUID")
	stepTable.ColMap("UUID").SetMaxSize(36)
	stepTable.ColMap("Description").SetMaxSize(65536 / 4)
	stepTable.ColMap("DescriptionRaw").SetMaxSize(65536 / 4)

	if err := DBM.CreateTablesIfNotExists(); err != nil {
		revel.ERROR.Fatal(err)
	}

	models.ResetWalkthroughMainPage(DBM)
}

type GorpController struct {
	*revel.Controller
	Txn *gorp.Transaction
}

func (c *GorpController) Begin() revel.Result {
	txn, err := DBM.Begin()
	if err != nil {
		panic(err)
	}

	c.Txn = txn

	return nil
}

func (c *GorpController) Commit() revel.Result {
	if c.Txn == nil {
		return nil
	}

	if err := c.Txn.Commit(); err != nil && err != sql.ErrTxDone {
		panic(err)
	}

	c.Txn = nil
	return nil
}

func (c *GorpController) Rollback() revel.Result {
	if c.Txn == nil {
		return nil
	}

	if err := c.Txn.Rollback(); err != nil && err != sql.ErrTxDone {
		panic(err)
	}

	c.Txn = nil
	return nil
}
