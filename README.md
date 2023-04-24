## egressproxy

### Overview

egressproxy is an HTTP and HTTPS proxy. It generates TLS certificates on demand when a client requests an HTTPS site, and signs certificates with a private CA key. Any client that is configured to trust the public CA key will trust the dynamically generated TLS certificate. egressproxy uses regex lists of hosts and URLs to allow access to sites: it blocks acccess by design unless a request matches an allowed host or URL.

### Running

#### CA Certificate

You will need to generate or provide a CA key pair (`CA_KEY` and `CA_CERT`) as PEM encoded environment variables at runtime. This key is used to sign certificates for client requests.

#### Dynamically Generated Certificates
The dynamically generated certificates need the following elements for an X.509 distinguished name, which need to be provided as environment variables at runtime:

```
CA_COUNTRY
CA_PROVINCE
CA_LOCALITY
CA_O
CA_OU
```

These can be set to anything that you like, since the `subject` field of the TLS certificate, and the signature from the private `CA_KEY` are all that thet client care about.

#### Ports

By default, egressproxy listens on `:38000` for HTTP and HTTPS requests. Internally, egress proxy also uses `127.0.0.1:38443` to handle TLS connections. This additional loopback socket is used to handle HTTP CONNECT requests. Both settings are configurable via the `-httpListenAddr` and `-tlsListenAddr` flags respectively.

#### Access Control Lists

egressproxy supports two lists of ACLs: `ALLOWED_HOSTS` and `ALLOWED_URLS`. Both lists are configured as environment variables. They use regular expressions and are checked in sequential order. The first matching entry in a list will allow access.

The `ALLOWED_HOSTS` list is checked first. Entries in both lists are separated by a newline.

For example, to allow access to GitHub hosts, set the following:

```
export ALLOWED_HOSTS="^github\.com$"
^gist\.github\.com$
^raw\.githubusercontent\.com$"
<other hosts here>
<separated by newlines>"
```

To allow URLs, add them to an `ALLOWED_URLS` environment variable. For example the following URLS would allow access to Python Flask packages hosted on pypi:

```
export ALLOWED_URLS="^https:\/\/pypi\.org\/simple\/flask\/$
^https:\/\/files\.pythonhosted\.org\/packages\/.+\/Flask.+(tar\.gz|whl)$
<other urls here>
<separated by newlines>
"

### Using

#### HTTP

For HTTP connections, set an `http_proxy` environment variable to point at the running instance. For example `export http_proxy="http://127.0.0.1:38000"`. Some clients use upper-case `HTTP_PROXY` so be sure to set both variables.

#### HTTPS

For HTTPS connections, set an `https_proxy` environment variable to point at the running instance. Be sure the URL uses `http://` since this will ensure the client uses an [HTTP CONNECT request to tunnel the connection](https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods/CONNECT). For example, `export https_proxy="http://127.0.0.1:38000"`. Some clients use upper-case `HTTPS_PROXY` so be sure to set both variables.

### Example Request/Response Headers

Here is an example HTTPS request to `github.com` via egressproxy:

```
http_proxy=http://127.0.0.1:38000 https_proxy=http://127.0.0.1:38000 curl -sIkv https://github.com
* Uses proxy env variable https_proxy == 'http://127.0.0.1:38000'
*   Trying 127.0.0.1:38000...
* Connected to 127.0.0.1 (127.0.0.1) port 38000 (#0)
* allocate connect buffer
* Establish HTTP proxy tunnel to github.com:443
> CONNECT github.com:443 HTTP/1.1
> Host: github.com:443
> User-Agent: curl/7.88.1
> Proxy-Connection: Keep-Alive
>
< HTTP/1.1 200 OK
HTTP/1.1 200 OK
< Date: Mon, 24 Apr 2023 17:18:33 GMT
Date: Mon, 24 Apr 2023 17:18:33 GMT
< Transfer-Encoding: chunked
Transfer-Encoding: chunked
* Ignoring Transfer-Encoding in CONNECT 200 response
<

* CONNECT phase completed
* CONNECT tunnel established, response 200
* ALPN: offers h2,http/1.1
* TLSv1.3 (OUT), TLS handshake, Client hello (1):
* TLSv1.3 (IN), TLS handshake, Server hello (2):
* TLSv1.3 (IN), TLS handshake, Encrypted Extensions (8):
* TLSv1.3 (IN), TLS handshake, Certificate (11):
* TLSv1.3 (IN), TLS handshake, CERT verify (15):
* TLSv1.3 (IN), TLS handshake, Finished (20):
* TLSv1.3 (OUT), TLS change cipher, Change cipher spec (1):
* TLSv1.3 (OUT), TLS handshake, Finished (20):
* SSL connection using TLSv1.3 / TLS_AES_128_GCM_SHA256
* ALPN: server did not agree on a protocol. Uses default.
* Server certificate:
*  subject: C=CA; ST=ON; L=Toronto; O=GitHub; OU=jamonation; CN=github.com
*  start date: Apr 24 17:18:33 2023 GMT
*  expire date: Apr 24 18:18:33 2023 GMT
*  issuer: C=CA; ST=ON; L=Toronto; O=GitHub; OU=jamonation; CN=jamon.ca
*  SSL certificate verify result: unable to get local issuer certificate (20), continuing anyway.
* using HTTP/1.x
> HEAD / HTTP/1.1
> Host: github.com
> User-Agent: curl/7.88.1
> Accept: */*
>
* TLSv1.3 (IN), TLS handshake, Newsession Ticket (4):
< HTTP/1.1 200 OK
HTTP/1.1 200 OK
. . .
```

Note the `subject` field in the certificate is set to the configured CA_COUNTRY, CA_ORGANIZATION etc. values, and the `issuer` is set to the details in the configured CA_CERT public certificate.
