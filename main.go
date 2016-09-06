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
	"strings"

	"github.com/koron/go-dproxy"
)

func main() {
	var (
		act         string
		format      string
		date_preset string
		fields      string
		colsep      string
		token       string
		api_version string
	)

	flag.StringVar(&act, "act", "", "Ad account ID")
	flag.StringVar(&format, "format", "jsonl", "Output format")
	flag.StringVar(&date_preset, "date_preset", "this_month", "Date preset")
	flag.StringVar(&fields, "fields", "ad_id,ad_name,impressions,inline_link_clicks,spend", "Insights fields")
	flag.StringVar(&colsep, "colsep", ",", "Column separator(only for CSV format)")
	flag.StringVar(&api_version, "api_version", "v2.7", "Marketing API version")
	flag.Parse()

	if colsep == "\\t" {
		colsep = "\t"
	}

	token = os.Getenv("FB_ACCESS_TOKEN")
	if token == "" {
		log.Fatal("Access token is required. Please set your valid access token to FB_ACCESS_TOKEN environment variable.")
	}

	u, err := url.Parse("https://graph.facebook.com")
	if err != nil {
		log.Fatal(err)
	}

	u.Path = fmt.Sprintf("/%s/act_%s/ads", api_version, act)

	q := u.Query()
	q.Set("fields", fmt.Sprintf("insights.date_preset(%s){%s}", date_preset, fields))
	q.Set("limit", "100")
	q.Set("access_token", token)
	u.RawQuery = q.Encode()

	req, _ := http.NewRequest("GET", u.String(), nil)

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

	if format == "json" {
		printAsJson(a)
	} else if format == "csv" {
		keys := strings.Split(fields, ",")
		fmt.Println(strings.Join(keys, colsep))
		printAsCsv(a, keys, colsep)
	} else {
		printAsJsonl(a)
	}
}

func getInsights(a []interface{}) []map[string]interface{} {
	var result []map[string]interface{}

	for _, v := range a {
		ad := dproxy.New(v)
		insights, err := ad.M("insights").M("data").A(0).Map()

		if err != nil {
			continue
		}

		result = append(result, insights)
	}

	return result
}

func getJsonlLines(a []interface{}) []string {
	var result []string

	for _, v := range a {
		ad := dproxy.New(v)
		insights, err := ad.M("insights").M("data").A(0).Map()

		if err == nil {
			bytes, err := json.Marshal(insights)

			if err != nil {
				log.Fatal(err)
			}

			result = append(result, string(bytes))
		}
	}

	return result
}

func printAsJsonl(a []interface{}) {
	for _, json := range getJsonlLines(a) {
		fmt.Println(json)
	}
}

func printAsJson(a []interface{}) {
	fmt.Println("[")

	jsonlLines := getJsonlLines(a)

	for i, json := range jsonlLines {
		fmt.Print("  ")
		fmt.Print(json)

		if i != len(jsonlLines)-1 {
			fmt.Print(",")
		}

		fmt.Print("\n")
	}

	fmt.Println("]")
}

func printAsCsv(a []interface{}, keys []string, sep string) {
	for _, insights := range getInsights(a) {
		var row []string

		for _, key := range keys {
			row = append(row, fmt.Sprintf("%v", insights[key]))
		}

		fmt.Println(strings.Join(row, sep))
	}
}
