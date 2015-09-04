package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	apacheFormatPattern = "%s %.4f - [%s] \"%s\" %d %d %s %s\n"
	version             = "1.0.0"
)

func main() {
	address := flag.String("address", "127.0.0.1", "The address to listen on")
	path := flag.String("path", "", "Path to the document root")
	port := flag.String("port", "8080", "The port to listen on")
	tls := flag.Bool("tls", false, "Use TLS")
	cert := flag.String("cert", "cert.pem", "The TLS certificate to use")
	key := flag.String("key", "key.pem", "The TLS key to use.")
	flag.Parse()

	if *tls == true {
		if *port == "8080" {
			*port = "8443"
		}
	}

	addrFmt := fmt.Sprintf("%s:%s", *address, *port)
	mux := http.DefaultServeMux
	mux.Handle("/", http.FileServer(http.Dir(*path)))

	println("Starting server with document root: '" + filepath.Base(*path) + "'")

	if *tls == true {
		println("	at https://" + addrFmt + "/")
	} else {
		println("	at http://" + addrFmt + "/")
	}
	println("dserve", version, "built with go version: ", runtime.Version())

	loggingHandler := NewApacheLoggingHandler(mux, os.Stderr)
	server := &http.Server{
		Addr:    addrFmt,
		Handler: loggingHandler,
	}

	if *tls == true {
		if err := server.ListenAndServeTLS(*cert, *key); err != nil {
			log.Fatal(err)
		}
	} else {
		if err := server.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}
}

type ApacheLogRecord struct {
	http.ResponseWriter

	ip                    string
	time                  time.Time
	method, uri, protocol string
	status                int
	responseBytes         int64
	referrer              string
	userAgent             string
	elapsedTime           time.Duration
}

func (r *ApacheLogRecord) Log(out io.Writer) {
	timeFormatted := r.time.Format("02/Jan/2006 03:04:05")
	requestLine := fmt.Sprintf("%s %s %s", r.method, r.uri, r.protocol)
	fmt.Fprintf(out, apacheFormatPattern,
		r.ip,
		r.elapsedTime.Seconds(),
		timeFormatted,
		requestLine,
		r.status,
		r.responseBytes,
		r.referrer,
		r.userAgent)
}

func (r *ApacheLogRecord) Write(p []byte) (int, error) {
	written, err := r.ResponseWriter.Write(p)
	r.responseBytes += int64(written)
	return written, err
}

func (r *ApacheLogRecord) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

type ApacheLoggingHandler struct {
	handler http.Handler
	out     io.Writer
}

func NewApacheLoggingHandler(handler http.Handler, out io.Writer) http.Handler {
	return &ApacheLoggingHandler{
		handler: handler,
		out:     out,
	}
}

func (h *ApacheLoggingHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	clientIP := r.RemoteAddr
	if colon := strings.LastIndex(clientIP, ":"); colon != -1 {
		clientIP = clientIP[:colon]
	}

	record := &ApacheLogRecord{
		ResponseWriter: rw,
		ip:             clientIP,
		time:           time.Time{},
		method:         r.Method,
		uri:            r.RequestURI,
		protocol:       r.Proto,
		status:         http.StatusOK,
		referrer:       r.Referer(),
		userAgent:      r.UserAgent(),
		elapsedTime:    time.Duration(0),
	}

	startTime := time.Now()
	h.handler.ServeHTTP(record, r)
	finishTime := time.Now()

	record.time = finishTime.UTC()
	record.elapsedTime = finishTime.Sub(startTime)

	record.Log(h.out)
}
