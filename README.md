[![License](https://img.shields.io/github/license/MunifTanjim/stremthru?style=for-the-badge)](https://github.com/MunifTanjim/stremthru/blob/master/LICENSE)

# StremThru

## Features

- [Byte Serving](https://en.wikipedia.org/wiki/Byte_serving)
- HTTP Proxy
- Proxy Authorization

## Configuration

Configuration is done using environment variables.

**`STREMTHRU_HTTP_PROXY`**

HTTP Proxy URL.

**`STREMTHRU_HTTPS_PROXY`**

HTTPS Proxy URL.

**`STREMTHRU_PROXY_AUTH_CREDENTIALS`**

Comma separated list of credentials for proxy authorization, supports:

- plain text, e.g. `username:password`
- or base64 encoded string, e.g. `dXNlcm5hbWU6cGFzc3dvcmQ=`

## Endpoints

### `/proxy`

**Methods**: `HEAD` and `GET`

**Headers**

- `Proxy-Authorization`: Basic auth header, e.g. `Basic dXNlcm5hbWU6cGFzc3dvcmQ=`

  `Proxy-Authorization` is required when `STREMTHRU_PROXY_AUTH_CREDENTIALS` is set and `token` query param is missing.

**Query Params**

- `url` _(required)_: URL to proxy, should be url encoded
- `token`: auth credential

  `token` is required when `STREMTHRU_PROXY_AUTH_CREDENTIALS` is set and `Proxy-Authorization` header is missing.

## Usage

**Source**

```sh
git clone https://github.com/MunifTanjim/stremthru
cd stremthru

# configure
export STREMTHRU_PROXY_AUTH_CREDENTIALS=username:password

# run
make run

# build and run
make build
./stremthru
```

**Docker**

```sh
docker run --name stremthru -p 8080:8080 \
    -e STREMTHRU_PROXY_AUTH_CREDENTIALS=username:password \
    muniftanjim/stremthru
```

**Docker Compose**

```yml
stremthru:
  image: muniftanjim/stremthru
  ports:
    - 8080:8080
  env_file:
    - .env
  restart: unless-stopped
```

## Related Resources

Cloudflare WARP:

- [github.com/cmj2002/warp-docker](https://github.com/cmj2002/warp-docker)

## License

Licensed under the MIT License. Check the [LICENSE](./LICENSE) file for details.
