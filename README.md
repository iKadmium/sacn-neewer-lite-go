# Sacn-Neewer-Lite-Go

## What it does

The app listens for sACN / E1.31 multicast packets and tries to send them to Neewer lights via Bluetooth. It uses a simple 3 Channel DMX configuration:
```
1: Red
2: Green
3: Blue
```

## Quick Start

Download the binary from the latest release on the right and extract it somewhere. Run `neewer_lite_sacn_go scan` to scan for devices, and follow the on-screen instructions to create a config. Save it, and run `neewer_lite_sacn_go` to listen for packets and send them to your lights.