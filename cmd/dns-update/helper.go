package main

import (
    "net/http"
    "fmt"
    "regexp"
    "time"
    "errors"
    "encoding/json"
    "bytes"
)

type ZoneResult struct {
    Result []ZoneInfo `json: "result"`
    Success bool `json: success`
}

type ZoneInfo struct {
    Id string `json: "id"`
    Name string `json: "name"`
}

type DnsResult struct {
    Result []DnsInfo `json: "result"`
    Success bool `json: success`
}

type DnsInfo struct {
    Id string `json: "id"`
    Name string `json: "name"`
    Content string `json: "content"`
}

type Domain struct {
    Name string
    IP string
}

type UpdateParams struct {
    Type string `json: "type"`
    Name string `json: "name"`
    Content string `json: "content"`
    TTL string `json: "ttl"`
    Proxied bool `json: "proxied"`
}

type UpdateResult struct {
    Success bool `json: "success"`
}

func validateMsg(msg string) (bool, error) {
    if len(msg) == 0 {
        return false, errors.New("message is empty")
    }

    r := regexp.MustCompile(`^[a-z0-9\.]+:[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$`)
    if !r.MatchString(msg) {
        return false, fmt.Errorf("message does not match ip address format: %s", msg)
    }

    return true, nil
}

func updateDomain(token string, zoneId string, dnsId string, params UpdateParams) error {
    // Encode
    js, err := json.Marshal(params)
    if err != nil {
        return fmt.Errorf("failed to parse parameters: %v", err)
    }

    req, err := http.NewRequest(
        "PUT",
        fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s", zoneId, dnsId),
        bytes.NewBuffer(js),
    )
    if err != nil {
        return fmt.Errorf("failed to create http request: %v", err)
    }

    // Header set
    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
    req.Header.Set("Content-Type", "application/json")

    // Timeout set
    timeout := time.Duration(5 * time.Second)
    c := &http.Client{
        Timeout: timeout,
    }

    res, err := c.Do(req)
    if err != nil {
        return fmt.Errorf("failed to send request: %v", err)
    }
    defer res.Body.Close()

    if res.StatusCode < 200 || res.StatusCode >= 400 {
        return fmt.Errorf("status code NG: %v", res.StatusCode)
    }

    var result UpdateResult
    if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
        return fmt.Errorf("failed to decode response: %v", err)
    }

    if !result.Success {
        return fmt.Errorf("failed to update")
    }

    return nil
}

func getDomain(message string) (Domain, error) {
    r := regexp.MustCompile(`^([a-z0-9\.]+):([0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3})$`)
    match := r.FindStringSubmatch(message)
    if match == nil {
        return Domain{}, errors.New("could not get domain info from received message")
    }

    return Domain{
        Name: match[1],
        IP: match[2],
    }, nil
}

func createRequest(url string, query map[string]string, token string) (*http.Request, error) {
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create http request: %v", err)
    }

    // Header set
    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
    req.Header.Set("Content-Type", "application/json")

    // Query params set
    q := req.URL.Query()
    for key, val := range query {
        q.Add(key, val)
    }
    req.URL.RawQuery = q.Encode()

    return req, nil
}

func getZoneInfo(token string, domain string) (ZoneInfo, error) {
    req, err := createRequest(
        "https://api.cloudflare.com/client/v4/zones",
        map[string]string{"name":domain},
        token,
    )
    if err != nil {
        return ZoneInfo{}, fmt.Errorf("failed to create http request: %v", err)
    }

    // Timeout set
    timeout := time.Duration(5 * time.Second)
    c := &http.Client{
        Timeout: timeout,
    }

    res, err := c.Do(req)
    if err != nil {
        return ZoneInfo{}, fmt.Errorf("failed to send request: %v", err)
    }
    defer res.Body.Close()

    if res.StatusCode < 200 || res.StatusCode >= 400 {
        return ZoneInfo{}, fmt.Errorf("status code NG: %v", res.StatusCode)
    }

    var result ZoneResult
    if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
        return ZoneInfo{}, fmt.Errorf("failed to decode response: %v", err)
    }

    if len(result.Result) == 0 {
        return ZoneInfo{}, fmt.Errorf("no result found")
    }

    return result.Result[0], nil
}

func getDnsInfo(token string, zoneId string) (DnsInfo, error) {
    req, err := createRequest(
        fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records", zoneId),
        map[string]string{"type":"A"},
        token,
    )
    if err != nil {
        return DnsInfo{}, fmt.Errorf("failed to create http request: %v", err)
    }

    // Timeout set
    timeout := time.Duration(5 * time.Second)
    c := &http.Client{
        Timeout: timeout,
    }

    res, err := c.Do(req)
    if err != nil {
        return DnsInfo{}, fmt.Errorf("failed to send request: %v", err)
    }
    defer res.Body.Close()

    if res.StatusCode < 200 || res.StatusCode >= 400 {
        return DnsInfo{}, fmt.Errorf("status code NG: %v", res.StatusCode)
    }

    var result DnsResult
    if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
        return DnsInfo{}, fmt.Errorf("failed to decode response: %v", err)
    }

    if len(result.Result) == 0 {
        return DnsInfo{}, fmt.Errorf("no result found")
    }

    return result.Result[0], nil
}
