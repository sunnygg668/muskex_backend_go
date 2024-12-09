/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"github.com/spf13/pflag"
	"log"
	"muskex/cmds/job/cmd"
	"muskex/utils"
	"os"
)

func main() {
	cmd.Execute()
}
func init() {
	log.SetFlags(log.LstdFlags)
	log.SetOutput(os.Stdout)
	printFlags := false
	var dbUrl string
	//--db 'admin:BMkOfAnChSqsC84BgdLJ@tcp(one-database-1.cvqm0oyo6ebo.ap-southeast-1.rds.amazonaws.com:3306)/kline?loc=Local&parseTime=true&multiStatements=true'  test2 -L 3306:one-database-1.cvqm0oyo6ebo.ap-southeast-1.rds.amazonaws.com:3306
	//pflag.StringVarP(&dbUrl, "db", "d", "admin:BMkOfAnChSqsC84BgdLJ@tcp(localhost:3306)/kline?loc=Local&parseTime=true&multiStatements=true", "mysql database url")
	pflag.StringVarP(&dbUrl, "db", "d", "admin:UhwZR0BApOSD57qATCao@tcp(musk-ex2-2024-11-155.cuyych8yxu1j.ap-southeast-1.rds.amazonaws.com:3306)/musk_test?loc=Local&parseTime=true&multiStatements=true", "mysql database url")
	pflag.BoolVarP(&printFlags, "print", "", false, "")
	pflag.Parse()
	if printFlags {
		log.Println("db:", dbUrl)
		os.Exit(0)
	}
	utils.InitDb(dbUrl)
}
