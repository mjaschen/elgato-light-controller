// TODO:
//
// - MQTT Publish

package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/akamensky/argparse"
	"github.com/go-resty/resty/v2"
	"github.com/tidwall/gjson"
)

type AccessoryWifiInfo struct {
	SSID         string `json:"ssid"`
	FrequencyMHz int    `json:"frequencyMHz"`
	RSSI         int    `json:"rssi"`
}

type AccessoryInfo struct {
	ProductName         string            `json:"productName"`
	HardwareBoardType   int               `json:"hardwareBoardType"`
	HardwareRevision    int               `json:"hardwareRevision"`
	MacAddress          string            `json:"macAddress"`
	FirmwareBuildNumber int               `json:"firmwareBuildNumber"`
	FirmwareVersion     string            `json:"firmwareVersion"`
	SerialNumber        string            `json:"serialNumber"`
	DisplayName         string            `json:"displayName"`
	Features            []string          `json:"features"`
	WifiInfo            AccessoryWifiInfo `json:"wifi-info"`
}

type LightStatus struct {
	OnOffState  bool `json:"state"`
	Brightness  int  `json:"brightness"`
	Temperature int  `json:"temperature"`
}

type Arguments struct {
	Url     string
	Command string
	Value   int
	Verbose bool
	Format  string
}

var http_client resty.Client
var arguments Arguments

func main() {
	parser := argparse.NewParser("elc", "Control Elgato Lights")
	command := parser.StringPositional(&argparse.Options{Default: "info", Help: "Command: info, status (s), on (1), off (0), brightness (b), temperature (t)"})
	value := parser.IntPositional(&argparse.Options{Default: -1, Required: false, Help: "Value for brightness (0-100) and color (2900-7000) commands"})
	url := parser.String("u", "url", &argparse.Options{Required: false, Help: "URL for Light, e.g. http://keylight.local:9123 or http://10.0.0.10:9123; can be omitted if environment variable ELGATO_LIGHT_URL is set with a valid URL"})
	verbose := parser.Flag("v", "verbose", &argparse.Options{Required: false, Help: "Show response data for sent commands"})
	format := parser.String("f", "format", &argparse.Options{Required: false, Help: "Select output format: text (default) or json"})

	error := parser.Parse(os.Args)

	if error != nil {
		fmt.Print(parser.Usage(error))
		os.Exit(1)
	}

	arguments.Command = *command
	arguments.Value = *value
	arguments.Verbose = *verbose
	arguments.Format = *format

	envUrl, envUrlFound := os.LookupEnv("ELGATO_LIGHT_URL")
	if *url != "" {
		arguments.Url = *url
	} else if envUrlFound {
		arguments.Url = envUrl
	} else {
		fmt.Print("Light URL not specified (option --url or environment variable ELGATO_LIGHT_URL)")
		os.Exit(1)
	}

	init_http_client(arguments.Url)

	if arguments.Command == "info" {
		response := http_fetch("/elgato/accessory-info")
		print_accessory_info(parse_info_response(*response))
	}

	if arguments.Command == "status" || arguments.Command == "s" {
		print_light_status(parse_status_response(*http_fetch("/elgato/lights")))
	}

	if arguments.Command == "on" || arguments.Command == "1" {
		response := light_on()
		if arguments.Verbose {
			print_light_status(parse_status_response(*response))
		}
	}

	if arguments.Command == "off" || arguments.Command == "0" {
		response := light_off()
		if arguments.Verbose {
			print_light_status(parse_status_response(*response))
		}
	}

	if arguments.Command == "brightness" || arguments.Command == "b" {
		response := light_brightness(arguments.Value)
		if arguments.Verbose {
			print_light_status(parse_status_response(*response))
		}
	}

	if arguments.Command == "temperature" || arguments.Command == "t" {
		response := light_color(arguments.Value)
		if arguments.Verbose {
			print_light_status(parse_status_response(*response))
		}
	}
}

func init_http_client(url string) {
	if url == "" {
		fmt.Println("Hostname/IP Address missing")
		os.Exit(1)
	}

	// https://github.com/go-resty/resty
	http_client = *resty.New()
	http_client.SetScheme("http").
		SetHostURL(url)
}

func http_fetch(path string) *resty.Response {
	response, error := http_client.
		R().
		EnableTrace().
		Get(path)

	if error != nil {
		fmt.Println("Error connecting to light: ", error)
		os.Exit(1)
	}

	return response
}

