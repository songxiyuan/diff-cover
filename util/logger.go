package util

import (
	"log"
	"os"
)

var Logger *log.Logger

func init() {
	////todo 切分日志
	//file, err := os.OpenFile("diff-cover.log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	//if err != nil {
	//	fmt.Println("init Logger err! err=" + err.Error())
	//}
	Logger = log.New(os.Stdout, "[diff-cover]", log.Lshortfile|log.Ldate|log.Ltime)
}
