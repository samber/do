
## 🔥 Migration from v1 to v2

### 1- Rename package imports

```sh
go get github.com/samber/do/v2
```

```sh
find . -type f -exec sed -i 's#samber/do"#samber/do/v2"#g' {} \;
```

```sh
go mod tidy
```

### 2- `do.Injector` interface

`do.Injector` has been transformed into an interface. Replace `*do.Injector` by `do.Injector`.

```sh
find . -type f -exec sed -i "s/*do.Injector/do.Injector/g" {} \;
```

### 3- `do.Shutdown****` output

Shutdown functions used to return only 1 argument.

```go
# from
err := injector.Shutdown()

# to
signal, err := injector.Shutdown()
```
