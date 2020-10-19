# PLG

```
$ ./plg --help
Usage of ./plg:
  -config string
    	configuration file (default "config.json")
  -type string
    	record|serve
```

The configuration file:

```
{
    "exporters": [
        {
            "url": "http://192.168.1.172:42000/metrics?collect%5B%5D=custom_query.hr&collect%5B%5D=global_status&collect%5B%5D=info_schema.innodb_metrics&collect%5B%5D=standard.go&collect%5B%5D=standard.process",
            "username": "pmm", 
            "password": "/agent_id/964cac3c-d3f7-4d93-839f-3cb4a2a22f7d",
            "duration": 5,
            "name": "mysql_hr_5s"
        },
        ...
```

| Option Name  | Description  |
|---|---|
| URL  | Addres of the exporter. For record the whole URL will be used. For serving, the only the `path` part will be used.  |
|  Username | Username for recording  |
|  Password |  Password for recording  |
|  Duration | Scrape interval |
|  Name     | Where to store the output |

```
...
    "time": 20,
    "bind": "localhost:8082"
}
```

| Option Name  | Description  |
|---|---|
| Time  | How long exporters should be scrapped  |
| Bind  | Where to listen (serve action only)  |