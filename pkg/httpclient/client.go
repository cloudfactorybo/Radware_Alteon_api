package httpclient

import (
	"crypto/tls"
	"net/http"
	"net/http/cookiejar"
	"time"
)

type Client struct {
	*http.Client
}

func NewSecureClient(insecureSkipVerify bool) *Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecureSkipVerify,
		},
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
	}

	// Cookie jar para mantener viva la sesión del Alteon. Al autenticar con Basic
	// Auth, el equipo responde con una cookie de sesión; el jar la guarda y la
	// reenvía en las siguientes peticiones, evitando que el Alteon haga login en
	// cada request (lo que le consume CPU y ocupa slots de sesión). Las cookies se
	// guardan por host, así que cada Alteon conserva su propia sesión.
	// cookiejar.New(nil) nunca devuelve error con opciones nil.
	jar, _ := cookiejar.New(nil)

	return &Client{
		Client: &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
			Jar:       jar,
		},
	}
}
