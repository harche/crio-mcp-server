package redhat

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

var Do = http.DefaultClient.Do

// AccessToken exchanges an offline token for a short-lived access token.
func AccessToken(offlineToken string) (string, error) {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("client_id", "rhsm-api")
	data.Set("refresh_token", offlineToken)
	req, err := http.NewRequest("POST", "https://sso.redhat.com/auth/realms/redhat-external/protocol/openid-connect/token", strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token request failed: %s", resp.Status)
	}
	var result struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}
	if result.AccessToken == "" {
		return "", fmt.Errorf("access_token not found")
	}
	return result.AccessToken, nil
}

// SearchKCS queries the Red Hat Knowledge Base for articles matching the query.
func SearchKCS(query string, rows int, offlineToken string) (string, error) {
	token, err := AccessToken(offlineToken)
	if err != nil {
		return "", err
	}
	if rows <= 0 {
		rows = 20
	}
	endpoint := "https://access.redhat.com/hydra/rest/search/kcs?format=json&q=" + url.QueryEscape(query) + "&rows=" + fmt.Sprint(rows)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("search request failed: %s", resp.Status)
	}
	return string(body), nil
}

// CVEInfo fetches details for a given CVE ID using the Security Data API.
func CVEInfo(cveID string) (string, error) {
	endpoint := "https://access.redhat.com/hydra/rest/securitydata/cve/" + url.QueryEscape(cveID) + ".json"
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}
	resp, err := Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("cve request failed: %s", resp.Status)
	}
	return string(body), nil
}
