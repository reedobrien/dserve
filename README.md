dserve
======

A simple server which may be used for local development, fileserving, or
whatever.

Install
-------

`go get github.com/reedobrien/dserve`


Usage
-----

```
dserve
  -address string
    	The address to listen on (default "127.0.0.1")
  -cert string
    	The TLS certificate to use (default "cert.pem")
  -key string
    	The TLS key to use. (default "key.pem")
  -path string
    	Path to the document root
  -port string
    	The port to listen on (default "8080")
  -tls
    	Use TLS
```

Serve the current directory on localhost port 8080:

`dserve`

Serve the current directory with TLS on localhost port 8443:

`dserve -tls`

You can generate certs suitable with th egenerate_cert.go tool in go's crypto/tls package:

`go run  /usr/local/go/src/crypto/tls/generate_cert.go --host localhost`

Run on all listening addresses:

`dserve -address 0.0.0.0`

Apache logging was originally lifted from [this gist](https://gist.github.com/cespare/3985516) and modifid slightly to add elapsed time.
