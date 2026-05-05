package main

import (
	"net/http"
	"time"
)

var githubHTTPClient = &http.Client{Timeout: 30 * time.Second}
