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


## Contributing

1. Create an issue, describe the bugfix/feature you wish to implement.
2. Fork the repository
3. Create your feature branch (`git checkout -b my-new-feature`)
4. Commit your changes (`git commit -am 'Add some feature'`)
5. Push to the branch (`git push origin my-new-feature`)
6. Create a new Pull Request