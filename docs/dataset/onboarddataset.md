# Onboard a dataset

1. Use the json format given below to onboard a dataset. Use <code>POST /api/dataset</code>

```json
{
	"id": "datakaveri.org/f7e044eee8122b5c87dce6e7ad64f3266afa41dc/cat-sandbox.iudx.org.in/bangalore-foi",
	"accessPolicy": "OPEN",
	"createdAt": "2023-03-27T22:00:00.977Z",
	"description": "In this notebook, we illustrate the use of GSX and the OGC Features api to access Feature Layers in Bangalore such as Water bodies, buildings, etc. In addition, we illustrate the use of an open source library to handle this GeoJSON based feature data. \n\n <br/>  <img src=\"https://raw.githubusercontent.com/datakaveri/sandbox-bangalore-foi/master/docs/foi1.jpg\" alt=\"FOI\" height=\"400\" width=\"600\"/>",
	"domain": "",
	"icon": "https://iudx-catalogue-assets.s3.ap-south-1.amazonaws.com/instances/default-city.png",
	"instance": "sandbox",
	"itemCreatedAt": "2021-07-05T06:35:26.000Z",
	"itemStatus": "ACTIVE",
	"iudxResourceAPIs": ["ATTR", "TEMPORAL"],
	"label": "Features of Interest in Bangalore",
	"location": {
		"address": "Bangalore",
		"type": "Place"
	},
	"name": "bangalore-foi",
	"provider": {
		"_id": "62387a479072f13ab80be4b9",
		"description": "Sandbox Admin"
	},
	"repositoryURL": "https://github.com/datakaveri/sandbox-bangalore-foi.git",
	"resourceServer": "datakaveri.org/f7e044eee8122b5c87dce6e7ad64f3266afa41dc/cat-sandbox.iudx.org.in",
	"resourceType": "MESSAGESTREAM",
	"resources": 0,
	"schema": "https://voc.iudx.org.in/Features",
	"tags": ["features", "lakes", "rivers", "buildings"],
	"type": ["iudx:ResourceGroup", "iudx:Features"],
	"unique_id": "datakaveri.org-f7e044eee8122b5c87dce6e7ad64f3266afa41dc-cat-sandbox.iudx.org.in-bangalore-foi",
	"updatedAt": "2023-04-12T04:00:01.124Z",
	"views": 6
}
```
