package gokeeper

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	GokeeperHost = "http://10.143.168.10:17000"

	GokeeperAPINodeList = "/node/list"
)

func GetAliveNodeList(domain, component string) ([]string, error) {
	u := fmt.Sprintf("%s%s?domain=%s&component=%s", GokeeperHost, GokeeperAPINodeList, domain, component)

	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

}
