package main

import (
	"encoding/json"
	"fmt"
	"github.com/ayopop-tech/go-elastic"
	"log"
	"strings"
)

func main() {
	esClient := elastic.NewClient("http", "localhost", "9200", "", "")

	date := "2019-02-02 11:55:23"
	filteredDate := strings.Replace(strings.Replace(date, ":", "", -1), " ", "", -1)
	partnerId := "1"
	transactionId := "22"
	transactionStatus := "refunded"
	maxResult := 50
	indexName := filteredDate + "_" + partnerId + "_" + transactionStatus
	resp2, err := esClient.IndexExists(indexName)
	if err != nil {
		panic(err.Error())
	}

	if !resp2 {
		_, err := esClient.CreateIndex(indexName, "")
		if err != nil {
			panic(err.Error())
		}
	}

	data := map[string]string{
		"tid":            transactionId,
		"current_status": transactionStatus,
		"time":           date,
	}

	marshalledData, _ := json.Marshal(data)

	_, err = esClient.InsertDocument(indexName, transactionId, marshalledData)
	if err != nil {
		panic(err.Error())
	}

	searchResults, err := esClient.FindDocuments(indexName, transactionId, maxResult)
	if err != nil {
		panic(err.Error())
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
		result = append(result, source)
	}

	jsonResult, err := json.Marshal(result)

	if err != nil {
		panic(err.Error())
	}

	fmt.Println(string(jsonResult))

}
