## News Parser
This application allows you to parse and aggregate news from multiple websites according to passed patterns.

### Getting started
- `make run`

### Endpoints
- Add website to be parsed: ```POST: localhost:8083/websites``` with ```application/json``` body:
```
{
  "main_url": "https://habr.com/ru/",
  "url_pattern": "https://habr.com/ru/post/[0-9]+/$",
  "title_pattern": "span.post__title-text",
  "text_pattern": "div.post__body"
}
```
or
```
{
  "main_url": "https://rbc.ru",
  "url_pattern": "https://www.rbc.ru/society/[0-9]{2}/[0-9]{2}/[0-9]{4}/[a-f0-9]{24}",
  "title_pattern": "h1.article__header__title-in ",
  "text_pattern": "div.article__text__overview"
}
```
where ```main_url``` is URL of main page of the website to be parsed, ```url_pattern``` is regexp for URL 
of news page on the website, ```title_pattern``` and ```text_pattern``` are HTML paths to elements containing
 title and text of news.

- ```GET: localhost:8083/news``` with optional query parameters ```?limit=10``` and ```?offset=0``` 
(10 and 0 are default values)

- ```GET: localhost:8083/news/search?q=example``` where ```q``` is substring to be searched by

Website is being checked every ~5 minutes.

### TODO:
- Add logging and better error handling
- Add unit tests
- Profile ```/news``` endpoint and make optimizations