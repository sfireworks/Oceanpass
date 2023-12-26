// ///////////////////////////////////////
// 2022 SHAILab Storage all rights reserved
// Author: Shiqian Yan
// ///////////////////////////////////////
package config

import (
	"database/sql"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type DBConfig struct {
	XMLName xml.Name `xml:"db_config"`
	DbBase  DbBase   `xml:"db_base"`
}

type DbBase struct {
	DBType     string `xml:"db_type"`
	Username   string `xml:"username"`
	Password   string `xml:"password"`
	IPProtocol string `xml:"ip_protocol"`
	DBName     string `xml:"db_name"`
	IPAddress  string `xml:"ip_address"`
	Port       string `xml:"port"`
	TableName  string `xml:"table_name"`
}

func (cfg *DBConfig) LoadXMLDBConfig(config_path string) {
	if config_path == "" {
		config_path = "db_config.xml"
	}
	xmlFile, err := os.Open(config_path)
	if err != nil {
		log.Fatalf("Error: db_config.go [ParseDBConfig] %serror! \n", config_path)
	}
	defer xmlFile.Close()
	xmlData, err := ioutil.ReadAll(xmlFile)
	if err != nil {
		log.Fatalln("Error reading XML data:", err)
	}
	xml.Unmarshal(xmlData, &cfg)
	log.Println("XMLName : ", cfg.XMLName)
}

func (cfg *DBConfig) ConnDB() *sql.DB {
	dbConfigInfo0 := cfg.DbBase
	dataSourceName := fmt.Sprintf("%s:%s@%s(%s:%s)/%s", dbConfigInfo0.Username, dbConfigInfo0.Password, dbConfigInfo0.IPProtocol,
		dbConfigInfo0.IPAddress, dbConfigInfo0.Port, dbConfigInfo0.DBName)
	driverName := dbConfigInfo0.DBType
	log.Println("driverName = ", driverName)
	log.Println("dataSourceName = ", dataSourceName)
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		log.Fatalln(err)
	}
	db.SetMaxOpenConns(2000)
	db.SetMaxIdleConns(1000)
	db.SetConnMaxLifetime(time.Minute * 60)
	return db
}
