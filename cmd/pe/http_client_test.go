package main

import "testing"

func TestGitHubHTTPClientHasTimeout(t *testing.T) {
	if githubHTTPClient.Timeout <= 0 {
		t.Fatal("githubHTTPClient has no timeout")
	}
}
