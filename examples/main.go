package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ayopop-tech/go-elastic"
	"io"
	"log"
)

type Product struct {
	Name   string
	Colors []string
}

func BulkIndexConstant(indexName, documentType string) string {
	return `{"index": { "_index": "` + indexName + `", "_type": "` + documentType + `" } }`
}

func main() {
	esClient := elastic.NewClient("http", "localhost", "9200", "", "")

	var buffer bytes.Buffer
	userId := "1"
	articleId := "22"
	articleStatus := "published"
	publishedAt := "2019-02-02 11:55:23"
	maxResult := 50

	// Bulk Insert
	bulkInsertData := [...]Product{
		Product{Name: "Jeans", Colors: []string{"blue", "red"}},
		Product{Name: "Polo", Colors: []string{"yellow", "red"}},
		Product{Name: "Shirt", Colors: []string{"brown", "blue"}},
	}

	bulkProduct := make([]interface{}, len(bulkInsertData))
	for i := range bulkInsertData {
		bulkProduct[i] = bulkInsertData[i]
	}

	for _, value := range bulkProduct {
		buffer.WriteString(BulkIndexConstant("2019-01-01115523_1_refunded", "22"))
		buffer.WriteByte('\n')

		jsonProduct, _ := json.Marshal(value)
		buffer.Write(jsonProduct)
		buffer.WriteByte('\n')
	}

	_, err := esClient.BulkInsert(buffer.Bytes())
	if err != nil {
		fmt.Println(err.Error())
	}

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

	// Insert single document
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

	// Search documents
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

	fmt.Println([]byte(jsonResult))
}
