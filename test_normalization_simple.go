package apigateway
package main

import (
	"encoding/json"
	"fmt"
)

// Simple test to verify normalization logic
func main() {
	// Simulate the Gomunime nested structure
	gomunimeResponse := `{
		"data": {
			"confidence_score": 1,
			"data": {
				"anime_slug": "test-anime",
				"cover": "https://example.com/cover.jpg",
				"details": {
					"Duration": "24 min"
				}
			},
			"message": "Data berhasil diambil",
			"source": "gomunime.co"
		}
	}`

	var response map[string]interface{}
	json.Unmarshal([]byte(gomunimeResponse), &response)

	fmt.Println("=== Testing Nested Structure Detection ===")
	
	dataField, hasData := response["data"]
	fmt.Printf("Has data field: %v\n", hasData)
	
	if hasData {
		if dataMap, isMap := dataField.(map[string]interface{}); isMap {
			nestedData, hasNestedData := dataMap["data"]
			fmt.Printf("Has nested data.data: %v\n", hasNestedData)
			
			if hasNestedData {
				if nestedDataMap, isNestedMap := nestedData.(map[string]interface{}); isNestedMap {
					fmt.Printf("Nested data is a map: %v\n", isNestedMap)
					fmt.Printf("Nested data keys: %v\n", getKeys(nestedDataMap))
				}
			}
			
			fmt.Printf("Outer data keys: %v\n", getKeys(dataMap))
		}
	}
	
	fmt.Println("\n=== Testing Normalization Logic ===")
	normalizedJSON := normalizeResponse([]byte(gomunimeResponse))
	fmt.Printf("Normalized response:\n%s\n", normalizedJSON)
}

func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func normalizeResponse(data []byte) []byte {
	var response map[string]interface{}
	if err := json.Unmarshal(data, &response); err != nil {
		return data
	}

	dataField, hasData := response["data"]
	if !hasData {
		return data
	}

	dataMap, isDataMap := dataField.(map[string]interface{})
	if !isDataMap {
		return data
	}

	nestedData, hasNestedData := dataMap["data"]
	if hasNestedData {
		fmt.Println("FOUND nested structure - normalizing...")
		
		if nestedDataMap, isNestedMap := nestedData.(map[string]interface{}); isNestedMap {
			normalizedResponse := make(map[string]interface{})
			
			// Copy top-level fields (excluding data)
			for key, value := range response {
				if key != "data" {
					normalizedResponse[key] = value
				}
			}
			
			// Create normalized data structure
			normalizedData := make(map[string]interface{})
			
			// Copy nested data fields to top level of data
			for key, value := range nestedDataMap {
				normalizedData[key] = value
			}
			
			// Preserve metadata fields from the original data level
			for key, value := range dataMap {
				if key != "data" {
					normalizedData[key] = value
				}
			}
			
			normalizedResponse["data"] = normalizedData
			
			if jsonBytes, err := json.MarshalIndent(normalizedResponse, "", "  "); err == nil {
				return jsonBytes
			}
		}
	}
	
	return data
}