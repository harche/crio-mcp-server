package redhat

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

func withDoMock(f func(req *http.Request) (*http.Response, error), test func()) {
	orig := Do
	Do = f
	defer func() { Do = orig }()
	test()
}

func TestSearchKCS(t *testing.T) {
	calls := 0
	withDoMock(func(req *http.Request) (*http.Response, error) {
		calls++
		if calls == 1 {
			if req.URL.Host != "sso.redhat.com" {
				t.Fatalf("unexpected host %s", req.URL.Host)
			}
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"access_token":"tok"}`))}, nil
		}
		if calls == 2 {
			if !strings.Contains(req.URL.Path, "/search/kcs") {
				t.Fatalf("unexpected path %s", req.URL.Path)
			}
			if req.Header.Get("Authorization") != "Bearer tok" {
				t.Fatalf("missing auth header")
			}
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("result"))}, nil
		}
		return nil, fmt.Errorf("extra call")
	}, func() {
		out, err := SearchKCS("bug", 10, "off")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out != "result" {
			t.Fatalf("unexpected output %q", out)
		}
	})
}

func TestCVEInfo(t *testing.T) {
	withDoMock(func(req *http.Request) (*http.Response, error) {
		if !strings.Contains(req.URL.Path, "/CVE-1234.json") {
			t.Fatalf("unexpected path %s", req.URL.Path)
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("cve"))}, nil
	}, func() {
		out, err := CVEInfo("CVE-1234")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out != "cve" {
			t.Fatalf("unexpected output %q", out)
		}
	})
}
