package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func parseRequestBody(r *http.Request, i interface{}) error {
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	if err := json.Unmarshal(b, i); err != nil {
		return err
	}

	return nil
}

func httpAddrMaker(addr string) string {
	if strings.HasSuffix(addr, "/") {
		addr = strings.TrimRight(addr, "/")
	}

	if !strings.HasPrefix(addr, "http://") {
		return fmt.Sprintf("http://%s", addr)
	}

	return addr
}

func Schemastripper(addr string) string {
	schemas := []string{"https://", "http://"}

	for _, schema := range schemas {
		if strings.HasPrefix(addr, schema) {
			return strings.TrimLeft(addr, schema)
		}
	}

	return ""
}
