# ![☄️](https://na.finalfantasyxiv.com/favicon.ico) FFXIVAPI [![Build Status](https://travis-ci.org/roobre/ffxivapi.svg?branch=master)](https://travis-ci.org/roobre/ffxivapi) [![Maintainability](https://api.codeclimate.com/v1/badges/167cc30b58f5a04acca7/maintainability)](https://codeclimate.com/github/roobre/ffxivapi/maintainability)

FFXIVAPI is a simple, fast and feature-incomplete REST API for FFXIV service which works on top of the information provided by the [Lodestone](https://eu.finalfantasyxiv.com/lodestone), by parsing the HTML output and converting it into JSON models.

FFXIVAPI is aggresive and heavily parallelizes requests to the Lodestone when possible, to minimize the already big latency the aforementioned service has. Typically, it can retrieve character data and the full list of achievements in about 5 seconds:

```
curl localhost:8080/character/31688528\?achievements\=yes  0.03s user 0.00s system 0% cpu 4.998 total
```

API documentation is available as a [Swagger spec](https://github.com/roobre/ffxivapi/blob/master/http/swagger.yaml) and is made available through the API itself in the root (`/`) path (assuming the binary is executed in the root source directory.

## Features

Currently supported endpoints are:

* [x] `/character/search`: Search for characters given their name and world
* [x] `/character/{id}`: Retrieve character data, including achievements
* [x] `/character/{id}/avatar`: Hotlink character avatar given its ID

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
    },
    "// Truncated for readability",
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
