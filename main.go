package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	fb "github.com/huandu/facebook"
)

type config struct {
	EndpointURL string
	Fields      []string
	Params      map[string]interface{}
	Version     string
}

func main() {
	accessToken := os.Getenv("FB_ACCESS_TOKEN")
	confFileName := os.Args[1]

	var conf config
	if _, err := toml.DecodeFile(confFileName, &conf); err != nil {
		log.Fatal(err)
	}

	session := &fb.Session{}
	session.SetAccessToken(accessToken)
	session.Version = conf.Version
	err := session.Validate()
	if err != nil {
		log.Fatal(err)
	}

	endpointUrl := conf.EndpointURL
	params := fb.Params{}
	if len(conf.Fields) > 0 {
		params["fields"] = strings.Join(conf.Fields, ",")
	}

	for k, v := range conf.Params {
		params[k] = v
	}

	res, err := session.Get(endpointUrl, params)
	if err != nil {
		log.Fatal(err)
	}

	paging, err := res.Paging(session)
	if err != nil {
		log.Fatal(err)
	}

	for {
		for _, result := range paging.Data() {
			jsonStr, err := json.Marshal(result)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(string(jsonStr))
		}

		noMore, err := paging.Next()
		if err != nil {
			log.Fatal(err)
		}
		if noMore {
			break
		}
	}
}
