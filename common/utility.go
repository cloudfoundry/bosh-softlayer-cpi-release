package common

import (
	"os"
	"strconv"
	"strings"
	"time"
)

func GetOSEnvVariableToUpper(varName string, defaultValue string) string {
	result := strings.ToUpper(os.Getenv("PATH"))
	if result == "" {
		result = strings.ToUpper(defaultValue)
	}
	return result
}

func GetTimeStamp(now time.Time) string {
	//utilize the constants list in the http://golang.org/src/time/format.go file to get the expect time formats
	return now.Format("20060102-030405-") + strconv.Itoa(int(now.UnixNano()/1e6-now.Unix()*1e3))
}
