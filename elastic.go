package elastic

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"sync"
)

var once sync.Once
var instance *client

// Contract to manage indices and find data
type Client interface {
	CreateIndex(indexName, mapping string) (bool, error)
	DeleteIndex(indexName string) (bool, error)
	IndexExists(indexName string) (bool, error)
	InsertDocument(indexName string, documentType string, identifier string, data []byte) (bool, error)
	BulkInsertDocuments(data []byte) (bool, error)
	FindDocuments(indexName string, documentType string, maxResults int) ([]interface{}, error)
}

// A SearchClient describes the client configuration to manage an ElasticSearch index.
type client struct {
	Host url.URL
}

type FindDocumentResponse struct {
	Took    int   `json:"took"`
	Timeout bool  `json:"timed_out"`
	Shards  Shard `json:"_shards"`
	Hits    Hits  `json:"hits"`
}

type Shard struct {
	Total      int `json:"total"`
	Successful int `json:"successful"`
	Skipped    int `json:"skipped"`
	Failed     int `json:"failed"`
}

type Hits struct {
	Total    interface{} `json:"total"`
	Maxscore float64     `json:"max_score"`
	Hits     []Record    `json:hits`
}

type Record struct {
	Source interface{} `json:"_source"`
}

func sendHTTPRequest(method, url string, body io.Reader) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	if method == "POST" || method == "PUT" {
		req.Header.Set("Content-Type", "application/json")
	}

	newReq, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer newReq.Body.Close()
	response, err := ioutil.ReadAll(newReq.Body)
	if err != nil {
		return nil, err
	}

	if newReq.StatusCode > http.StatusCreated && newReq.StatusCode < http.StatusNotFound {
		return nil, errors.New(string(response))
	}

	return response, nil
}

func Connect() *client {

	err := godotenv.Load()
	if err != nil {
		fmt.Println("Please add env file")
		return nil
	}

	scheme := os.Getenv("ELASTICSEARCH_SCHEME")
	username := os.Getenv("ELASTICSEARCH_USERNAME")
	password := os.Getenv("ELASTICSEARCH_PASSWORD")
	host := os.Getenv("ELASTICSEARCH_HOST")
	port := os.Getenv("ELASTICSEARCH_PORT")

	if scheme == "" || host == "" || port == "" {
		fmt.Println("Please add necessary parameters to env file")
		return nil
	}

	u := url.URL{
		Scheme: scheme,
		Host:   host + ":" + port,
		User:   url.UserPassword(username, password),
	}

	if username == "" && password == "" {
		u = url.URL{
			Scheme: scheme,
			Host:   host + ":" + port,
		}
	}

	once.Do(func() {
		instance = &client{Host: u}
	})

	return instance
}

// NewSearchClient creates and initializes a new ElasticSearch client, implements core api for Indexing and searching.
func NewClient() *client {
	client := Connect()
	return client
}

// CreateIndex instantiates an index
func (c *client) CreateIndex(indexName, mapping string) (bool, error) {
	esUrl := c.Host.String() + "/" + indexName
	reader := bytes.NewBufferString(mapping)
	_, err := sendHTTPRequest("PUT", esUrl, reader)
	if err != nil {
		return false, err
	}
	return true, nil
}

// DeleteIndex deletes an existing index.
func (c *client) DeleteIndex(indexName string) (bool, error) {
	esUrl := c.Host.String() + "/" + indexName
	_, err := sendHTTPRequest("DELETE", esUrl, nil)
	if err != nil {
		return false, err
	}

	return true, nil
}

// IndexExists allows to check if the index exists or not.
func (c *client) IndexExists(indexName string) (bool, error) {
	esUrl := c.Host.String() + "/" + indexName
	httpClient := &http.Client{}
	resp, err := httpClient.Head(esUrl)
	if resp.StatusCode != 200 {
		return false, err
	}
	return true, nil
}

// InsertDocument adds or updates a typed JSON document in a specific index, making it searchable
func (c *client) InsertDocument(indexName, documentType string, data []byte) (bool, error) {
	esUrl := c.Host.String() + "/" + indexName + "/" + documentType
	reader := bytes.NewBuffer(data)
	_, err := sendHTTPRequest("POST", esUrl, reader)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (c *client) BulkInsert(data []byte) (bool, error) {
	esUrl := c.Host.String() + "/_bulk"
	reader := bytes.NewBuffer(data)
	_, err := sendHTTPRequest("POST", esUrl, reader)
	if err != nil {
		return false, err
	}

	return true, nil
}

// Finds document list for specific index
func (c *client) FindDocuments(indexName string, data []byte) ([]interface{}, error) {
	esUrl := c.Host.String() + "/" + indexName + "/_search"
	queryData := bytes.NewBuffer(data)
	resp, err := sendHTTPRequest("POST", esUrl, queryData)
	if err != nil {
		return nil, err
	}

	transformedResults := transformSearchResults(resp)

	return transformedResults, nil
}

func transformSearchResults(searchResults []byte) []interface{} {

	data := new(FindDocumentResponse)
	err := json.Unmarshal(searchResults, &data)
	if err != nil {
		panic("Error wile decoding response")
	}
	result := []interface{}{}

	hitsData := data.Hits.Hits

	//Iterate the document "hits"
	for _, hitItem := range hitsData {
		source := hitItem.Source
		result = append(result, source)
	}

	return result

}
