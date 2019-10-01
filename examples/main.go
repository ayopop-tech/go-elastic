package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ayopop-tech/go-elastic"
	"io"
	"log"
)

type Article struct {
	ArticleId     string
	ArticleStatus string
	PublishedAt   string
}

func BulkIndexConstant(indexName, documentType string) string {
	return `{"index": { "_index": "` + indexName + `", "_type": "` + documentType + `" } }`
}

func main() {
	esClient := elastic.NewClient()

	var buffer bytes.Buffer
	userId := "2"
	articleId := "22"
	articleStatus := "published"
	publishedAt := "2019-02-02 11:55:23"
	maxResult := 50
	indexName := userId + "_" + articleStatus

	// Bulk-insert data
	bulkInsertData := [...]Article{
		Article{
			ArticleId:     "1",
			ArticleStatus: "processing",
			PublishedAt:   "2019-02-02 11:55:23",
		},
		Article{
			ArticleId:     "2",
			ArticleStatus: "processing",
			PublishedAt:   "2019-02-02 11:55:23",
		},
		Article{
			ArticleId:     "3",
			ArticleStatus: "processing",
			PublishedAt:   "2019-02-02 11:55:23",
		},
	}

	bulkProduct := make([]interface{}, len(bulkInsertData))
	for i := range bulkInsertData {
		bulkProduct[i] = bulkInsertData[i]
	}

	for _, value := range bulkProduct {
		buffer.WriteString(BulkIndexConstant(indexName, articleId))
		buffer.WriteByte('\n')

		jsonProduct, _ := json.Marshal(value)
		buffer.Write(jsonProduct)
		buffer.WriteByte('\n')
	}

	_, err := esClient.BulkInsert(buffer.Bytes())
	if err != nil {
		fmt.Println(err.Error())
	}

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
		"article_id":   articleId,
		"status":       articleStatus,
		"published_at": publishedAt,
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

	_ = json.NewDecoder(searchResults).Decode(&mapResp)

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
