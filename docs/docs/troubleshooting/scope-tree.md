---
title: Scope tree
description: Print a samber/do injector's scope tree with do.ExplainInjector to visualize scopes, service types, and which services implement Healthchecker or Shutdowner.
sidebar_position: 1
---

# Scope tree

`do.ExplainInjector` prints the full tree of scopes attached to an injector — every child scope, the services registered in each one, their loading type (lazy, eager, transient), and whether they implement `Healthchecker` or `Shutdowner`. Use it to debug an unexpected dependency resolution or to understand an unfamiliar codebase's module layout.

## Spec {#spec}

```go
do.ExplainInjector(do.Injector) do.ExplainInjectorOutput
```

## Print scope tree {#print-scope-tree}

**Play: https://go.dev/play/p/mRs4Q7TJ5jj**

```go
debug := do.ExplainInjector(scope)
println(debug.String())
```

Output:

```txt
Scope ID: 35d18a30-0cb9-4fad-aa3c-1438937612ad
Scope name: [root]

DAG:
 |
  \_ [root] (ID: 35d18a30-0cb9-4fad-aa3c-1438937612ad)
      * 😴 PostgreSQLClientService 🫀 🙅
      * 😴 RedisClientService 🫀 🙅
      * 🔁 Config
      * 🔗 Logger
      |
      |
      |\_ api (ID: dce6f365-fc52-4407-ac02-a6a857e0cbc9)
      |    |
      |    |
      |    |\_ public-api (ID: b9cac0c2-da74-4ee1-a5e0-eb30dca912c3)
      |    |    * 😴 PublicApiRouterService
      |    |    |
      |    |    |
      |    |    |\_ public-api-v1 (ID: a4407628-5b3e-4abf-8d6d-82557d5ddb13)
      |    |    |    * 😴 UserControllerService
      |    |    |
      |    |    |
      |    |     \_ public-api-v2 (ID: adc75f0a-a2dc-488f-9457-510e151d9e34)
      |    |         * 😴 UserControllerService
      |    |         * 😴 ProductControllerService
      |    |         * 😴 CartControllerService
      |    |
      |    |
      |     \_ internal-api (ID: a9e3adfc-ad10-4cbd-9b0d-85cfad95d99c)
      |         * 😴 InternalApiRouterService
      |         * 😴 UserManagementControllerService
      |
      |
       \_ jobs (ID: 53406825-b7cc-46cc-9baf-ab0319c8a3bd)
           * 😴 StatisticsJobService
           * 😴 RateLimitResetJobService
```

```mermaid
graph TD
    A[Root Scope] --> B[PostgreSQLClientService]
    A --> C[RedisClientService]
    A --> D[Config]
    A --> E[Logger]

    A --> F[api Scope]
    A --> G[jobs Scope]

    F --> H[public-api Scope]
    F --> I[internal-api Scope]

    H --> J[public-api-v1 Scope]
    H --> K[public-api-v2 Scope]

    J --> L[UserControllerService]
    K --> M[UserControllerService]
    K --> N[ProductControllerService]
    K --> O[CartControllerService]

    H --> P[PublicApiRouterService]
    I --> Q[InternalApiRouterService]
    I --> R[UserManagementControllerService]

    G --> S[StatisticsJobService]
    G --> T[RateLimitResetJobService]

    style A fill:#e1f5fe
    style F fill:#f3e5f5
    style G fill:#e8f5e8
    style H fill:#fff3e0
    style I fill:#fff3e0
    style J fill:#fce4ec
    style K fill:#fce4ec

    style F fill:#fff3e0
    style G fill:#fff3e0
    style H fill:#fff3e0
    style I fill:#fff3e0
    style J fill:#fff3e0
    style K fill:#fff3e0
```

Emojis describe service type:

- 😴 Lazy service
- 🔁 Eager service
- 🏭 Transient service
- 🔗 Service alias

...and capabilities:

- 🫀 Implements Healthchecker
- 🙅 Implements Shutdowner
