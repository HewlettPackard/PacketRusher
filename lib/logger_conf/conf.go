package logger_conf

import (
	"log"
	"os"
	"strconv"

	"my5G-RANTester/lib/path_util"
)

var Free5gcLogDir string = path_util.Gofree5gcPath("free5gc/log") + "/"
var LibLogDir string = Free5gcLogDir + "lib/"
var NfLogDir string = Free5gcLogDir + "nf/"

var Free5gcLogFile string = Free5gcLogDir + "free5gc.log"

func init() {
	_ = os.MkdirAll(LibLogDir, 0775)
	_ = os.MkdirAll(NfLogDir, 0775)

	f, err := os.OpenFile(Free5gcLogFile, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		log.Printf("Cannot Open %s\n", Free5gcLogFile)
	} else {
		f.Close()
	}

	sudoUID, errUID := strconv.Atoi(os.Getenv("SUDO_UID"))
	sudoGID, errGID := strconv.Atoi(os.Getenv("SUDO_GID"))

	if errUID == nil && errGID == nil {
		os.Chown(Free5gcLogDir, sudoUID, sudoGID)
		os.Chown(LibLogDir, sudoUID, sudoGID)
		os.Chown(NfLogDir, sudoUID, sudoGID)

		if err == nil {
			os.Chown(Free5gcLogFile, sudoUID, sudoGID)
		}
	}
}
