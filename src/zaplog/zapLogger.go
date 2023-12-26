// ///////////////////////////////////////
// 2023 SHAILab Storage all rights reserved
// ///////////////////////////////////////
package zaplog

import (
	"log"
	"os"

	"go.uber.org/zap"
)

var Logger *zap.Logger

func init() {
	var err error
	Logger, err = zap.NewDevelopment()
	if err != nil {
		log.Printf("ZapLogger init failed. \n")
		os.Exit(1)
	}
	defer Logger.Sync()
}
