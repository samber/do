
# do - Dependency Injection

[![tag](https://img.shields.io/github/tag/samber/do.svg)](https://github.com/samber/do/releases)
![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-%23007d9c)
[![GoDoc](https://godoc.org/github.com/samber/do?status.svg)](https://pkg.go.dev/github.com/samber/do)
![Build Status](https://github.com/samber/do/actions/workflows/test.yml/badge.svg)
[![Go report](https://goreportcard.com/badge/github.com/samber/do)](https://goreportcard.com/report/github.com/samber/do)
[![Coverage](https://img.shields.io/codecov/c/github/samber/do)](https://codecov.io/gh/samber/do)
[![License](https://img.shields.io/github/license/samber/do)](./LICENSE)

**⚙️ A dependency injection toolkit based on Go 1.18+ Generics.**

This library implements the Dependency Injection design pattern. It may replace the fantastic `uber/dig` package. `samber/do` uses Go 1.18+ generics and therefore offers a type‑safe API.

![image](https://github.com/user-attachments/assets/81b91fa7-cdb4-4094-94ba-a0179abc6bf7)

**See also:**

- [samber/ro](https://github.com/samber/ro): Reactive Programming for Go: declarative and composable API for event-driven applications
- [samber/lo](https://github.com/samber/lo): A Lodash-style Go library based on Go 1.18+ Generics
- [samber/mo](https://github.com/samber/mo): Monads based on Go 1.18+ Generics (Option, Result, Either...)
- [samber/cc-skills-golang](https://github.com/samber/cc-skills-golang): AI Agent Skills for Golang

----

<div align="center">
  <sup><b>💖 Sponsored by:</b></sup>
  <br>
  <a href="https://www.dbos.dev/?utm_campaign=gh-smbr">
    <div>
	  <img width="200" alt="dbos" src="https://github.com/user-attachments/assets/d583cb62-7735-4d3c-beb7-e6ef1a5faf49" />
    </div>
    <div>
      DBOS - Durable workflow orchestration library for Go
    </div>
  </a>
</div>

----

**Why this name?**

I love the **short name** for such a utility library. This name is the sum of `DI` and `Go` and no Go package uses this name.

## 💡 Features

- **📒 Service registration**
  - Register by type
  - Register by name
  - Register multiple services from a package at once
- **🪃 Service invocation**
  - Eager loading
  - Lazy loading
  - Transient loading
  - Tag-based invocation
  - Circular dependency detection
- **🧙‍♂️ Service aliasing**
  - Implicit (provide struct, invoke interface)
  - Explicit (provide struct, bind interface, invoke interface)
- **🔁 Service lifecycle**
  - Health check
  - Graceful unload (shutdown)
  - Dependency-aware parallel shutdown
  - Lifecycle hooks
- **📦 Scope (a.k.a module) tree**
  - Visibility control
  - Dependency grouping
- **📤 Container**
  - Dependency graph resolution and visualization
  - Default container
  - Container cloning
  - Service override
- **🧪 Debugging & introspection**
  - Explain APIs: scope tree and service dependencies
  - Web UI & HTTP middleware (std, Gin, Fiber, Echo, Chi)
- **🌈 Lightweight, no dependencies**
- **🔅 No code generation**
- **😷 Type‑safe API**

## 🚀 Install

```sh
# v2 (latest)
go get github.com/samber/do/v2@latest

# v1
go get github.com/samber/do@v1.6.0

# AI Agent Skill
npx skills add https://github.com/samber/cc-skills-golang --skill golang-samber-do
```

This library is v2 and follows SemVer strictly.

No breaking changes will be made to exported APIs before v3.0.0.

This library has no dependencies except the Go std lib.

### 🔥 Migration from v1 to v2

[Documentation here](https://do.samber.dev/docs/upgrading/from-v1-x-to-v2)

## 🤠 Documentation

- [GoDoc: https://godoc.org/github.com/samber/do/v2](https://godoc.org/github.com/samber/do/v2)
- [Documentation](https://do.samber.dev/docs/getting-started)
- [Examples](https://github.com/samber/do/tree/master/examples)
- [Project templates](https://do.samber.dev/examples)

## 🎬 Project boilerplate

- [do-template-api](https://github.com/samber/do-template-api)
- [do-template-worker](https://github.com/samber/do-template-worker)
- [do-template-cli](https://github.com/samber/do-template-cli)

## 🤝 Contributing

- Ping me on Twitter [@samuelberthe](https://twitter.com/samuelberthe) (DMs, mentions, whatever :))
- Fork the [project](https://github.com/samber/do)
- Fix [open issues](https://github.com/samber/do/issues) or request new features

Don't hesitate ;)

```bash
# Install some dev dependencies
make tools

# Run tests
make test
# or
make watch-test
```

## 👤 Contributors

![Contributors](https://contrib.rocks/image?repo=samber/do)

## 💫 Show your support

Give a ⭐️ if this project helped you!

[![GitHub Sponsors](https://img.shields.io/github/sponsors/samber?style=for-the-badge)](https://github.com/sponsors/samber)

## 📝 License

Copyright © 2022 [Samuel Berthe](https://github.com/samber).

This project is [MIT](./LICENSE) licensed.
