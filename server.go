package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"
	"sync"

	"github.com/eriklima/http3-quic/utils"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/quic-go/logging"
	"github.com/quic-go/quic-go/qlog"
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
	qlog := flag.Bool("qlog", false, "output a qlog (in the same directory)")
	flag.Parse()

	var wg sync.WaitGroup

	wg.Add(1)
	go startServer(*addr, *qlog, &wg)
	wg.Wait()

	fmt.Println("Server finished")
}

func startServer(addr string, enableQlog bool, wg *sync.WaitGroup) {
	defer wg.Done()

	handler := setupHandler()
	quicConf := setupQuicConfig(enableQlog)

	server := http3.Server{
		Addr:       addr,
		Handler:    handler,
		QuicConfig: quicConf,
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

func setupQuicConfig(enableQlog bool) *quic.Config {
	config := &quic.Config{}

	if enableQlog {
		config.Tracer = qlog.NewTracer(func(_ logging.Perspective, connID []byte) io.WriteCloser {
			err := os.MkdirAll("qlog", 0666)
			if err != nil {
				log.Fatal(err)
			}
			filename := fmt.Sprintf("qlog/server_%x.qlog", connID)
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

func getCertificatePaths() (string, string) {
	return path.Join(certPath, "cert.pem"), path.Join(certPath, "priv.key")
}
