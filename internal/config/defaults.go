package config

import "time"

const pollIntervalDefault = time.Duration(2 * time.Second)
const reportIntervalDefault = time.Duration(10 * time.Second)
const serverAddressDefault = "127.0.0.1:8080"
const storeIntervalDefault = time.Duration(300 * time.Second)
const storeFileDefault = "/tmp/devops-metrics-db.json"
const restoreDefault = true

// label for Encrypt/Decrypt functions
const RsaLabel = "metrics"
