# DCS-ICE (Intelligent Command Engine)

A system for evaluating battlefield events from DCS World missions using rule-based logic.

## Overview

DCS-ICE receives battlefield events and facts from DCS World missions (via Exports.lua), evaluates them using the grule rule engine, and returns structured actions (e.g., spawn a unit, change ROE, trigger audio).

## Features

- HTTP API for receiving and evaluating facts
- WebSocket support for real-time communication
- Rule-based evaluation using grule-rule-engine
- Clean, modular architecture

## Getting Started

### Prerequisites

- Go 1.18 or later
- DCS World installation (for integration)

   ```
   go mod download
   ```

3. Build the server:
   ```
   go build -o bin/dcs-ice cmd/server/main.go
   ```

### Running the Server

```
./bin/dcs-ice --port 8080 --rules ./config/rules
```

## API Documentation

### POST /facts

Evaluates a set of facts against the loaded rules and returns matched rules and actions.

#### Request

```json
{
  "facts": [
    {
      "event": "unit_destroyed",
      "unit": "convoy_alpha",
      "zone": "BRAVO",
      "alertLevel": "red"
    }
  ]
}
```

#### Response

```json
{
  "matchedRules": ["UnitDestroyedInZoneBravo"],
  "actions": [
    {
      "type": "spawn",
      "target": "reinforcement",
      "parameters": {
        "location": "BRAVO",
        "type": "SAM"
      }
    }
  ]
}
```

## License

[MIT](LICENSE)
