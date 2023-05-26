# Onboard a resource

1. Use the json format given below to onboard a dataset. Use <code>POST /api/resource</code

    #### _Note: Ensure that the value of the resourceServer is the same as the value of the id of the dataset to which the resource belongs to_

```json
{
	"id": "datakaveri.org/f7e044eee8122b5c87dce6e7ad64f3266afa41dc/cat-sandbox.iudx.org.in/surat-itms-dp-1/itms-CompressedRawData",
	"createdAt": "2022-10-06T05:27:10.300Z",
	"dataset": "633e672e9072f13ab845be89",
	"description": "Compressed ITMS dataset",
	"downloadURL": "https://iudx-cat-sandbox-dev.s3.ap-south-1.amazonaws.com/dp-itms-1/CompressedRawDataset.csv",
	"icon": "https://iudx-catalogue-assets.s3.ap-south-1.amazonaws.com/instances/default-city.png",
	"instance": "sandbox",
	"itemCreatedAt": "2022-08-11T22:06:37.000Z",
	"itemStatus": "ACTIVE",
	"label": "Realtime bus positions",
	"name": "itms-CompressedRawData",
	"provider": "62387a479072f13ab80be4b9",
	"resourceGroup": "datakaveri.org/f7e044eee8122b5c87dce6e7ad64f3266afa41dc/cat-sandbox.iudx.org.in/surat-itms-dp-1",
	"tags": [
		"itms",
		"mobility",
		"vehicle",
		"transport",
		"buses",
		"gtfs",
		"commute",
		"route",
		"bus tracking"
	],
	"type": ["iudx:Resource", "iudx:TransitManagement"],
	"updatedAt": "2023-04-12T07:00:00.915Z"
}
```
