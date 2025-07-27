# Go Track 

Program to help you track the TV Shows you watch. 

## Setup 

First, grab a TVDB API token from [TVDB](https://www.themoviedb.org/). 

Next, take a copy of the `docker-compose.example.yml` and add your token. 

## Cloudflare Authentication (Optional)

If you want to support multiple users then you need to use Cloudflare Zero Trust for authentication. 

Once that's setup, uncomment `DISABLE_AUTH`. 

## Development 

```shell 
tailwindcss.exe -o assets/tailwind.css -w -m
go run main.go
```

## Tech Stack 

* Golang 
    * html/template
* SQLite DB 
* Tailwind CSS
    * heroicons
* TVDB API 
