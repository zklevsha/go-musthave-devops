package config

import (
	"net"
	"time"
)

const pollIntervalDefault = time.Duration(2 * time.Second)
const reportIntervalDefault = time.Duration(10 * time.Second)
const serverAddressDefault = "127.0.0.1:8080"
const storeIntervalDefault = time.Duration(300 * time.Second)
const storeFileDefault = "/tmp/devops-metrics-db.json"
const restoreDefault = true

var trunstedSubnetDefault = net.IPNet{IP: net.IPv4(0, 0, 0, 0), Mask: net.IPv4Mask(0, 0, 0, 0)}

// label for Encrypt/Decrypt functions
const RsaLabel = "metrics"
