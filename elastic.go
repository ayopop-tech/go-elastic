package elastic

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

// Contract to manage indices and find data
type Client interface {
	CreateIndex(indexName, mapping string) (bool, error)
	DeleteIndex(indexName string) (bool, error)
	IndexExists(indexName string) (bool, error)
	InsertDocument(indexName string, documentType string, identifier string, data []byte) (bool, error)
	BulkInsertDocuments(data []byte) (bool, error)
	FindDocuments(indexName string, documentType string, maxResults int) (interface{}, error)
}

// A SearchClient describes the client configuration to manage an ElasticSearch index.
type client struct {
	Host url.URL
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

// NewSearchClient creates and initializes a new ElasticSearch client, implements core api for Indexing and searching.
func NewClient(scheme, host, port, username, password string) *client {

	err := godotenv.Load()
	if err != nil {
		fmt.Println("Please add env file")
		return nil
	}

	if os.Getenv("ELASTICSEARCH_SCHEMA") == "" || os.Getenv("ELASTICSEARCH_HOST") == "" || os.Getenv("ELASTICSEARCH_PORT") == "" {
		fmt.Println("Please add necessary parameters to env file")
		return nil
	}

	u := url.URL{
		Scheme: os.Getenv("ELASTICSEARCH_SCHEMA"),
		Host:   os.Getenv("ELASTICSEARCH_HOST") + ":" + os.Getenv("ELASTICSEARCH_PORT"),
		User:   url.UserPassword(os.Getenv("ELASTICSEARCH_USERNAME"), os.Getenv("ELASTICSEARCH_PASSWORD")),
	}

	if username == "" && password == "" {
		u = url.URL{
			Scheme: scheme,
			Host:   host + ":" + port,
		}
	}

	return &client{Host: u}
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
	_, err := httpClient.Head(esUrl)
	if err != nil {
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

// Finds document list for specific index
func (c *client) FindDocuments(indexName string, documentType string, maxResults int) (io.ReadCloser, error) {
	esUrl := c.Host.String() + "/" + indexName + "/" + documentType + "/_search"
	if maxResults >= 0 {
		esUrl += "?size=" + strconv.Itoa(maxResults)
	}
	resp, err := http.Get(esUrl)
	if err != nil {
		panic("Error on getting search result")
	}

	return resp.Body, nil
}

func (c *client) BulkInsert(data []byte) (bool, error) {
	esUrl := c.Host.String() + "/_bulk"
	reader := bytes.NewBuffer(data)
	resp, err := sendHTTPRequest("POST", esUrl, reader)
	if err != nil {
		return false, err
	}

	fmt.Println("Bulk insert response", resp)

	return true, nil
}
