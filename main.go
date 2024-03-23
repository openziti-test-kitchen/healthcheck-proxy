package main

import (
	"context"
	"log"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/openziti/sdk-golang/ziti"
	"github.com/sirupsen/logrus"
)

func main() {
	idFile := getEnvOrDefault("OPENZITI_HEALTHCHECK_IDENTITY", "/opt/openziti/underlay-host-proxy/identity.json")
	port := getEnvOrDefault("OPENZITI_HEALTHCHECK_PROXY_PORT", "2171")
	allowedPathRegex := getEnvOrDefault("OPENZITI_HEALTHCHECK_ALLOWED_PATH", "^.*/ping$")
	slog.Info("allowed path regex set", "regex", allowedPathRegex)
	allowedVerbRegex := getEnvOrDefault("OPENZITI_HEALTHCHECK_ALLOWED_VERB", "GET")
	slog.Info("allowed verb regex set", "regex", allowedVerbRegex)
	searchPattern := getEnvOrDefault("OPENZITI_HEALTHCHECK_SEARCH_REGEX", "(.*)")
	replacePattern := getEnvOrDefault("OPENZITI_HEALTHCHECK_REPLACE_REGEX", "$1")
	slog.Info("host replacement", "replace", searchPattern, "with", replacePattern)
	debug := getEnvOrDefault("OPENZITI_HEALTHCHECK_DEBUG", "")
	if strings.ToLower(debug) == "debug" {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	} else {
		slog.SetLogLoggerLevel(slog.LevelInfo)
	}
	serverCertChain := getEnvOrDefault("OPENZITI_HEALTHCHECK_CERT", "")
	serverKey := getEnvOrDefault("OPENZITI_HEALTHCHECK_KEY", "")

	pathRegex, err := regexp.Compile(allowedPathRegex)
	if err != nil {
		log.Fatalf("error compiling path regex: %v", err)
		return
	}
	verbRegex, err := regexp.Compile(allowedVerbRegex)
	if err != nil {
		log.Fatalf("error compiling verb regex: %v", err)
		return
	}
	searchRegex, err := regexp.Compile(searchPattern)
	if err != nil {
		log.Fatalf("Error compiling search regex:", err)
		return
	}

	ReplaceDefaultHTTPTransport(idFile)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if !pathRegex.MatchString(r.URL.Path) {
			slog.Warn("request disallowed.", "path", r.URL.Path)
			return
		}

		if !verbRegex.MatchString(r.Method) {
			slog.Warn("request disallowed.", "verb", r.Method)
			return
		}

		scheme := "http"
		if r.URL.Scheme != "" {
			scheme = r.URL.Scheme
		}

		destHost := searchRegex.ReplaceAllString(r.Host, replacePattern)

		dest := &url.URL{
			Scheme: scheme,
			Host:   destHost,
		}
		slog.Debug("proxying request to OpenZiti", "from", r.Host, "to", dest.Host)
		proxy := httputil.NewSingleHostReverseProxy(dest)
		proxy.ServeHTTP(w, r)
	})

	if serverCertChain != "" && serverKey != "" {
		slog.Info("server is listening", "scheme", "https", "port", port)
		log.Fatal(http.ListenAndServeTLS(":"+port, serverCertChain, serverKey, nil))
	} else {
		slog.Info("server is listening", "scheme", "http", "port", port)
		log.Fatal(http.ListenAndServe(":"+port, nil))
	}
}

func ReplaceDefaultHTTPTransport(idFile string) {
	ctx := ContextFromFile(idFile)
	zitiDialContext := ZitiDialContext{context: ctx}
	zitiTransport := http.DefaultTransport.(*http.Transport).Clone() // copy default transport
	zitiTransport.DialContext = zitiDialContext.Dial
	zitiTransport.TLSClientConfig.InsecureSkipVerify = true
	http.DefaultTransport = zitiTransport
}

func ContextFromFile(idFile string) ziti.Context {
	cfg, err := ziti.NewConfigFromFile(idFile)
	if err != nil {
		logrus.Fatal(err)
	}

	ctx, err := ziti.NewContext(cfg)
	if err != nil {
		logrus.Fatal(err)
	}
	return ctx
}

func getEnvOrDefault(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}

type ZitiDialContext struct {
	context ziti.Context
}

func (dc *ZitiDialContext) Dial(_ context.Context, _ string, addr string) (net.Conn, error) {
	service := strings.Split(addr, ":")[0] // will always get passed host:port
	return dc.context.Dial(service)
}
