package civo

import (
	"bytes"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"
)

type Clusters struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
	Pages   int `json:"pages"`
	Items   []struct {
		ID                string        `json:"id"`
		Name              string        `json:"name"`
		Version           string        `json:"version"`
		Status            string        `json:"status"`
		Ready             bool          `json:"ready"`
		NumTargetNodes    int           `json:"num_target_nodes"`
		TargetNodesSize   string        `json:"target_nodes_size"`
		BuiltAt           time.Time     `json:"built_at"`
		Kubeconfig        string        `json:"kubeconfig"`
		KubernetesVersion string        `json:"kubernetes_version"`
		APIEndpoint       string        `json:"api_endpoint"`
		DNSEntry          string        `json:"dns_entry"`
		Tags              []interface{} `json:"tags"`
		CreatedAt         time.Time     `json:"created_at"`
		Instances         []struct {
			Hostname   string    `json:"hostname"`
			Size       string    `json:"size"`
			Region     string    `json:"region"`
			CreatedAt  time.Time `json:"created_at"`
			Status     string    `json:"status"`
			FirewallID string    `json:"firewall_id"`
			PublicIP   string    `json:"public_ip"`
			Tags       []string  `json:"tags"`
		} `json:"instances"`
		InstalledApplications []struct {
			Application   string      `json:"application"`
			Title         interface{} `json:"title"`
			Version       string      `json:"version"`
			Dependencies  interface{} `json:"dependencies"`
			Maintainer    string      `json:"maintainer"`
			Description   string      `json:"description"`
			PostInstall   string      `json:"post_install"`
			Installed     bool        `json:"installed"`
			URL           string      `json:"url"`
			Category      string      `json:"category"`
			UpdatedAt     time.Time   `json:"updated_at"`
			ImageURL      string      `json:"image_url"`
			Plan          interface{} `json:"plan"`
			Configuration struct {
			} `json:"configuration"`
		} `json:"installed_applications"`
	} `json:"items"`
}

type webhook struct {
	Event   string `json:"event"`
	Payload struct {
		ID                    string        `json:"id"`
		Name                  string        `json:"name"`
		Version               string        `json:"version"`
		Status                string        `json:"status"`
		Ready                 bool          `json:"ready"`
		NumTargetNodes        int           `json:"num_target_nodes"`
		TargetNodesSize       string        `json:"target_nodes_size"`
		BuiltAt               interface{}   `json:"built_at"`
		Kubeconfig            interface{}   `json:"kubeconfig"`
		KubernetesVersion     string        `json:"kubernetes_version"`
		APIEndpoint           interface{}   `json:"api_endpoint"`
		DNSEntry              string        `json:"dns_entry"`
		Tags                  []interface{} `json:"tags"`
		CreatedAt             time.Time     `json:"created_at"`
		Instances             []interface{} `json:"instances"`
		InstalledApplications []interface{} `json:"installed_applications"`
	} `json:"payload"`
}

func ServeHandler() {
	http.HandleFunc("/", handleWebhook)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	webhookData := webhook{}
	err := json.NewDecoder(r.Body).Decode(&webhookData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Printf("got webhook payload: %+v", webhookData)
}

func GetCluster(name string) (string, error) {
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

func DeleteCluster(name string) error {
	uri := fmt.Sprintf("https://api.civo.com/v2/kubernetes/clusters/%s", name)

	// Create a Bearer string by appending string access token
	var bearer = "Bearer " + os.Getenv("CIVO_API_KEY")

	// Create a new request using http
	req, err := http.NewRequest(http.MethodDelete, uri, nil)
	req.Header.Add("Authorization", bearer)

	// Send req using http Client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}

	log.Infof("status: %s", resp.Status)

	return nil
}

func CreateCluster(name string) error {
	uri := "https://api.civo.com/v2/kubernetes/clusters"

	// Create a Bearer string by appending string access token
	var bearer = "Bearer " + os.Getenv("CIVO_API_KEY")

	data := url.Values{}
	data.Set("name", name)

	// Create a new request using http
	req, err := http.NewRequest("POST", uri, bytes.NewBufferString(data.Encode()))
	req.Header.Add("Authorization", bearer)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

	// Send req using http Client
	client := &http.Client{}
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
	url := "https://api.civo.com/v2/kubernetes/clusters"

	// Create a Bearer string by appending string access token
	var bearer = "Bearer " + os.Getenv("CIVO_API_KEY")

	// Create a new request using http
	req, err := http.NewRequest("GET", url, nil)

	// add authorization header to the req
	req.Header.Add("Authorization", bearer)

	// Send req using http Client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error on response.\n[ERRO] -", err)
	}

	body, _ := ioutil.ReadAll(resp.Body)

	v := Clusters{}
	json.Unmarshal(body, &v)
	return &v, nil
}
