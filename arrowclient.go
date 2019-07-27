package arrowclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

type Localcloud struct {
	Address string
	Port    int
	Debug   bool
}

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

type OrchestrationFlags struct {
	OverrideStore          bool `json:"overrideStore,omitempty"`
	ExternalServiceRequest bool `json:"externalServiceRequest,omitempty"`
	EnableInterCloud       bool `json:"enableInterCloud,omitempty"`
	Matchmaking            bool `json:"matchmaking,omitempty"`
	MetadataSearch         bool `json:"metadataSearch,omitempty"`
	TriggerInterCloud      bool `json:"triggerInterCloud,omitempty"`
	PingProviders          bool `json:"pingProviders,omitempty"`
	//OnlyPreferred          bool `json:"onlyPreferred,omitempty"`
	//EnableQoS              bool `json:"enableQoS,omitempty"` //not implemented in 4.0
}

type OrchestrationForm struct {
	Service     ServiceDescription `json:"serviceDescription"`
	Provider    Service            `json:"provider"`
	ServiceURI  string             `json:"serviceURI"`
	Instruction string             `json:"instruction"`
	Warnings    []string           `json:"warnings"`
}

type orchestrationResponse struct {
	OrchestrationForm []OrchestrationForm `json:"response"`
}

func (l Localcloud) InitializeDatabase() error {
	service := "serviceregistry"
	subpath := "mgmt/all"
	url := "http://" + l.Address + ":" + strconv.Itoa(l.Port) + "/" + service + "/" + subpath

	client := &http.Client{}

	request, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (l Localcloud) RegisterService(description ServiceDescription, service Service, serviceUri string, version int, udp bool, ttl int) error {
	p := ServiceRegistryEntry{
		ProvidedService: description,
		Provider:        service,
		ServiceURI:      serviceUri,
		Version:         version,
		Udp:             udp,
		Ttl:             ttl,
	}

	bytearray, err := json.MarshalIndent(p, "", "\t")
	if err != nil {
		return err
	}

	_, err = l.arrowheadPOST("serviceregistry", "register", bytearray)

	return err
}

func (l Localcloud) RemoveService(description ServiceDescription, service Service, serviceUri string, version int, udp bool, ttl int) error {
	p := ServiceRegistryEntry{
		ProvidedService: description,
		Provider:        service,
		ServiceURI:      serviceUri,
		Version:         version,
		Udp:             udp,
		Ttl:             ttl,
	}

	bytearray, err := json.MarshalIndent(p, "", "\t")
	if err != nil {
		return err
	}

	_, err = l.arrowheadPUT("serviceregistry", "remove", bytearray)
	return err
}

func (l Localcloud) RequestService(requester Service, requestedService ServiceDescription, flags OrchestrationFlags) ([]OrchestrationForm, error) {
	var orchestrationResponse orchestrationResponse

	p := struct {
		RequesterSystem    Service            `json:"requesterSystem"`
		RequestedService   ServiceDescription `json:"requestedService"`
		OrchestrationFlags OrchestrationFlags `json:"orchestrationFlags"`
		//RequestedQoS       map[string]string  `json:"requestedQoS"`
		//PreferredProviders []Service          `json:"preferredProviders"`
	}{
		RequesterSystem:    requester,
		RequestedService:   requestedService,
		OrchestrationFlags: flags,
	}

	bytearray, _ := json.MarshalIndent(p, "", "\t")

	resp, err := l.arrowheadPOST("orchestrator", "orchestration", bytearray)
	if err != nil {
		return orchestrationResponse.OrchestrationForm, fmt.Errorf("RequestService failed: %s\n", err)
	}

	err = json.Unmarshal(resp, &orchestrationResponse)

	return orchestrationResponse.OrchestrationForm, err
}

func (l Localcloud) Subscribe(eventName string, consumer Service, providers []Service, notifyUri string, matchMetadata bool) error {
	p := eventFilter{
		EventType:     eventName,
		Consumer:      consumer,
		Sources:       providers,
		NotifyUri:     notifyUri,
		MatchMetadata: matchMetadata,
	}

	bytearray, err := json.MarshalIndent(p, "", "\t")
	if err != nil {
		return err
	}

	_, err = l.arrowheadPOST("eventhandler", "subscription", bytearray)
	return err
}

func (l Localcloud) Unsubscribe(eventName string, consumer Service, providers []Service, notifyUri string, matchMetadata bool) error {
	p := eventFilter{
		EventType:     eventName,
		Consumer:      consumer,
		Sources:       providers,
		NotifyUri:     notifyUri,
		MatchMetadata: matchMetadata,
	}

	bytearray, err := json.MarshalIndent(p, "", "\t")
	if err != nil {
		return err
	}

	_, err = l.arrowheadPUT("eventhandler", "subscription", bytearray)
	return err
}

func (l Localcloud) Publish(eventName string, payload string, source Service, callbackUri string) error {
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

	bytearray, err := json.MarshalIndent(p, "", "\t")
	if err != nil {
		return err
	}

	_, err = l.arrowheadPOST("eventhandler", "publish", bytearray)
	return err
}

func (l Localcloud) AuthorizeIntercloud(otherCloud Cloud, serviceList []ServiceDescription) error {
	p := struct {
		Cloud       Cloud                `json:"cloud"`
		ServiceList []ServiceDescription `json:"serviceList"`
	}{
		otherCloud,
		serviceList,
	}

	bytearray, err := json.MarshalIndent(p, "", "\t")
	if err != nil {
		return err
	}

	_, err = l.arrowheadPOST("authorization", "mgmt/intercloud", bytearray)
	return err
}

func (l Localcloud) AuthorizeIntracloud(consumer Service, providers []Service, service ServiceDescription) error {
	p := struct {
		Consumer  Service            `json:"consumer"`
		Providers []Service          `json:"providers"`
		Service   ServiceDescription `json:service`
	}{
		Consumer:  consumer,
		Providers: providers,
		Service:   service,
	}

	bytearray, err := json.MarshalIndent(p, "", "\t")
	if err != nil {
		return err
	}

	_, err = l.arrowheadPOST("authorization", "mgmt/intracloud", bytearray)
	return err
}

func (l Localcloud) arrowheadPOST(service string, subpath string, payload []byte) ([]byte, error) {
	url := "http://" + l.Address + ":" + strconv.Itoa(l.Port) + "/" + service + "/" + subpath

	if l.Debug {
		fmt.Println(url)
		fmt.Println(string(payload))
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(payload))

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if l.Debug {
		fmt.Println("************************PUT******************")
		fmt.Println(string(body))
		fmt.Println("*********************************************")
	}

	return body, nil
}

func (l Localcloud) arrowheadPUT(service string, subpath string, payload []byte) ([]byte, error) {
	url := "http://" + l.Address + ":" + strconv.Itoa(l.Port) + "/" + service + "/" + subpath

	if l.Debug {
		fmt.Println(url)
		fmt.Println(string(payload))
	}

	client := &http.Client{}

	request, err := http.NewRequest("PUT", url, bytes.NewReader(payload))
	request.Header.Set("Content-Type", "application/json")
	request.ContentLength = int64(len(payload))

	resp, err := client.Do(request)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if l.Debug {
		fmt.Println("*********************************************")
		fmt.Println(string(body))
		fmt.Println("*********************************************")
	}

	return body, nil
}
