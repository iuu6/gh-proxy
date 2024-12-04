package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/proxy"
)

type Config struct {
	WhiteList []string `json:"white_list"`
	BlackList []string `json:"black_list"`
	SizeLimit int64    `json:"size_limit"`
	Socks5    string   `json:"socks5"`
}

var (
	exp1      = regexp.MustCompile(`^(?:https?://)?github\.com/(?P<author>.+?)/(?P<repo>.+?)/(?:releases|archive)/.*$`)
	exp2      = regexp.MustCompile(`^(?:https?://)?github\.com/(?P<author>.+?)/(?P<repo>.+?)/(?:blob|raw)/.*$`)
	exp3      = regexp.MustCompile(`^(?:https?://)?github\.com/(?P<author>.+?)/(?P<repo>.+?)/(?:info|git-).*$`)
	exp4      = regexp.MustCompile(`^(?:https?://)?raw\.(?:githubusercontent|github)\.com/(?P<author>.+?)/(?P<repo>.+?)/.+?/.+$`)
	exp5      = regexp.MustCompile(`^(?:https?://)?gist\.(?:githubusercontent|github)\.com/(?P<author>.+?)/.+?/.+$`)
	config    Config
	transport *http.Transport
)

func main() {
	config_path := flag.String("c", "config.json", "Path to the config file")
	flag.Parse()
	config = *readConfig(*config_path)

	http.HandleFunc("/", handler)
	http.HandleFunc("/favicon.ico", iconHandler)
	log.Println("Listening on :5340")
	http.ListenAndServe(":5340", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	u := r.URL.Path[1:]

	// 修正 URL，确保路径开头有两个斜杠
	u = strings.Replace(u, "https:/", "https://", 1)
	u = strings.Replace(u, "http:/", "http://", 1)

	if u == "" {
		index(w, r)
		return
	}

	log.Printf("Received URL: %s\n", u)

	// 解码 URL
	u, err := url.PathUnescape(u)
	if err != nil {
		http.Error(w, "Failed to decode URL.", http.StatusInternalServerError)
		return
	}

	if m := checkURL(u); m != nil {
		// For demonstration, just printing the matched groups
		m["repo"] = strings.TrimSuffix(m["repo"], ".git")
		log.Printf("Author: %s, Repo: %s\n", m["author"], m["repo"])
		if allowDownload(m["author"], m["repo"]) {
			proxyHandler(w, r)
		} else {
			http.Error(w, "Download not allowed.", http.StatusForbidden)
		}
	} else {
		http.Error(w, "Invalid input.", http.StatusForbidden)
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	// 提供 index.html 文件内容
	http.ServeFile(w, r, "index.html")
}

func iconHandler(w http.ResponseWriter, r *http.Request) {
	// Return favicon.ico
	// You need to replace this with your actual favicon
	http.ServeFile(w, r, "favicon.ico")
}

func checkURL(u string) map[string]string {
	for _, exp := range []*regexp.Regexp{exp1, exp2, exp3, exp4, exp5} {
		match := exp.FindStringSubmatch(u)
		if match != nil {
			result := make(map[string]string)
			for i, name := range exp.SubexpNames() {
				if i > 0 && i <= len(match) {
					result[name] = match[i]
				}
			}
			return result
		}
	}
	return nil
}

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	u := r.URL.Path[1:]
	// 修正 URL，确保路径开头有两个斜杠
	u = strings.Replace(u, "https:/", "https://", 1)
	u = strings.Replace(u, "http:/", "http://", 1)
	// 获取第三个/之前的内容，即https://github.com
	url, _ := url.Parse(strings.Join(strings.Split(u, "/")[:3], "/"))
	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.Transport = transport

	// 删除第三个/及之前的内容，即path
	r.URL.Path = "/" + strings.Join(strings.Split(r.URL.Path, "/")[3:], "/")
	r.Host = url.Host
	r.URL.Host = url.Host
	r.URL.Scheme = url.Scheme
	r.Header.Set("Host", url.Host)

	proxy.ServeHTTP(w, r)
}

func allowDownload(author, repo string) bool {
	// Check blacklist
	for _, entry := range config.BlackList {
		if entry == author || entry == author+"/"+repo {
			return false
		}
	}

	// Check whitelist
	if len(config.WhiteList) > 0 {
		for _, entry := range config.WhiteList {
			if entry == author || entry == author+"/"+repo {
				return true
			}
		}
		// If whitelist is defined but entry not found, disallow download
		return false
	}

	// If whitelist is not defined, allow download by default
	return true
}

func readConfig(filename string) *Config {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Error opening config file: %v", err)
		return nil
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	config := Config{}
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatalf("Error decoding config file: %v", err)
		return nil
	}
	if config.Socks5 == "" {
		log.Println("Using direct connection.")
		transport = &http.Transport{
			MaxIdleConns:        10,
			IdleConnTimeout:     30 * time.Second,
			TLSHandshakeTimeout: 10 * time.Second,
		}
	} else {
		log.Println("Using SOCKS5 proxy:", config.Socks5)
		dialer, err := proxy.SOCKS5("tcp", config.Socks5, nil, proxy.Direct)
		if err != nil {
			log.Fatalf("Failed to connect to the proxy: %v", err)
		}
		transport = &http.Transport{
			Dial:                dialer.Dial,
			MaxIdleConns:        10,
			IdleConnTimeout:     30 * time.Second,
			TLSHandshakeTimeout: 10 * time.Second,
		}
	}
	return &config
}
