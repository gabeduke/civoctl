package civo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

const baseUri = "https://api.civo.com/v2/kubernetes/clusters"

// Civoer interfaces with the Civo API
type Civoer interface {
	getClusters() (*Clusters, error)

	GetClusterId(name string) (id string, err error)
	GetClusterNames() []string
	CreateCluster(name string) error
	DeleteCluster(name string) error
}

// Civo holds the Civoer interface and http client
type Civo struct {
	Civoer
	client *http.Client
}

// Clusters is the Civo API response for Clusters list
type Clusters struct {
	Items []Cluster `json:"items"`
}

type Cluster struct {
	ID                string        `json:"id"`
	Name              string        `json:"name"`
	Version           string        `json:"version"`
	Status            string        `json:"status"`
	Ready             bool          `json:"ready"`
	NumTargetNodes    int           `json:"num_target_nodes"`
	TargetNodesSize   string        `json:"target_nodes_size"`
	Kubeconfig        string        `json:"kubeconfig"`
	KubernetesVersion string        `json:"kubernetes_version"`
	DNSEntry          string        `json:"dns_entry"`
	Tags              []interface{} `json:"tags"`
}

// newCivoHandler generates an authenticated client for the Civo API
func newCivoHandler(token string) *Civo {
	c := Civo{}
	c.client = getCivoHttpClient(token)

	return &c
}

// GetClusterId gets an ID for a given cluster name
func (c *Civo) GetClusterId(name string) (id string, err error) {
	v, err := c.getClusters()
	if err != nil {
		return "", err
	}

	for _, cluster := range v.Items {
		if cluster.Name == name {
			return cluster.ID, nil
		}
	}
	return "", nil
}

func getCivoHttpClient(token string) *http.Client {
	ctx := context.Background()
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: token,
		TokenType:   "Bearer",
	}))

	return client
}

// DeleteCluster calls delete from the Civo API
func (c *Civo) DeleteCluster(name string) error {

	uri := fmt.Sprintf("%s/%s", baseUri, name)

	// Create a new request using http
	req, err := http.NewRequest(http.MethodDelete, uri, nil)

	// Send req using http Client
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	err = handleResponse(resp)
	if err != nil {
		return err
	}

	log.Infof("status: %s", resp.Status)

	return nil
}

// CreateCluster calls create from the Civo API
func (c *Civo) CreateCluster(cluster *Cluster) error {

	fields := log.Fields{
		"API call": "Create Cluster",
	}

	data := url.Values{}
	log.WithFields(fields).Debugf("Cluster Name: %s", cluster.Name)
	data.Set("name", cluster.Name)

	if cluster.NumTargetNodes != 0 {
		t := strconv.Itoa(cluster.NumTargetNodes)
		log.WithFields(fields).Debugf("NumTargetNodes: %s", t)
		data.Set("num_target_nodes", t)
	}

	// Create a new request using http
	req, err := http.NewRequest("POST", baseUri, bytes.NewBufferString(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	log.WithFields(fields).Debugf("request: %v", req.URL)

	// Send req using http Client
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	log.WithFields(fields).Debugf("response: %v", resp.StatusCode)

	err = handleResponse(resp)
	if err != nil {
		return err
	}

	return nil
}

// GetClusterNames gets a list of Clusters from the Civo API
func (c *Civo) GetClusterNames() ([]string, error) {
	var clusters []string

	v, err := c.getClusters()
	if err != nil {
		return nil, err
	}
	if len(v.Items) == 0 {
		return clusters, nil
	}

	for _, c := range v.Items {
		clusters = append(clusters, c.Name)
	}
	return clusters, nil
}

func (c *Civo) getClusters() (*Clusters, error) {

	fields := log.Fields{
		"API call": "Get Clusters",
	}

	log.WithFields(fields).Debugf("baseUrl: %v", baseUri)

	resp, err := c.client.Get(baseUri)
	if err != nil {
		return nil, err
	}
	log.WithFields(fields).Tracef("response: %v", resp)

	err = handleResponse(resp)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	log.WithFields(fields).Tracef("body: %v", body)

	v := Clusters{}
	json.Unmarshal(body, &v)
	return &v, nil
}

func handleResponse(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		log.Debugf("Response status: %v", resp.StatusCode)
	} else {
		return fmt.Errorf("civo response status: %v", resp.StatusCode)
	}
	return nil
}
