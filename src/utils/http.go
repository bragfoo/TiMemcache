package util

import (
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// HTTPGet compatible http & https
func HTTPGet(reqURL string) map[string]interface{} {
	req, _ := http.NewRequest("GET", reqURL, nil)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	c := &http.Client{
		Transport: tr,
		Timeout:   300 * time.Second,
	}
	res, perr := c.Do(req)
	if perr != nil {
		ErrorLogger(perr)
		return nil
	}
	resBody, berr := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if berr != nil {
		ErrorLogger(berr)
	}
	responeDate := make(map[string]interface{})
	json.Unmarshal(resBody, &responeDate)
	return responeDate
}

// HTTPGetReturnByte compatible http & https return byte with custom header
func HTTPGetReturnByte(reqURL string, header map[string]string) []byte {
	req, _ := http.NewRequest("GET", reqURL, nil)
	for key, value := range header {
		req.Header.Set(key, value)
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	c := &http.Client{
		Transport: tr,
		Timeout:   300 * time.Second,
	}
	res, perr := c.Do(req)
	if perr != nil {
		ErrorLogger(perr)
		return nil
	}
	resBody, berr := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if berr != nil {
		ErrorLogger(berr)
	}
	return resBody
}

// HTTPPost is post func
func HTTPPost(reqURL, reqData string) map[string]interface{} {
	req, _ := http.NewRequest("POST", reqURL, strings.NewReader(reqData))
	req.Header.Set("Content-Type", "application/json")
	c := &http.Client{
		Timeout: 300 * time.Second,
	}
	res, perr := c.Do(req)
	if perr != nil {
		ErrorLogger(perr)
		return nil
	}
	resBody, berr := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if berr != nil {
		ErrorLogger(berr)
	}
	responeDate := make(map[string]interface{})
	json.Unmarshal(resBody, &responeDate)
	return responeDate
}

// HTTPPut is post func
func HTTPPut(reqURL, reqData string) map[string]interface{} {
	req, _ := http.NewRequest("PUT", reqURL, strings.NewReader(reqData))
	req.Header.Set("Content-Type", "application/json")
	c := &http.Client{
		Timeout: 300 * time.Second,
	}
	res, perr := c.Do(req)
	if perr != nil {
		ErrorLogger(perr)
		return nil
	}
	resBody, berr := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if berr != nil {
		ErrorLogger(berr)
	}
	responeDate := make(map[string]interface{})
	json.Unmarshal(resBody, &responeDate)
	return responeDate
}

// HTTPDelete is delete func
func HTTPDelete(reqURL, reqData string) map[string]interface{} {
	req, _ := http.NewRequest("DELETE", reqURL, nil)
	c := &http.Client{
		Timeout: 300 * time.Second,
	}
	res, perr := c.Do(req)
	if perr != nil {
		ErrorLogger(perr)
		return nil
	}
	resBody, berr := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if berr != nil {
		ErrorLogger(berr)
	}
	responeDate := make(map[string]interface{})
	json.Unmarshal(resBody, &responeDate)
	return responeDate
}
