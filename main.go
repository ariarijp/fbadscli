package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/koron/go-dproxy"
)

type InsightsSlice []map[string]interface{}

func (is InsightsSlice) Len() int {
	return len(is)
}

func (is InsightsSlice) Swap(i, j int) {
	is[i], is[j] = is[j], is[i]
}

type ByFloat64 struct {
	InsightsSlice
	key string
}

func (bf ByFloat64) Less(i, j int) bool {
	var a, b float64
	var err error

	switch bf.InsightsSlice[i][bf.key].(type) {
	case float64:
		a = bf.InsightsSlice[i][bf.key].(float64)
		b = bf.InsightsSlice[j][bf.key].(float64)
	case string:
		a, err = strconv.ParseFloat(bf.InsightsSlice[i][bf.key].(string), 64)
		if err != nil {
			log.Fatal(err)
		}

		b, err = strconv.ParseFloat(bf.InsightsSlice[j][bf.key].(string), 64)
		if err != nil {
			log.Fatal(err)
		}
	}

	return a < b
}

func main() {
	var (
		act                       string
		format                    string
		date_preset               string
		fields                    string
		colsep                    string
		token                     string
		api_version               string
		sort_key                  string
		sort_order                string
		action                    string
		action_attribution_window string
	)

	flag.StringVar(&act, "act", "REQUIRED", "Ad account ID")
	flag.StringVar(&format, "format", "jsonl", "Output format")
	flag.StringVar(&date_preset, "date_preset", "this_month", "Date preset")
	flag.StringVar(&fields, "fields", "ad_id,ad_name,impressions,inline_link_clicks,spend", "Insights fields")
	flag.StringVar(&colsep, "colsep", ",", "Column separator(only for CSV format)")
	flag.StringVar(&api_version, "api_version", "v2.7", "Marketing API version")
	flag.StringVar(&sort_key, "sort_key", "OPTIONAL", "Sort key")
	flag.StringVar(&sort_order, "sort_order", "asc", "Sort order")
	flag.StringVar(&action, "action", "OPTIONAL", "Action")
	flag.StringVar(&action_attribution_window, "action_attribution_window", "1d_click", "Action attribution windows")
	flag.Parse()

	if colsep == "\\t" {
		colsep = "\t"
	}

	if act == "REQUIRED" {
		log.Fatal("act is required.")
	}
	log.Fatal(sort_key)

	token = os.Getenv("FB_ACCESS_TOKEN")
	if token == "" {
		log.Fatal("Access token is required. Please set your valid access token to FB_ACCESS_TOKEN environment variable.")
	}

	req := buildRequest(token, api_version, act, date_preset, fields, action, action_attribution_window)

	for {
		client := new(http.Client)
		resp, _ := client.Do(req)
		defer resp.Body.Close()

		byteArray, _ := ioutil.ReadAll(resp.Body)
		var v interface{}
		json.Unmarshal(byteArray, &v)

		a, err := dproxy.New(v).M("data").Array()
		if err != nil {
			log.Fatal(err)
		}

		keys := strings.Split(fields, ",")
		if action != "" {
			keys = append(keys, action)
		}

		insights := getInsights(a, action, action_attribution_window)

		if sort_key != "OPTIONAL" {
			if sort_order == "desc" {
				sort.Sort(sort.Reverse(ByFloat64{insights, sort_key}))
			} else {
				sort.Sort(ByFloat64{insights, sort_key})
			}
		}

		if format == "json" {
			printAsJson(insights)
		} else if format == "csv" {
			fmt.Println(strings.Join(keys, colsep))
			printAsCsv(insights, keys, colsep)
		} else {
			printAsJsonl(insights)
		}

		paging, _ := dproxy.New(v).M("paging").Map()

		if next, ok := paging["next"]; ok {
			u, err := url.Parse(next.(string))
			if err != nil {
				log.Fatal(err)
			}

			req, err = http.NewRequest("GET", u.String(), nil)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			break
		}
	}
}
