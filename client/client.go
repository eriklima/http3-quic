package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"
	"strconv"
	"sync"

	"github.com/lucas-clemente/quic-go/http3"
)

var certPath string
var loopCount int = 10

func init() {
	setupCertPath()
}

func setupCertPath() {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("Failed to get current frame")
	}

	certPath = path.Dir(filename)
}

func main() {
	pool := getCertPool()
	addRootCA(pool)

	client := createClient(pool)
	// url := "https://localhost:3001"
	url := "https://193.167.0.1:4433"
	// url := "https://193.167.0.1:57832"
	// url := "https://193.167.0.2:57832"
	var wg sync.WaitGroup

	wg.Add(loopCount)
	for loopCount > 0 {
		go executeClient(client, url+"/"+strconv.Itoa(loopCount), &wg)
		loopCount--
	}
	wg.Wait()

	fmt.Println("Client finished.")
}

func getCertPool() *x509.CertPool {
	pool, err := x509.SystemCertPool()

	if err != nil {
		log.Fatal(err)
	}

	return pool
}

func addRootCA(certPool *x509.CertPool) {
	caCertPath := path.Join(certPath, "ca.pem")
	caCertRaw, err := os.ReadFile(caCertPath)
	if err != nil {
		panic(err)
	}
	if ok := certPool.AppendCertsFromPEM(caCertRaw); !ok {
		panic("FAILURE: Could not add root ceritificate to pool.")
	}
}

func createClient(pool *x509.CertPool) *http.Client {
	tlsConfig := &tls.Config{
		RootCAs:            pool,
		InsecureSkipVerify: true,
		// KeyLogWriter: ,
	}

	roundTripper := &http3.RoundTripper{
		TLSClientConfig: tlsConfig,
		QuicConfig:      nil,
	}

	defer roundTripper.Close()

	hclient := &http.Client{
		Transport: roundTripper,
	}

	return hclient
}

func executeClient(client *http.Client, url string, wg *sync.WaitGroup) {
	defer wg.Done()

	response, err := client.Get(url)

	if err != nil {
		log.Fatal("Request error: ", err)
	}

	fmt.Printf("Resposta para %s: %#v\n", url, response)

	body := getBody(response)
	fmt.Printf("Body: %s\n", body)
}

func getBody(response *http.Response) []byte {
	body := &bytes.Buffer{}

	_, err := io.Copy(body, response.Body)

	if err != nil {
		log.Fatal(err)
	}

	return body.Bytes()
}
