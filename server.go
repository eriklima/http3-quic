package main

import (
	"flag"
	"fmt"
	"net/http"
	"path"
	"runtime"
	"sync"

	"github.com/lucas-clemente/quic-go/http3"
)

var certPath string

func init() {
	setupCertPath()
}

func setupCertPath() {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("Failed to get current frame")
	}

	certPath = path.Join(path.Dir(filename), "keys")
}

func main() {
	addr := flag.String("addr", "localhost:4433", "Server listening to IP:PORT")
	flag.Parse()

	var wg sync.WaitGroup

	wg.Add(1)
	go startServer(*addr, &wg)
	wg.Wait()

	fmt.Println("Server finished")
}

func startServer(addr string, wg *sync.WaitGroup) {
	defer wg.Done()

	handler := setupHandler()

	server := http3.Server{
		Addr:    addr,
		Handler: handler,
	}

	pem, key := getCertificatePaths()

	fmt.Printf("Server listening on %s\n", addr)

	err := server.ListenAndServeTLS(pem, key)

	if err != nil {
		fmt.Printf("Server error: %s\n", err)
	}
}

func setupHandler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Request recebido: %s\n", r.RequestURI)
		// w.Write([]byte(r.RequestURI))
		w.Write([]byte(r.URL.String()))
	})

	return mux
}

func getCertificatePaths() (string, string) {
	return path.Join(certPath, "cert.pem"), path.Join(certPath, "priv.key")
}
