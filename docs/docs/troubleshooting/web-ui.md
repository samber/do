---
title: Web UI
description: Learn how to troubleshoot scopes and services via Web UI
sidebar_position: 4
---

# Web UI

> Caution
>
> Do not expose the debug Web UI publicly in production. It reveals internal
> information about your DI graph (service names, dependencies, etc.). Protect
> the routes with authentication (for example, Basic Auth) and/or network
> restrictions (IP allowlist, VPN). Apply your auth middleware to the router
> group before mounting the debug handlers.

## Without framework

```bash
go get github.com/samber/do/http/std/v2
```

```go
import "github.com/samber/do/http/std/v2"

injector := startProgram()

mux := http.NewServeMux()
// Protect with your own middleware (e.g., Basic Auth) before mounting
// the debug handler in production.
mux.Handle("/debug/do/", std.Use("/debug/do", injector))

http.ListenAndServe(":8080", mux)
```

## Gin

```bash
go get github.com/samber/do/http/gin/v2
```

```go
import "github.com/samber/do/http/gin/v2"

injector := startProgram()

router := gin.New()
// Attach auth middleware to the group to protect debug UI in production.
ginhttp.Use(router.Group("/debug/do"), injector)

router.Run(":8080")
```

## Fiber

```bash
go get github.com/samber/do/http/fiber/v2
```

```go
import "github.com/samber/do/http/fiber/v2"

injector := startProgram()

router := fiber.New()
// Attach auth middleware to the group to protect debug UI in production.
fiberhttp.Use(router.Group("/debug/do"), "/debug/do", injector)

router.Listen(":8080")
```

## Echo

```bash
go get github.com/samber/do/http/echo/v2
```

```go
import "github.com/samber/do/http/echo/v2"

injector := startProgram()

router := echo.New()
// Attach auth middleware to the group to protect debug UI in production.
echohttp.Use(router.Group("/debug/do"), "/debug/do", injector)

router.Start(":8080")
```

## Chi

```bash
go get github.com/samber/do/http/chi/v2
```

```go
import "github.com/samber/do/http/chi/v2"

injector := startProgram()

router := chi.NewRouter()
// Protect with your own middleware (e.g., Basic Auth) before mounting
// the debug handler in production.
chihttp.Use(router, "/debug/do", injector)

http.ListenAndServe(":8080", router)
```
