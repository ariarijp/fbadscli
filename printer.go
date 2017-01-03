package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

func insightsMapAsJsonlLines(insights []map[string]interface{}) []string {
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
	for _, json := range insightsMapAsJsonlLines(insights) {
		fmt.Println(json)
	}
}

func printAsJson(insights []map[string]interface{}) {
	fmt.Println("[")

	jsonlLines := insightsMapAsJsonlLines(insights)

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