func http_put(path string, body string) *resty.Response {
	response, error := http_client.
		R().
		SetBody(body).
		Put("/elgato/lights")

	if error != nil {
		fmt.Println("Error connecting to light: ", error)
		os.Exit(1)
	}

	return response
}
func light_on() *resty.Response {
	return http_put("/elgato/lights", `{"numberOfLights":1,"lights":[{"on":1}]}`)
}

func light_off() *resty.Response {
	return http_put("/elgato/lights", `{"numberOfLights":1,"lights":[{"on":0}]}`)
}

func light_brightness(brightness int) *resty.Response {
	if brightness < 0 || brightness > 100 {
		fmt.Println("Brightness value out of range (valid values: 0-100)")
		os.Exit(1)
	}

	body := `{"numberOfLights":1,"lights":[{"brightness":%d}]}`

	return http_put("/elgato/lights", fmt.Sprintf(body, brightness))
}

func light_color(temperature int) *resty.Response {
	if temperature < 2900 || temperature > 7000 {
		fmt.Println("Color temperature out of range (valid values: 2900-7000)")
		os.Exit(1)
	}

	body := `{"numberOfLights":1,"lights":[{"temperature":%d}]}`

	return http_put("/elgato/lights", fmt.Sprintf(body, kelvin_to_elgato_color_value(temperature)))
}

// see https://gitlab.com/obviate.io/pyleglight/-/blob/master/leglight/leglight.py
func kelvin_to_elgato_color_value(kelvin int) int {
	return int(math.Round(987007 * math.Pow(float64(kelvin), -0.999)))
}

func elgato_color_value_to_kelvin(value int) int {
	return int(math.Round(1000000 * math.Pow(float64(value), -1)))
}

func parse_status_response(response resty.Response) LightStatus {
	return LightStatus{
		OnOffState:  gjson.Get(response.String(), "lights.0.on").Bool(),
		Brightness:  int(gjson.Get(response.String(), "lights.0.brightness").Int()),
		Temperature: elgato_color_value_to_kelvin(int(gjson.Get(response.String(), "lights.0.temperature").Int())),
	}
}

func parse_info_response(response resty.Response) AccessoryInfo {
	var accessory_info AccessoryInfo

	error := json.Unmarshal(response.Body(), &accessory_info)

	if error != nil {
		fmt.Println("Can not parse accessory info JSON")
		os.Exit(1)
	}

	return accessory_info
}

func print_light_status(status LightStatus) {
	if arguments.Format == "json" {
		print_light_status_json(status)
		return
	}

	if status.OnOffState {
		fmt.Printf("%-17s : %s\n", "State", "on")
	} else {
		fmt.Printf("%-17s : %s\n", "State", "off")
	}
	fmt.Printf("%-17s : %d %%\n", "Brightness", status.Brightness)
	fmt.Printf("%-17s : %d K\n", "Color Temperature", status.Temperature)
}

func print_light_status_json(status LightStatus) {
	output, error := json.Marshal(status)

	if error != nil {
		fmt.Print("Cannot create JSON")
		os.Exit(1)
	}

	fmt.Print(string(output))
}

func print_accessory_info(info AccessoryInfo) {
	if arguments.Format == "json" {
		print_accessory_info_json(info)
		return
	}
	fmt.Printf("%-21s : %s\n", "Product Name", info.ProductName)
	fmt.Printf("%-21s : %d\n", "Hardware Board Type", info.HardwareBoardType)
	fmt.Printf("%-21s : %d\n", "Hardware Revision", info.HardwareRevision)
	fmt.Printf("%-21s : %s\n", "MAC Address", info.MacAddress)
	fmt.Printf("%-21s : %d\n", "Firmware Build Number", info.FirmwareBuildNumber)
	fmt.Printf("%-21s : %s\n", "Firmware Version", info.FirmwareVersion)
	fmt.Printf("%-21s : %s\n", "Serial Number", info.SerialNumber)
	fmt.Printf("%-21s : %s\n", "Display Name", info.DisplayName)
	fmt.Printf("%-21s : %s\n", "Features", strings.Join(info.Features, ", "))
	fmt.Printf("%-21s : %s\n", "Wifi SSID", info.WifiInfo.SSID)
	fmt.Printf("%-21s : %d\n", "Wifi Frequency MHz", info.WifiInfo.FrequencyMHz)
	fmt.Printf("%-21s : %d\n", "Wifi RSSI", info.WifiInfo.RSSI)
}

func print_accessory_info_json(info AccessoryInfo) {
	output, error := json.Marshal(info)

	if error != nil {
		fmt.Print("Cannot create JSON")
		os.Exit(1)
	}

	fmt.Print(string(output))
}
