package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
)

var serviceAccountTokenPath = "/var/run/secrets/kubernetes.io/serviceaccount/token"

var apiEndpoint = "https://kubernetes.default.svc/apis/example.com/v1/users"

var caCertPath = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"

func getUsersList(w http.ResponseWriter, r *http.Request) {

	caCert, err := os.ReadFile(caCertPath)
	erChk(err)

	serviceAccountToken, err := os.ReadFile(serviceAccountTokenPath)
	erChk(err)

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		RootCAs: caCertPool,
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	req, err := http.NewRequest("GET", apiEndpoint, nil)
	erChk(err)
	req.Header.Set("Authorization", "Bearer "+string(serviceAccountToken))
	res, err := client.Do(req)
	erChk(err)
	defer res.Body.Close()
	content, err := io.ReadAll(res.Body)
	erChk(err)
	fmt.Println("request received..")
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5500")
	w.Write(content)

}

func usersCounts(w http.ResponseWriter, r *http.Request) {
	var users map[string]interface{}

	caCert, err := os.ReadFile(caCertPath)
	erChk(err)

	serviceAccountToken, err := os.ReadFile(serviceAccountTokenPath)
	erChk(err)

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		RootCAs: caCertPool,
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	req, err := http.NewRequest("GET", apiEndpoint, nil)
	erChk(err)
	req.Header.Set("Authorization", "Bearer "+string(serviceAccountToken))
	res, err := client.Do(req)
	erChk(err)
	defer res.Body.Close()
	content, err := io.ReadAll(res.Body)
	erChk(err)
	fmt.Println("request received..")
	json.Unmarshal(content, &users)
	length := len(users["items"].([]interface{}))
	// Allow all origins - for testing
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write([]byte(strconv.Itoa(length)))

}

func erChk(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {

	http.HandleFunc("/get-users-list", getUsersList)
	http.HandleFunc("/users-count", usersCounts)
	http.ListenAndServe(":8080", nil)
}
