package main

import (
	"encoding/json"
	"fmt"
	"github.com/ayopop-tech/go-elastic"
	"io"
	"log"
)

func main() {
	esClient := elastic.NewClient("http", "localhost", "9200", "", "")

	userId := "1"
	articleId := "22"
	articleStatus := "published"
	publishedAt := "2019-02-02 11:55:23"
	maxResult := 50

	indexName := userId + "_" + articleStatus

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
		"tid":            articleId,
		"current_status": articleStatus,
		"time":           publishedAt,
	}

	marshalledData, _ := json.Marshal(data)

	_, err = esClient.InsertDocument(indexName, articleId, marshalledData)
	if err != nil {
		panic(err.Error())
	}

	searchResults, err := esClient.FindDocuments(indexName, articleId, maxResult)
	if err != nil {
		panic(err.Error())
	}

	transformSearchResults(searchResults)
}

func transformSearchResults(searchResults io.ReadCloser) {
	var mapResp map[string]interface{}
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
