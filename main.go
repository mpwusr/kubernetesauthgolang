package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {
	apiServer := "https://api.openshift.example.com:6443"
	namespace := "my-namespace"
	saName := "my-sa"
	userToken := "your-admin-token"
	caPath := "/path/to/ca.crt"

	// Load CA cert
	caCert, err := ioutil.ReadFile(caPath)
	if err != nil {
		panic(err)
	}
	caPool := x509.NewCertPool()
	caPool.AppendCertsFromPEM(caCert)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{RootCAs: caPool},
	}
	client := &http.Client{Transport: tr}

	headers := map[string]string{
		"Authorization": "Bearer " + userToken,
		"Accept":        "application/json",
	}

	// Get SA
	saUrl := fmt.Sprintf("%s/api/v1/namespaces/%s/serviceaccounts/%s", apiServer, namespace, saName)
	req, _ := http.NewRequest("GET", saUrl, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, _ := client.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)
	var saData map[string]interface{}
	json.Unmarshal(body, &saData)

	// Extract secret
	secrets := saData["secrets"].([]interface{})
	var secretName string
	for _, s := range secrets {
		name := s.(map[string]interface{})["name"].(string)
		if len(name) > 0 && contains(name, "token") {
			secretName = name
			break
		}
	}

	// Get secret data
	secretUrl := fmt.Sprintf("%s/api/v1/namespaces/%s/secrets/%s", apiServer, namespace, secretName)
	req2, _ := http.NewRequest("GET", secretUrl, nil)
	for k, v := range headers {
		req2.Header.Set(k, v)
	}
	resp2, _ := client.Do(req2)
	body2, _ := ioutil.ReadAll(resp2.Body)
	var secretData map[string]interface{}
	json.Unmarshal(body2, &secretData)
	tokenB64 := secretData["data"].(map[string]interface{})["token"].(string)
	decoded, _ := base64.StdEncoding.DecodeString(tokenB64)

	fmt.Println("Retrieved Token:")
	fmt.Println(string(decoded))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s[:len(substr)] == substr)
}
