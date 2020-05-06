# Arrowclient - Arrowhead Client Library (Golang)
##### An Arrowhead Client Library for  Arrowhead Framework 4.0 (lw)

This Arrowhead Client Library covers all communication with the Arrowhead Core Framework 4.0. 

### Usage

#### Import arrowclient 
```go
import (
	influx "github.com/ratzakogl/arrowclient"
)
```

#### Register service "management-service" and allows Intercloud access from "management-cloud"
```go
func registerService(){
arrowhead := arrowclient.Localcloud{
		Address: "172.18.0.6",
		Port:    8440,
		Debug:   true,
	}

	service := arrowclient.Service{
		SystemName: "management-service",
		Address:    "172.18.0.10",
		Port:       42070,
	}

	serviceDescription := arrowclient.ServiceDescription{
		ServiceDefinition: "management-interface",
		Interfaces:        []string{"JSON"},
		ServiceMetadata: map[string]string{
			"type": "management-interface",
		},
	}

	err := arrowhead.RegisterService(serviceDescription, service, "stack", 1, false, 0)
	if err != nil {
		log.Println("Arrowhead: Registering service failed: ", err)
	}
  
  	otherCloud := arrowclient.Cloud{
		Operator:             "Management-Operator",
		CloudName:            "Management-Cloud",
		Address:              "172.19.0.1", 
		Port:                 18446,
		GatekeeperServiceURI: "gatekeeper", 
		Secure:               false,
	}

	err = arrowhead.AuthorizeIntercloud(otherCloud, []arrowclient.ServiceDescription{serviceDescription})
	if err != nil {
		log.Println("Arrowhead: AuthorizeIntercloud failed: ", err)
	}
}
```
#### Request the local service "actuator-service"
```go
func requestLocalService() (string, error) {
	arrowhead := arrowclient.Localcloud{
		Address: "172.18.0.6",
		Port:    8440,
		Debug:   false,
	}

	requester := arrowclient.Service{
		SystemName: "management-service",
		Address:    "172.18.0.10",
		Port:       42070,
	}

	serviceDescription := arrowclient.ServiceDescription{
		ServiceDefinition: "actuator-interface",
		Interfaces:        []string{"JSON"},
		ServiceMetadata: map[string]string{
			"type": "actuator-interface",
		},
	}

	flags := arrowclient.OrchestrationFlags{
		OverrideStore:  true,
		MetadataSearch: true,
	}

	or, err := arrowhead.RequestService(requester, serviceDescription, flags)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("http://%s:%d/%s/", or[0].Provider.Address, or[0].Provider.Port, or[0].ServiceURI), nil
}
```



