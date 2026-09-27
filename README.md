# HealthCheckMonitor

[![Go Report Card](https://goreportcard.com/badge/github.com/example/healthcheckmonitor)](https://goreportcard.com/report/github.com/example/healthcheckmonitor)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

## Overview

HealthCheckMonitor is a production‑ready Go service that periodically checks the health of configurable HTTP endpoints and exposes their status via a JSON API. 
Features:

- Configurable list of services (name, URL, interval) using a YAML file. 
- Concurrent health checks with per‑service intervals. 
- Centralised, structured logging with Logrus. 
- Graceful shutdown on SIGINT/SIGTERM. 
- Simple HTTP API (`/health` & `/health/{name}`) to query current status. 
- Fully typed, unit‑tested core monitor logic.

## Prerequisites

- Go 1.22+ 
- `git` 

## Installation