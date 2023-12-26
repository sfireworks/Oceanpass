// ///////////////////////////////////////
// 2023 SHAILab Storage all rights reserved
// ///////////////////////////////////////
package main

import (
	"fmt"
	"log"
	"net/http"
	server "oceanpass/src/server"
	_ "oceanpass/src/zaplog"
	"os"

	prome "oceanpass/src/prometheus"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func logOutput() {
	file := "./" + "log" + ".txt"
	logFile, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0766)
	if err != nil {
		panic(err)
	}
	log.SetOutput(logFile) // set logfile
	log.SetPrefix("[logTool]")
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.LUTC)
}

// ////////////////////////////////
// oceanpass initialization
// ////////////////////////////////

func init() {
	//set log to file
	// logOutput()

	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("os.Getwd() error! \n")
	}
	dirConfig := dir + "/conf/server_config.xml"
	dbConfig := dir + "/conf/db_config.xml"
	log.Println("Directory of server_config file:", dirConfig)

	server := server.OcnServer{}
	server.DirConfig = dirConfig
	server.DirDBConfig = dbConfig

	server.Init()
	server.RegisterHttpHandler()
	go server.DeleteExpireItemInDB()
	ServerAddr = server.Config.AddrConfig.IpAddr + ":" + server.Config.AddrConfig.HttpPort

	prome.InitPrometheus()
	log.Println("Finish init() ! \n\n ")
}

var ServerAddr string

func main() {
	log.Println("main() serverAddress: ", ServerAddr, " \n\n ")
	http.Handle("/metrics", promhttp.Handler())

	err := http.ListenAndServe(ServerAddr, nil)
	if err != nil {
		fmt.Println("Listen to http requests failed", err)
	}
}
