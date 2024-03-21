package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"os"
)

var serviceAccountTokenPath = "/var/run/secrets/kubernetes.io/serviceaccount/token"

var apiEndpoint = "https://kubernetes.default.svc/api/v1/pods"

var caCertPath = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"

func getPods(w http.ResponseWriter, r *http.Request) {
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
	fmt.Println(content)
	w.Write(content)

}

func erChk(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {

	http.HandleFunc("/pods", getPods)

	http.ListenAndServe(":8080", nil)
}
