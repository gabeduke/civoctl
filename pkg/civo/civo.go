package civo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"io/ioutil"
	"net/http"
	"net/url"
)

const baseUri = "https://api.civo.com/v2/kubernetes/clusters"

type Clusters struct {
	Items   []struct {
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
	} `json:"items"`
}

func GetClusterId(name string) (id string, err error) {
	v, err := getClusters()
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

func getCivoHttpClient() *http.Client {
	ctx := context.Background()
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: viper.GetString("CIVO_API_KEY"),
		TokenType:   "Bearer",
	}))

	return client
}

func DeleteCluster(name string) error {

	uri := fmt.Sprintf("%s/%s", baseUri, name)

	// Create a new request using http
	req, err := http.NewRequest(http.MethodDelete, uri, nil)

	// Send req using http Client
	client := getCivoHttpClient()
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}

	log.Infof("status: %s", resp.Status)

	return nil
}

func CreateCluster(name string) error {
	data := url.Values{}
	data.Set("name", name)

	// Create a new request using http
	req, err := http.NewRequest("POST", baseUri, bytes.NewBufferString(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

	// Send req using http Client
	client := getCivoHttpClient()
	_, err = client.Do(req)
	if err != nil {
		log.Println(err)
	}

	return nil
}

func GetClusterNames() []string {
	var clusters []string

	v, err := getClusters()
	if err != nil {
		log.Error(err)
	}
	for _, c := range v.Items {
		clusters = append(clusters, c.Name)
	}
	return clusters
}

func getClusters() (*Clusters, error) {

	client := getCivoHttpClient()
	resp, err := client.Get(baseUri)
	if err != nil {
		log.Error(err)
	}

	body, _ := ioutil.ReadAll(resp.Body)

	v := Clusters{}
	json.Unmarshal(body, &v)
	return &v, nil
}
