package serve

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/crypto/acme/autocert"
)

// Options ...
type Options struct {
	LogOutput interface{}
	Handler   http.Handler
	Domain    string
	DevMode   bool
	Port      int
	CertFile  string
	KeyFile   string
}

type (
	fatalLogger interface {
		Fatal(...interface{})
	}
	printfLogger interface {
		Printf(format string, args ...interface{})
	}
)

func logFatal(opts Options, err error) {
	if opts.LogOutput == nil {
		opts.LogOutput = os.Stderr
	}
	if logger, ok := opts.LogOutput.(fatalLogger); ok {
		logger.Fatal(err)
	} else {
		log.Fatal(err)
	}
	panic("fatal")
}

func logPrintf(opts Options, format string, args ...interface{}) {
	if opts.LogOutput == nil {
		opts.LogOutput = os.Stderr
	}
	if logger, ok := opts.LogOutput.(printfLogger); ok {
		logger.Printf(format, args...)
	} else {
		log.Printf(format, args...)
	}
}

// Serve ...
func Serve(opts Options) {
	if opts.DevMode {
		startServerDev(opts)
	} else if opts.CertFile == "" && opts.KeyFile == "" {
		startServerAutoCert(opts)
	} else {
		startServerHTTP(opts)
		startServerHTTPS(opts)
	}
}

const timeout = time.Second * 15

func startServerDev(opts Options) {
	s := &http.Server{
		Handler:      opts.Handler,
		Addr:         fmt.Sprintf(":%d", opts.Port),
		WriteTimeout: timeout,
		ReadTimeout:  timeout,
	}
	go func() {
		time.Sleep(time.Second / 10)
		logPrintf(opts, "Serving on port %d", opts.Port)
	}()
	logFatal(opts, s.ListenAndServe())
}

func startServerHTTPS(opts Options) {
	go func() {
		s := &http.Server{
			Handler:      opts.Handler,
			Addr:         ":443",
			WriteTimeout: timeout,
			ReadTimeout:  timeout,
		}
		go func() {
			time.Sleep(time.Second / 10)
			logPrintf(opts, "Serving at https://%s", opts.Domain)
		}()
		logFatal(opts, s.ListenAndServeTLS(opts.CertFile, opts.KeyFile))
	}()
}

func startServerHTTP(opts Options) {
	go func() {
		go func() {
			time.Sleep(time.Second / 10)
			logPrintf(opts, "Serving at http://%s", opts.Domain)
		}()
		logFatal(opts, http.ListenAndServe(":80", http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "https://"+opts.Domain+r.RequestURI, 301)
			},
		)))
	}()
}

func startServerAutoCert(opts Options) {
	var certManager autocert.Manager
	certManager = autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(opts.Domain),
		Cache:      autocert.DirCache("certs"),
	}
	s := &http.Server{
		Handler: opts.Handler,
		Addr:    ":443",
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
		},
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
	}
	go func() {
		go func() {
			time.Sleep(time.Second / 10)
			log.Printf("Serving at https://%s", opts.Domain)
		}()
		logFatal(opts, s.ListenAndServeTLS("", ""))
	}()
	logFatal(opts, http.ListenAndServe(":80", certManager.HTTPHandler(nil)))
}
