## What is `go-elastic`?

`go-elastic` is a library to work with Elasticsearch which allows you to create a client, manipulate information in elasticsearch.


## Usage/examples:
You can refer to the example below:

```golang

package main

import (
	"github.com/ayopop-tech/go-elastic"
)

func main() {

	esClient := elastic.NewClient("http", "localhost", "9200", "", "")

	userId := "1"
	articleId := "22"
	articleStatus := "published"
	publishedAt := "2019-02-02 11:55:23"
	filteredDate := strings.Replace(strings.Replace(publishedAt, ":", "", -1), " ", "", -1)
	maxResult := 50

	indexName := filteredDate + "_" + userId + "_" + articleStatus

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

}
```

### For bulk insert:

```
golang 

// First create bulk data
bulkInsertData := [...]Product{
    Product{Name: "Jeans", Colors: []string{"blue", "red"}},
    Product{Name: "Polo", Colors: []string{"yellow", "red"}},
    Product{Name: "Shirt", Colors: []string{"brown", "blue"}},
}

// Transform it into bulk data needed as per elasticsearch
bulkProduct := make([]interface{}, len(bulkInsertData))
for i := range bulkInsertData {
    bulkProduct[i] = bulkInsertData[i]
}

for _, value := range bulkProduct {
    buffer.WriteString(BulkIndexConstant("your-index-name", "your-document-type"))
    buffer.WriteByte('\n')

    jsonProduct, _ := json.Marshal(value)
    buffer.Write(jsonProduct)
    buffer.WriteByte('\n')
}

// Send request to client
_, err := esClient.BulkInsert(buffer.Bytes())
if err != nil {
    fmt.Println(err.Error())
}

``` 

## Contributing

1. Create an issue, describe the bugfix/feature you wish to implement.
2. Fork the repository
3. Create your feature branch (`git checkout -b my-new-feature`)
4. Commit your changes (`git commit -am 'Add some feature'`)
5. Push to the branch (`git push origin my-new-feature`)
6. Create a new Pull Request