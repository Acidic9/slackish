package main

import (
	"fmt"
	"log"

	"database/sql"
	"regexp"

	_ "github.com/go-sql-driver/mysql"
)

const salt = "o*bF$S!49ohV@^%dO*Ib2s95!s32b3PD"

var (
	db                      *sql.DB
	users                   = make(map[int]*user)
	slackSearchRequirements = regexp.MustCompile(`[a-zA-Z0-9_]{2,}`)
)

func init() {
	log.SetFlags(log.Lshortfile)
	var err error
	db, err = sql.Open("mysql",
		fmt.Sprintf("%s:%s@tcp(%s)/%s",
			config.Database.Username,
			config.Database.Password,
			config.Database.Address,
			config.Database.Name))
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("Count not connect to the database")
	}
}
