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

func getInsights(a []interface{}, action string, action_attribution_window string) []map[string]interface{} {
	var result []map[string]interface{}

	for _, v := range a {
		ad := dproxy.New(v)

		insights, err := ad.M("insights").M("data").A(0).Map()
		if err != nil {
			continue
		}

		row := make(map[string]interface{})

		for k, v := range insights {
			switch v.(type) {
			case []interface{}:
				if k != "actions" {
					row[k] = v
					break
				}

				row[action] = 0.0

				for _, v2 := range v.([]interface{}) {
					actionType := v2.(map[string]interface{})["action_type"]

					if actionType == action {
						value := v2.(map[string]interface{})[action_attribution_window]
						if value != nil {
							row[action] = value.(float64)
						}
					}
				}
			case map[string]interface{}:
				row[k] = fmt.Sprintf("%v", v.(map[string]interface{}))
			case string:
				if k == "ad_id" {
					row[k] = v
					continue
				}

				f, err := strconv.ParseFloat(v.(string), 64)
				if err != nil {
					row[k] = v
					continue
				}

				row[k] = f
			case float64:
				row[k] = v.(float64)
			default:
				row[k] = v
			}
		}

		result = append(result, row)
	}

	return result
}

func getJsonlLines(insights []map[string]interface{}) []string {
	var result []string

	for _, insight := range insights {
		bytes, err := json.Marshal(insight)

		if err != nil {
			log.Fatal(err)
		}

		result = append(result, string(bytes))
	}

	return result
}

func printAsJsonl(insights []map[string]interface{}) {
	for _, json := range getJsonlLines(insights) {
		fmt.Println(json)
	}
}

func printAsJson(insights []map[string]interface{}) {
	fmt.Println("[")

	jsonlLines := getJsonlLines(insights)

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

func printAsCsv(insights []map[string]interface{}, keys []string, sep string) {
	for _, insight := range insights {
		var row []string

		for _, key := range keys {
			switch insight[key].(type) {
			case []interface{}:
				bytes, err := json.Marshal(insight[key])
				if err != nil {
					log.Fatal(err)
				}

				row = append(row, string(bytes))
			case float64:
				if key == "ad_id" {
					row = append(row, fmt.Sprintf("%d", int(insight[key].(float64))))
					continue
				}

				row = append(row, fmt.Sprintf("%v", insight[key]))
			case nil:
				row = append(row, "")
			default:
				row = append(row, fmt.Sprintf("%v", insight[key]))
			}
		}

		fmt.Println(strings.Join(row, sep))
	}
}

func buildRequest(token string, api_version string, act string, date_preset string, fields string, action string, action_attribution_window string) *http.Request {
	if action != "" {
		fields = fields + ",actions"
	}

	u, err := url.Parse("https://graph.facebook.com")
	if err != nil {
		log.Fatal(err)
	}

	u.Path = fmt.Sprintf("/%s/act_%s/ads", api_version, act)

	q := u.Query()
	if action != "" {
		q.Set("fields", fmt.Sprintf("insights.date_preset(%s).action_attribution_windows(%s){%s}", date_preset, fields))
	} else {
		q.Set("fields", fmt.Sprintf("insights.date_preset(%s){%s}", date_preset, fields))
	}

	q.Set("limit", "200")
	q.Set("access_token", token)

	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		log.Fatal(err)
	}

	return req
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

	flag.StringVar(&act, "act", "", "Ad account ID")
	flag.StringVar(&format, "format", "jsonl", "Output format")
	flag.StringVar(&date_preset, "date_preset", "this_month", "Date preset")
	flag.StringVar(&fields, "fields", "ad_id,ad_name,impressions,inline_link_clicks,spend", "Insights fields")
	flag.StringVar(&colsep, "colsep", ",", "Column separator(only for CSV format)")
	flag.StringVar(&api_version, "api_version", "v2.7", "Marketing API version")
	flag.StringVar(&sort_key, "sort_key", "", "Sort key")
	flag.StringVar(&sort_order, "sort_order", "asc", "Sort order")
	flag.StringVar(&action, "action", "", "Action")
	flag.StringVar(&action_attribution_window, "action_attribution_window", "1d_click", "Action attribution windows")
	flag.Parse()

	if colsep == "\\t" {
		colsep = "\t"
	}

	token = os.Getenv("FB_ACCESS_TOKEN")
	if token == "" {
		log.Fatal("Access token is required. Please set your valid access token to FB_ACCESS_TOKEN environment variable.")
	}

	req := buildRequest(token, api_version, act, date_preset, fields, action, action_attribution_window)

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

	if sort_key != "" {
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
}
