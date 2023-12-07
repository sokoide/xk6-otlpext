# xk6-otlpext

## How to build

* Install [prerequisites](https://k6.io/docs/extensions/get-started/create/javascript-extensions/)

```
go install go.k6.io/xk6/cmd/xk6@latest

```

* build
```
xk6 build --with xk6-otlpext=/path/to/xk6-otlpext
```

## How to run

```
SERVICENAME=svc ./k6 run -u 10 /path/to/xk6-otlpext/sample.js --duration 30s -v
```

## References

* [k6 extension](https://k6.io/docs/extensions/get-started/create/javascript-extensions/)
