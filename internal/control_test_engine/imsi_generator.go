package control_test_engine

import "strconv"

// generated a IMSI from integer.
func ImsiGenerator(i int) string {

	var base string
	switch true {
	case i < 10:
		base = "imsi-208930000000"
	case i < 100:
		base = "imsi-20893000000"
	case i >= 100:
		base = "imsi-2089300000"
	}

	imsi := base + strconv.Itoa(i)
	return imsi
}
