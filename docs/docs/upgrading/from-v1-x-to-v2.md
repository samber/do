---
title: From v1.x to v2
description: Upgrade samber/do from v1.x to v2
sidebar_position: 1
---

# Upgrade samber/do from v1.x to v2

This documentation will help you upgrade your application from `samber/do` v1 to `samber/do` v2.

`samber/do` v2 is a new major version, including breaking changes requiring you to adjust your applications accordingly. We will guide to during this process and also mention a few optional recommendations.

This release is a large rewrite, but the breaking changes are relatively easy to handle. Some updates can be done with a simple `sed` command.

Check the release notes [here](https://github.com/samber/do/releases).

No breaking change will be done until v3.

## 1- Upgrading package

Update go.mod:

```sh
go get -u github.com/samber/do/v2
```

Replace package import:

```sh
find . -name '*.go' -type f -exec sed -i '' 's#samber/do"$#samber/do/v2"#g' {} \;
```

Cleanup previous dependencies:

```sh
go mod tidy
```

## 2- `do.Injector` interface

`do.Injector` has been transformed into an interface. Replace `*do.Injector` by `do.Injector`.

```sh
find . -name '*.go' -type f -exec sed -i '' "s/*do.Injector/do.Injector/g" {} \;
```

## 3- `do.Shutdown****` output

Shutdown functions used to return only 1 argument.

```go
# from
err := injector.Shutdown()

# to
signal, err := injector.Shutdown()
```
