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

`ShutdownOnSignals` used to return only 1 argument.

```go
# from
err := injector.ShutdownOnSignals(syscall.SIGTERM, os.Interrupt)

# to
signal, err := injector.ShutdownOnSignals(syscall.SIGTERM, os.Interrupt)
```

`injector.ShutdownOnSIGTERM()` has been removed. Use `injector.ShutdownOnSignals(syscall.SIGTERM)` instead.

## 4- Internal service naming

Internally, the DI container stores a service by its name (string) that represents its type. In `do@v1`, some developers reported collisions in service names, because the package name was not included.

Eg: `*mypkg.MyService` -> `*github.com/samber/example.MyService`.

In case you invoke a service by its name (highly discouraged), you should make some changes.

To scan a project at the speed light, just run:

```bash
grep -nrE 'InvokeNamed|OverrideNamed|HealthCheckNamed|ShutdownNamed' .
```
