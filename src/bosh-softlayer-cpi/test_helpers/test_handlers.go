package test_helpers

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/onsi/gomega/ghttp"
)

func ReadJsonTestFixtures(workingDir, packageName, fileName string) ([]byte, error) {
	return ioutil.ReadFile(filepath.Join(workingDir, "test_fixtures", packageName, fileName))
}

func readResponseDataFromFile(serviceMethodFile string) ([]byte, error) {
	wd, err := os.Getwd()
	if err != nil {
		return []byte{}, err
	}

	fixture := filepath.Join(wd, "../../test_fixtures/services", serviceMethodFile)
	resp, err := ioutil.ReadFile(fixture)
	if err != nil {
		return []byte{}, err
	}

	return resp, nil
}

func SpecifyServerResps(respParas []map[string]interface{}, server *ghttp.Server) error {
	var respData []byte
	var err error
	for _, respPara := range respParas {
		respData, err = readResponseDataFromFile(respPara["filename"].(string))
		if err != nil {
			return err
		}

		if path, ok := respPara["path"]; ok {
			server.RouteToHandler(
				respPara["method"].(string),
				path.(string),
				ghttp.RespondWith(
					respPara["statusCode"].(int),
					[]byte(respData),
					http.Header{"Content-Type": []string{"application/json"}},
				),
			)
		} else if _, ok := respPara["auth"]; ok {
			server.AppendHandlers(
				ghttp.RespondWith(
					respPara["statusCode"].(int),
					[]byte(respData),
					http.Header{
						"Content-Type":  []string{"application/json"},
						"X-Auth-Token":  []string{respPara["auth"].(string)},
						"X-Storage-Url": []string{respPara["storage_url"].(string) + "/" + respPara["auth"].(string)},
					},
				),
			)
		} else {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.RespondWith(
						respPara["statusCode"].(int),
						[]byte(respData),
						http.Header{"Content-Type": []string{"application/json"}},
					),
				),
			)
		}
	}

	return nil
}
