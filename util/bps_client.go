package util

import (
	"crypto/tls"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func CallBPS(method string, url string, data string) (result []byte, err error) {
	req, err := http.NewRequest(method, "http://159.8.154.62:8080"+url, strings.NewReader(data))
	if err != nil {
		// handle error
	}

	if method == "POST" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	//	if UserInfo != nil {
	//		req.Header.Set("Auth-Expire", strconv.FormatFloat(UserInfo["expire_time"].(float64), 'f', 0, 64))
	//		req.Header.Set("Auth-Role", UserInfo["role"].(string))
	//		req.Header.Set("Auth-Scope", UserInfo["scope"].(string))
	//		req.Header.Set("Auth-Tenant", UserInfo["tenant"].(string))
	//		req.Header.Set("Auth-Token", UserInfo["token"].(string))
	//		req.Header.Set("Auth-Username", UserInfo["username"].(string))
	//	}
	req.SetBasicAuth("admin", "admin")
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	//	log.Println( string(body) )
	return body, err
}
