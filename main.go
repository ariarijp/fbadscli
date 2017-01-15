package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/bitly/go-simplejson"
)

type Config struct {
	EndpointURL string
	Fields      []string
	Limit       int
}

func main() {
	accessToken := os.Getenv("FB_ACCESS_TOKEN")
	confFileName := os.Args[1]

	var conf Config
	if _, err := toml.DecodeFile(confFileName, &conf); err != nil {
		log.Fatal(err)
	}

	client := &http.Client{}
	endpointUrl := conf.EndpointURL
	req, err := http.NewRequest("GET", endpointUrl, nil)
	if err != nil {
		log.Fatal(err)
	}

	values := url.Values{}
	values.Add("fields", strings.Join(conf.Fields, ","))
	values.Add("limit", strconv.Itoa(conf.Limit))
	values.Add("access_token", accessToken)
	req.URL.RawQuery = values.Encode()

	for {
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}

		if resp.StatusCode != 200 {
			bytes, _ := ioutil.ReadAll(resp.Body)
			log.Fatal(string(bytes))
		}

		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		resp.Body.Close()

		sj, err := simplejson.NewJson(bytes)
		if err != nil {
			log.Fatal(err)
		}

		data, err := sj.Get("data").Array()
		if err != nil {
			log.Fatal(err)
		}

		ts := time.Now().Format("2006-01-02T15:04:05-0700")
		for _, d := range data {
			d := d.(map[string]interface{})
			d["timestamp"] = ts
			jsonStr, _ := json.Marshal(d)
			fmt.Println(string(jsonStr))
		}

		endpointUrl, _ = sj.Get("paging").Get("next").String()
		if endpointUrl == "" {
			break
		}
	}
}
