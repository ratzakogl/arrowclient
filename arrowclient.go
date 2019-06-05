package arrowclient

//TODO: Refactoring: arrowheadLocalCloud.RegisterService() ? (contains address and port for request)
//TODO: Errorhandling
//TODO: Comments

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

//Identifies a Service (An Application) on a System (Computer/Microcontroller) e.g. RPI3-Garden2
//(IP-)Address, Port: Where is the API of the Device located
type Service struct {
	SystemName         string `json:"systemName"`
	Address            string `json:"address"`
	Port               int    `json:"port"`
	AuthenticationInfo string `json:"authenticationInfo,omitempty"`
	Secure             bool   `json:"secure,omitempty"`
}

//Identifies the Services a System has to offer.
//ServiceDefinition: Name of the Service. E.g. Thermometer service
//Interfaces: List of Interfaces that this Service (on this Device) offers. E.g. REST, SOAP, CoAP
//ServiceMetadata: Map/Dictionary Datenstruktur: Key-Value Pair das den Service klar Definiert. z.B. typ=thermometer, unit=celsius, sensor=tmp235
type ServiceDescription struct {
	ServiceDefinition string            `json:"serviceDefinition"`
	Interfaces        []string          `json:"interfaces,omitempty"`
	ServiceMetadata   map[string]string `json:"serviceMetadata"`
}

type ServiceRegistryEntry struct {
	ProvidedService ServiceDescription `json:"providedService"`
	Provider        Service            `json:"provider"`
	ServiceURI      string             `json:"serviceUri"`
	Version         int                `json:"version"`
	Udp             bool               `json:"udp"`
	Ttl             int                `json:"ttl,omitempty"`
}

type eventFilter struct {
	EventType     string    `json:"eventType"`
	Consumer      Service   `json:"consumer"`
	Sources       []Service `json:"sources"`
	NotifyUri     string    `json:"notifyUri"`
	MatchMetadata bool      `json:"matchMetadata"`
}

type event struct {
	Name      string `json:"type"`
	Payload   string `json:"payload"`
	Timestamp string `json:"timestamp"` //muss von java.time.LocalDateTime geparst werden können (<60" alt )
}

type Cloud struct {
	Operator             string `json:"operator"`
	CloudName            string `json:"cloudName"`
	Address              string `json:"address"`
	Port                 int    `json:"port"`
	GatekeeperServiceURI string `json:"gatekeeperServiceUri"`
	AuthenticationInfo   string `json:"authenticationInfo,omitempty"`
	Secure               bool   `json:secure,omitempty`
}

func RegisterService(description ServiceDescription, service Service, serviceUri string, version int, udp bool, ttl int) {
	p := ServiceRegistryEntry{
		ProvidedService: description,
		Provider:        service,
		ServiceURI:      serviceUri,
		Version:         version,
		Udp:             udp,
		Ttl:             ttl,
	}

	bytearray, _ := json.MarshalIndent(p, "", "\t")

	arrowheadPOST("serviceregistry", "register", bytearray)
}

func RemoveService(description ServiceDescription, service Service, serviceUri string, version int, udp bool, ttl int) {
	p := ServiceRegistryEntry{
		ProvidedService: description,
		Provider:        service,
		ServiceURI:      serviceUri,
		Version:         version,
		Udp:             udp,
		Ttl:             ttl,
	}

	bytearray, _ := json.MarshalIndent(p, "", "\t")

	arrowheadPUT("serviceregistry", "remove", bytearray)
}

func Subscribe(eventName string, consumer Service, providers []Service, notifyUri string, matchMetadata bool) {
	p := eventFilter{
		EventType:     eventName,
		Consumer:      consumer,
		Sources:       providers,
		NotifyUri:     notifyUri,
		MatchMetadata: matchMetadata,
	}

	bytearray, _ := json.MarshalIndent(p, "", "\t")

	arrowheadPOST("eventhandler", "subscription", bytearray)
}

func Unsubscribe(eventName string, consumer Service, providers []Service, notifyUri string, matchMetadata bool) {
	p := eventFilter{
		EventType:     eventName,
		Consumer:      consumer,
		Sources:       providers,
		NotifyUri:     notifyUri,
		MatchMetadata: matchMetadata,
	}

	bytearray, _ := json.MarshalIndent(p, "", "\t")

	arrowheadPUT("eventhandler", "subscription", bytearray)
}

func Publish(eventName string, payload string, source Service, callbackUri string) {
	p := struct {
		Source              Service `json:"source"`
		Event               event   `json:"event"`
		DeliveryCompleteUri string  `json:"deliveryCompleteUri"`
	}{
		Source: source,
		Event: event{
			Name:      eventName,
			Payload:   payload,
			Timestamp: time.Now().Format("2006-01-02T15:04:05"),
		},
		DeliveryCompleteUri: callbackUri,
	}

	bytearray, _ := json.MarshalIndent(p, "", "\t")

	arrowheadPOST("eventhandler", "publish", bytearray)
}

func AuthorizeIntercloud(otherCloud Cloud, serviceList []ServiceDescription) {
	p := struct {
		Cloud       Cloud                `json:"cloud"`
		ServiceList []ServiceDescription `json:"serviceList"`
	}{
		otherCloud,
		serviceList,
	}

	bytearray, _ := json.MarshalIndent(p, "", "\t")

	arrowheadPOST("authorization", "mgmt/intercloud", bytearray)
}

func AuthorizeIntracloud(consumer Service, providers []Service, service ServiceDescription) {
	p := struct {
		Consumer  Service            `json:"consumer"`
		Providers []Service          `json:"providers"`
		Service   ServiceDescription `json:service`
	}{
		Consumer:  consumer,
		Providers: providers,
		Service:   service,
	}

	bytearray, _ := json.MarshalIndent(p, "", "\t")

	arrowheadPOST("authorization", "mgmt/intracloud", bytearray)
}

func arrowheadPOST(service string, subpath string, payload []byte) error {
	url := "http://172.18.0.3:8440/" + service + "/" + subpath

	fmt.Println(url)
	fmt.Println(string(payload))

	resp, err := http.Post(url, "application/json", bytes.NewReader(payload))

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println("*********************************************")
	fmt.Println(string(body))
	fmt.Println("*********************************************")

	//return err
	return nil
}

func arrowheadPUT(service string, subpath string, payload []byte) error {
	url := "http://172.18.0.3:8440/" + service + "/" + subpath

	fmt.Println(url)
	fmt.Println(string(payload))

	client := &http.Client{}

	request, err := http.NewRequest("PUT", url, bytes.NewReader(payload))
	request.Header.Set("Content-Type", "application/json")
	request.ContentLength = int64(len(payload))

	resp, err := client.Do(request)

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println("*********************************************")
	fmt.Println(string(body))
	fmt.Println("*********************************************")

	//return err
	return nil
}
