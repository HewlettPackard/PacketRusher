package main

import (
	"my5G-RANTester/config"
	"my5G-RANTester/internal/control_test_engine/ue"
)

func main() {
	cfg, err := config.GetConfig()
	if err != nil {
		//return nil
		//log.Fatal("Error in get configuration")
	}
	ue.RegistrationUe("200", cfg, 1)
}
