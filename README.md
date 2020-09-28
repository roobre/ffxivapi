# ![](https://na.finalfantasyxiv.com/favicon.ico) FFXIVAPI

FFXIVAPI is a service which exposes a REST API on top of the information provided by the [Lodestone](https://eu.finalfantasyxiv.com/lodestone), by parsing the HTML output and converting it into JSON models.

API documentation is made available through the API itself.

## Features

Currently supported endpoints are:

* [x] `/character/search`: Search for characters given their name and world
* [x] `/character/{id}`: Retrieve character data, including achievements

```json
{
  "ParsedAt": "2020-09-28T14:21:56.403712609Z",
  "ID": 31688528,
  "World": "Ragnarok (Chaos)",
  "Avatar": "https://img2.finalfantasyxiv.com/f/7eb4d62ddd701b2fc5cc06fc773187e9_40d57ba713628f3f1ef5ef204b6d76d2fc0_96x96.jpg?1601301983",
  "Portrait": "https://img2.finalfantasyxiv.com/f/7eb4d62ddd701b2fc5cc06fc773187e9_40d57ba713628f3f1ef5ef204b6d76d2fl0_640x873.jpg?1601301983",
  "Name": "Roobre Shiram",
  "Nameday": "22nd Sun of the 5th Astral Moon",
  "City": "Gridania",
  "GC": {
    "Name": "Order of the Twin Adder",
    "Rank": "Chief Serpent Sergeant"
  },
  "FC": {
    "ID": "9237023573225362244",
    "Name": "Chupipandi"
  },
  "Achievements": [
    {
      "ID": 1158,
      "Name": "Freebird: Dravanian Forelands",
      "Obtained": "2020-09-27T22:16:53Z"
    },
    {
      "ID": 1209,
      "Name": "Mapping the Realm: Sohm Al",
      "Obtained": "2020-09-27T22:16:53Z"
    }
    // ...
  ],
  "ClassJobs": [
    {
      "Name": "Ninja",
      "Level": 62,
      "Exp": 0,
      "ExpNext": 0
    }
  ]
}
```

If you'd like to see more endpoints, feel free to drop a PR or a feature request issue.
