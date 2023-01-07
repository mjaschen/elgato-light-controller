# elc - Elgato Light Controller

This is a simple CLI programm for controlling Elgato Key Lights.

I searched for a simple project for my very first steps in the Go programming language and decided a CLI controller for my Elgato light would be a good choice.

One could say this program is nothing else than a strongly typed shell script.

Tested with an Elgato Key Light Air.

## Usage

```shell
# Ask Keylight for information about itself
% elc
Product Name          : Elgato Key Light Air
Hardware Board Type   : 200
Hardware Revision     : 1
MAC Address           : 3C:6A:9D:AA:BB:CC
Firmware Build Number : 218
Firmware Version      : 1.0.3
Serial Number         : CW00L0A00000
Display Name          :
Features              : lights
Wifi SSID             : wifi_ssid_name
Wifi Frequency MHz    : 2400
Wifi RSSI             : -38

# Fetch current status ("status" or "s")
% elc status
State             : off
Brightness        : 40 %
Color Temperature : 5525 K

# Switch on/off ("on" or "off", resp. "1" or "0")
% elc on
% elc off
% elc 1
% elc 0

# Set brightness ("brightness" or "b"), value between 0 and 100
% elc brightness 50
% elc b 100

# Set color temperature ("temperature" or "t", value between 2900 and 7000)
% elc temperature 5500
% elc t 3300

# Get status in JSON format
% elc --format=json status
{"state":false,"brightness":40,"temperature":5525}
```


## Libraries

- https://github.com/go-resty/resty
- https://github.com/akamensky/argparse
- https://github.com/tidwall/gjson

## References

- https://gitlab.com/obviate.io/pyleglight/
