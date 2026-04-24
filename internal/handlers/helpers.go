package handlers

import "strings"

// slugFromHost extrai o primeiro segmento do hostname, ignorando a porta.
// "company.localhost:4000" → "company"
// "localhost:4000"         → "localhost"
func slugFromHost(host string) string {
	hostname := strings.Split(host, ":")[0]
	return strings.Split(hostname, ".")[0]
}
