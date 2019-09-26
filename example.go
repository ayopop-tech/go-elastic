package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

type SearchResult struct {
	tid            string
	current_status string
	time           string
}

func main() {

	esClient := NewClient("http", "localhost", "9200", "", "")

	date := "2019-02-02 11:55:23"
	filteredDate := strings.Replace(strings.Replace(date, ":", "", -1), " ", "", -1)
	partnerId := "1"
	transactionId := "22"
	transactionStatus := "refunded"
	maxResult := 50
	indexName := filteredDate + "_" + partnerId + "_" + transactionStatus
	resp2, err2 := esClient.IndexExists(indexName)
	if err2 != nil {
		fmt.Println(err2)
	}

	if !resp2 {
		fmt.Println("Found index", indexName, "?", resp2)
		resp, err := esClient.CreateIndex(indexName, "")
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("Index created with index-name", indexName, " and response ", resp)
	}

	data := map[string]string{
		"tid":            transactionId,
		"current_status": transactionStatus,
		"time":           date,
	}

	marshalledData, _ := json.Marshal(data)

	_, err1 := esClient.InsertDocument(indexName, transactionId, marshalledData)
	if err1 != nil {
		fmt.Println("Error while creating document", err1)
	}

	searchResults, err3 := esClient.FindDocuments(indexName, transactionId, maxResult)
	if err3 != nil {
		fmt.Println("Error while fetching docs for ", transactionId)
	}

	var mapResp map[string]interface{}
	//var buf bytes.Buffer

	if err := json.NewDecoder(searchResults).Decode(&mapResp); err != nil {
		log.Fatalf("Error parsing the response body: %s", err)
	}

	result := []interface{}{}

	//Iterate the document "hits" returned by API call
	for _, hit := range mapResp["hits"].(map[string]interface{})["hits"].([]interface{}) {
		// Parse the attributes/fields of the document
		doc := hit.(map[string]interface{})
		// The "_source" data is another map interface nested inside of doc
		source := doc["_source"]
		// Get the document's _id and print it out along with _source data
		//fmt.Println("_source:", source, "\n")
		result = append(result, source)
	}

	//fmt.Println(result)
	jsonResult, err := json.Marshal(result)

	if err != nil {
		panic(err.Error())
	}

	fmt.Println(string(jsonResult), len(result))
}
