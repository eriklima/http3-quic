package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"
	"strconv"
	"sync"

	"github.com/eriklima/http3-quic/utils"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/quic-go/logging"
	"github.com/quic-go/quic-go/qlog"
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
	url := flag.String("url", "localhost:4433", "IP:PORT for HTTP3 server")
	qlog := flag.Bool("qlog", false, "output a qlog (in the same directory)")
	flag.Parse()

	pool := getCertPool()
	addRootCA(pool)

	client := createClient(pool, *qlog)

	var wg sync.WaitGroup

	completedUrl := "https://" + *url

	wg.Add(loopCount)
	for loopCount > 0 {
		fmt.Printf("Call %s/%d \n", completedUrl, loopCount)
		go executeClient(client, completedUrl+"/"+strconv.Itoa(loopCount), &wg)
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

func createClient(pool *x509.CertPool, enableQlog bool) *http.Client {
	tlsConfig := &tls.Config{
		RootCAs:            pool,
		InsecureSkipVerify: true,
		// KeyLogWriter: ,
	}

	quicConfig := setupQuicConfig(enableQlog)

	roundTripper := &http3.RoundTripper{
		TLSClientConfig: tlsConfig,
		QuicConfig:      quicConfig,
	}

	defer roundTripper.Close()

	hclient := &http.Client{
		Transport: roundTripper,
	}

	return hclient
}

func setupQuicConfig(enableQlog bool) *quic.Config {
	config := &quic.Config{}

	if enableQlog {
		config.Tracer = qlog.NewTracer(func(_ logging.Perspective, connID []byte) io.WriteCloser {
			err := os.MkdirAll("qlog", 0666)
			if err != nil {
				log.Fatal(err)
			}
			filename := fmt.Sprintf("qlog/client_%x.qlog", connID)
			f, err := os.Create(filename)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Creating qlog file %s.\n", filename)
			return utils.NewBufferedWriteCloser(bufio.NewWriter(f), f)
		})

		fmt.Println("Qlog enabled!")
	}

	return config
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
