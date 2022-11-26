# HTTP(S) Proxy Server

## CLI arguments

```shell
--host host             Server host
--enable-http           Enable HTTP server (default: false)
--port-http port        HTTP port (default: 80)
--enable-https          Enable HTTPS server (default: false)
--port-https port       HTTPS port (default: 443)
--crt-file file         Location of the SSL certificate file
--key-file file         Location of the RSA private key file
--enable-auth           Enable authentication (default: false)
--username value        Username
--password value        Password
--timeout-read value    Maximum duration for reading the entire request, including the body (default: 0)
--timeout-write value   Maximum duration before timing out writes of the response (default: 0)
--timeout-dial value    Dial timeout
--enable-log            Enable file logging (default: false)
--log-dir value         Location of the log directory (default: "log")
--log-connections       Log HTTP(S) connections (default: true)
--config file, -c file  Location of the configuration file in .yml format
--quiet, -q             Activate quiet mode (default: false)
--help, -h              Print this help text and exit
--version, -v           Print program version and exit
```

## Configuration file

| Property | Type | Default | Description |
|:---|:---|:---|:---|
| `host` | `string` | | Server host |
| `enable-http` | `bool` | `false` | Enable HTTP server |
| `port-http` | `int` | `80` | HTTP port |
| `enable-https` | `bool` | `false` | Enable HTTPS server |
| `port-https` | `int` | `443` | HTTPS port |
| `crt-file` | `string` | | Location of the SSL certificate file |
| `key-file` | `string` | | Location of the RSA private key file |
| `enable-auth` | `bool` | `false` | Enable authentication |
| `username` | `string` | | Username |
| `password` | `string` | | Password |
| `timeout-read` | `int` | `0` | Maximum duration for reading the entire request, including the body |
| `timeout-write` | `int` | `0` | Maximum duration before timing out writes of the response |
| `timeout-dial` | `int` | `10` | Dial timeout |
| `enable-log ` | `bool` | `false` | Enable file logging |
| `log-dir` | `string` | `log` | Location of the log directory |
| `log-connections` | `bool` | `true` | Log HTTP(S) connections |
| `quiet` | `bool` | `false` | Activate quiet mode |

## Issues

Submit the [issues](https://github.com/attilabuti/https-proxy/issues) if you find any bug or have any suggestion.

## Contribution

Fork the [repo](https://github.com/attilabuti/https-proxy) and submit pull requests.

## License

This project is licensed under the [MIT License](https://github.com/attilabuti/https-proxy/blob/main/LICENSE).