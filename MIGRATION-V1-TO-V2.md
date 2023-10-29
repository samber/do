
## 🔥 Migration from v1 to v2

### 0- Rename package imports

```sh
go get github.com/samber/do@v2
```

```sh
find . -type f -exec sed -i 's#samber/do"#samber/do/v2"#g' {} \;
```

```sh
go mod tidy
```

### 1- `do.Injector` interface

`do.Injector` has been transformed into an interface. Replace `*do.Injector` by `do.Injector`.

```sh
find . -type f -exec sed -i "s/*do.Injector/do.Injector/g" {} \;
```
