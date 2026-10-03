package shared

import "log"

func Check(e error, msg string) {
	if e != nil {
		log.Fatalf("ERROR: %v (%v)", e, msg)
	}
}
