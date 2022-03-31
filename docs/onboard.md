# Onboard sandbox applications

1. Create a Github repo with your jupyter notebook.
2. Create a ResourceGroup item. For e.g 
    ``` json
      {
        "accessPolicy": "OPEN",
          "dataSample": {
          },
          "itemStatus": "ACTIVE",
          "type": [
            "iudx:ResourceGroup",
            "<data model types used in this sandbox>"
          ],
          "provider": "<provider id>",
          "location": {
            "address": "<city name>",
            "type": "Place"
          },
          "@context": "https://voc.iudx.org.in/",
          "tags": [
            "tag1",
            "tag2"
          ],
          "name": "<some name identifying this resource>",
          "description": "<Some description in markdown. Image links are accepted too.",
          "resourceServer": "<resourceServer id>",
          "resourceType": "MESSAGESTREAM",
          "label": "<Label identifying this resource>",
          "dataDescriptor": {
          },
          "repositoryURL": "https://github.com/datakaveri/sandbox-ambulance.git",
          "referenceResources": [
          {
            "id": "<id of a reference resource used in this sandbox>",
            "name": "<some name>",
            "description": "<Some description>",
            "additionalInfoURL": "<Addition info url if needed>"

          }
          ],
          "instance": "sandbox"
      }
    ```
    Note: `repositoryURL` is the url of the github repo where the code is hosted.


3. List all the datasets used by this sandbox as resource items, for e.g 
    ``` json
    {
      "tags": [
        "tag1",
        "tag2"
      ],
      "type": [
        "iudx:Resource",
        "<data model type>"
      ],
      "description": "<Description of this dataset>",
      "@context": "https://voc.iudx.org.in/",
      "label": "<Some label>",
      "provider": "<Provider id>",
      "name": "<unique name of this dataset within this group>",
      "itemStatus": "ACTIVE",
      "resourceGroup": "<sandbox resourceGroup id obtained while onboarding the above>",
      "downloadURL": "<Url to download this dataset>",
      "instance": "sandbox"
    }
    ```

4. Proceed to view the onboarded sandboxes on the UI.


Note: Instructions for onbaording these items can be found [here](https://docs.iudx.org.in).
