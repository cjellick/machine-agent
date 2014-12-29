package events

type Event struct {
	Name                         string
	Id                           string
	PreviousIds                  string
	previousNames                string
	publisher                    string
	replyTo                      string
	resourceId                   string
	resourceType                 string
	time                         string
	timeoutMillis                string
	transitioning                string
	transitioningInternalMessage string
	transitioningMessage         string
	transitioningProgress        string
	Data                         map[string]string
}

/*
{
    "context": {
        "logicName": "demo",
        "logicPath": "physicalhost.activate->(demo)",
        "prettyProcess": "physicalhost.activate",
        "prettyResource": "physicalHost:1",
        "processId": "30",
        "processName": "physicalhost.activate",
        "processUuid": "2649d4f9-5695-4f38-b49b-3e0b257ff325",
        "resouceId": "1",
        "resouceType": "physicalHost",
        "topProcessName": "physicalhost.activate",
        "topResourceId": "1",
        "topResourceType": "physicalHost"
    },
    "data": {
        "driver": "virtualbox",
        "kind": "dockerMachine",
        "name": "test-random-280937",
        "virtualboxMemory": "2048"
		"virtualboxDiskSize":
        "virtualboxBoot2dockerUrl":
        "digitaloceanImage":
        "digitaloceanRegion":
        "digitaloceanSize":
        "digitaloceanAccessToken":
    },
    "id": "190ad7e5-fa1d-4e28-97a2-b9b1bad3f6a8",
    "name": "physicalhost.activate;handler=demo",
    "previousIds": null,
    "previousNames": null,
    "publisher": null,
    "replyTo": "reply.7884953948567153747",
    "resourceId": "1ph1",
    "resourceType": "physicalHost",
    "time": 1419876894816,
    "timeoutMillis": 15000,
    "transitioning": null,
    "transitioningInternalMessage": null,
    "transitioningMessage": null,
    "transitioningProgress": null
}

*/
