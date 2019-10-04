package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ayopop-tech/go-elastic"
)

type Article struct {
	ArticleId     string
	ArticleStatus string
	PublishedAt   string
}

func BulkIndexConstant(indexName, documentType string) string {
	return `{"index": { "_index": "` + indexName + `", "_type": "` + documentType + `" } }`
}

func AddFindDocumentConstant(documentType string) []byte {
	documentQuery := `{
		"version": true,
		"query": {
			"term": {
				"_type": {
					"value": ` + documentType + `
				}
			}
		}
	}`

	return []byte(documentQuery)
}

func main() {
	esClient := elastic.NewClient()

	var buffer bytes.Buffer
	userId := "1"
	articleId := "21"
	articleStatus := "processing"
	publishedAt := "2019-08-09 11:55:23"
	formattedPublishedAt := strings.Replace(strings.Replace(publishedAt, ":", "", -1), " ", "", -1)
	indexName := formattedPublishedAt + "_" + userId

	fmt.Println(indexName)

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

	_, err = esClient.BulkInsert(buffer.Bytes())
	if err != nil {
		fmt.Println("Error", err.Error())
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

	findDocumentQuery := AddFindDocumentConstant(articleId)
	fmt.Println(string(findDocumentQuery))

	// Search documents
	searchResults, err := esClient.FindDocuments(indexName, findDocumentQuery)
	if err != nil {
		panic(err.Error())
	}

	fmt.Println(searchResults)
}
