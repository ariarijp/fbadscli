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

func printDebugInfo(debugInfo *fb.DebugInfo) {
	l := log.New(os.Stderr, "DEBUG ", log.LstdFlags)
	l.Println("fb.DebugInfo.FacebookApiVersion:", debugInfo.FacebookApiVersion)
	l.Println("fb.DebugInfo.FacebookDebug:", debugInfo.FacebookDebug)
	l.Println("fb.DebugInfo.FacebookRev:", debugInfo.FacebookRev)
	l.Println("fb.DebugInfo.Proto:", debugInfo.Proto)
	for k, v := range debugInfo.Header {
		l.Println("fb.DebugInfo.Header:", k, v)
	}
	for _, m := range debugInfo.Messages {
		l.Println("fb.DebugInfo.Messages:", m.Type, m.Message, m.Link)
	}
}

func main() {
	accessToken := os.Getenv("FB_ACCESS_TOKEN")
	confFileName := os.Args[1]

	debugMode := os.Getenv("FB_DEBUG") != ""
	if debugMode {
		fb.Debug = fb.DEBUG_ALL
	}

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
	if debugMode {
		printDebugInfo(res.DebugInfo())
	}

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
