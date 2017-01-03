package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/koron/go-dproxy"
)

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
	if action != "OPTIONAL" {
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
