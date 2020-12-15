package main

import (
    "github.com/google/logger"
    "github.com/nats-io/stan.go"
    "os"
)

func postUpdate(message *stan.Msg) {
    msg := string(message.Data)
    if _, err := validateMsg(msg); err != nil {
        logger.Errorf("skipping: %s", err)
        return
    }

    domain, err := getDomain(msg)
    if err != nil {
        logger.Errorf("skipping: %s", err)
    }

    readToken := os.Getenv("READ_TOKEN")
    editToken := os.Getenv("EDIT_TOKEN")
    zoneInfo, err := getZoneInfo(readToken, domain.Name)
    if err != nil {
        logger.Errorf("%v", err)
        return
    }

    dnsInfo, err := getDnsInfo(readToken, zoneInfo.Id)
    if err != nil {
        logger.Errorf("%v", err)
        return
    }

    params := UpdateParams{
        Type: "A",
        Name: domain.Name,
        Content: domain.IP,
        TTL: "120",
        Proxied: true,
    }

    if err := updateDomain(editToken, zoneInfo.Id, dnsInfo.Id, params); err != nil {
        logger.Errorf("%v", err)
        return
    }

    logger.Infof("update: domain: %s, ip: %s", domain.Name, domain.IP)
}
