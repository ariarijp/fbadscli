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

	session := &fb.Session{}
	session.SetAccessToken(accessToken)
	err := session.Validate()
	if err != nil {
		log.Fatal(err)
	}

	endpointUrl := conf.EndpointURL
	res, err := session.Get(endpointUrl, fb.Params{
		"fields": strings.Join(conf.Fields, ","),
		"limit":  conf.Limit,
	})
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
