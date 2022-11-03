# Go Web Template

This is base template for doing spikes for smaller applications. It uses the following:

- Golang Templates for UI
- sqlite3 for datastore (can use Postgres)
- Generic oauth2 login (tested with Google)

## Quick Start

Copy the `.env.template` file and change any values you like. You can see the possible values by looking at `internals/models/config.go`.

Once the settings are in place you can do:

```
make start
```

and then browse to `http://localhost:3000`.

```
.
├── cmd
│   └── server               <= Main entry point
├── docker-compose.yaml
├── Dockerfile
├── internals                <= Your App code
│   ├── handlers             <= HTTP handlers
│   ├── models               <= App structs
│   └── repository           <= Database queries
├── migrations               <= SQL migrations
│   └── 000000-init.sql
├── static                   <= Images, css, etc
│   └── plant-research.png
├── templates                <= HTML Pages
│   └── home.html
└── upload-temp              <= For uploaded files

```

## OAuth2 Setup (Google)

[Setup on Google Console](https://console.cloud.google.com/apis/dashboard)

## sqlite3 vs Postgres

The code can support both sqlite3 or postgres. By default it uses sqlite3, but if you look at `start_db` in the Makefile and the example values in `.env.template` you can see how to get Postgres working.
