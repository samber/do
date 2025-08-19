---
title: Web UI
description: Learn how to troubleshoot scopes and services via Web UI
sidebar_position: 4
---

# Web UI

## Without framework

```bash
go get github.com/samber/do/http/std/v2
```

```go
import "github.com/samber/do/http/std/v2"

injector := startProgram()

mux := http.NewServeMux()
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
chihttp.Use(router, "/debug/do", injector)

http.ListenAndServe(":8080", router)
```
