package elastic

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

// Contract to manage indices and find data
type Client interface {
	CreateIndex(indexName, mapping string) (bool, error)
	DeleteIndex(indexName string) (bool, error)
	IndexExists(indexName string) (bool, error)
	InsertDocument(indexName string, documentType string, identifier string, data []byte) (bool, error)
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

	return &client{Host: u}
}

// CreateIndex instantiates an index
// https://www.elastic.co/guide/en/elasticsearch/reference/5.6/indices-create-index.html
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
// https://www.elastic.co/guide/en/elasticsearch/reference/5.6/indices-delete-index.html
func (c *client) DeleteIndex(indexName string) (bool, error) {
	esUrl := c.Host.String() + "/" + indexName
	_, err := sendHTTPRequest("DELETE", esUrl, nil)
	if err != nil {
		return false, err
	}

	return true, nil
}

// IndexExists allows to check if the index exists or not.
// https://www.elastic.co/guide/en/elasticsearch/reference/5.6/indices-exists.html
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
// https://www.elastic.co/guide/en/elasticsearch/reference/current/docs-index_.html
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
