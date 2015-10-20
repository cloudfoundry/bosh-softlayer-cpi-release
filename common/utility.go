package common

import (
	"os"
	"strconv"
	"time"
)

func GetOSEnvVariable(varName string, defaultValue string) string {
	result := os.Getenv(varName)
	if result == "" {
		result = defaultValue
	}
	return result
}

func SetOSEnvVariable(varName string, varValue string) error {
	return os.Setenv(varName, varValue)
}

func GetTimeStamp(now time.Time) string {
	//utilize the constants list in the http://golang.org/src/time/format.go file to get the expect time formats
	return now.Format("20060102-030405-") + strconv.Itoa(int(now.UnixNano()/1e6-now.Unix()*1e3))
}
