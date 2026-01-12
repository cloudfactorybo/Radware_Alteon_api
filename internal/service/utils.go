package service

import (
	"net/url"
	"strings"
)

// extractIPFromURL extrae la IP de una URL
func extractIPFromURL(urlStr string) string {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}

	// Obtener el host (puede incluir puerto)
	host := parsedURL.Host

	// Si tiene puerto, quitarlo
	if colonIndex := strings.Index(host, ":"); colonIndex != -1 {
		host = host[:colonIndex]
	}

	return host
}
