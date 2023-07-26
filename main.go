package main

import (
	"fmt"
	"bufio"
	"net/http"
	"os"
	"strings"
	"text/template"

	"github.com/julienschmidt/httprouter"
	"net"
)

func main() {
	router := httprouter.New()
	router.GET("/", homeHandler)
	router.POST("/scan", scanHandler)

	fmt.Println("Server started on :80")
	http.ListenAndServe(":80", router)
}

func homeHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	renderTemplate(w, "home.html", nil)
}

func scanHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	r.ParseForm()
	domain := strings.TrimSpace(r.Form.Get("domain"))

	subdomains, err := readSubdomainsFromFile("subdomains.txt")
	if err != nil {
		http.Error(w, "Error reading subdomains file", http.StatusInternalServerError)
		return
	}

	results := make(map[string]string)

	for _, subdomain := range subdomains {
		fullDomain := fmt.Sprintf("%s.%s", subdomain, domain)
		ip, err := lookupIP(fullDomain)
		if err == nil {
			results[fullDomain] = ip
		}
	}

	data := struct {
		Domain  string
		Results map[string]string
	}{
		Domain:  domain,
		Results: results,
	}

	renderTemplate(w, "results.html", data)
}

func lookupIP(domain string) (string, error) {
	ips, err := net.LookupIP(domain)
	if err != nil {
		return "", err
	}
	return ips[0].String(), nil
}

func readSubdomainsFromFile(filename string) ([]string, error) {
	var subdomains []string

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		subdomain := strings.TrimSpace(scanner.Text())
		if subdomain != "" {
			subdomains = append(subdomains, subdomain)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return subdomains, nil
}

func renderTemplate(w http.ResponseWriter, filename string, data interface{}) {
	tmpl, err := template.ParseFiles(filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, data)
}
