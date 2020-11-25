package log

import "log"

func Info(msg string, data ...interface{}) {
	log.Printf("rao info  %v  %v \n", msg, data)
}
