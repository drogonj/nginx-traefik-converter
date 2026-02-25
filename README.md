# Ingress Traefik Converter

[![Go Report Card](https://goreportcard.com/badge/github.com/nikhilsbhat/nginx-traefik-converter)](https://goreportcard.com/report/github.com/nikhilsbhat/nginx-traefik-converter)
[![shields](https://img.shields.io/badge/license-MIT-blue)](https://github.com/nikhilsbhat/nginx-traefik-converter/blob/master/LICENSE)
[![shields](https://godoc.org/github.com/nikhilsbhat/nginx-traefik-converter?status.svg)](https://godoc.org/github.com/nikhilsbhat/nginx-traefik-converter)
[![shields](https://img.shields.io/github/v/tag/nikhilsbhat/nginx-traefik-converter.svg)](https://github.com/nikhilsbhat/nginx-traefik-converter/tags)
[![shields](https://img.shields.io/github/downloads/nikhilsbhat/nginx-traefik-converter/total.svg)](https://github.com/nikhilsbhat/nginx-traefik-converter/releases)

This CLI helps migrate Kubernetes Ingress resources from the NGINX Ingress Controller to Traefik v3 in a safe, explicit, and production-oriented way.

## Overview

Rather than attempting a naive 1:1 translation of annotations, the tool understands the semantic differences between NGINX and Traefik and generates Traefik-native CRDs (IngressRoute, Middleware, TLSOption) only when a correct and meaningful mapping exists. 

Annotations that cannot be safely expressed in Traefik are detected, warned about, and intentionally skipped to avoid silent behavior changes.

The goal of the CLI is to make migrations predictable and reviewable, not to hide complexity. CLI always know what was converted, what was not, and why.

---

## Features

### Core capabilities

- **NGINX Ingress → Traefik v3 migration**
    - Converts Kubernetes `Ingress` resources into Traefik `IngressRoute` objects when required
    - Generates Traefik `Middleware` and `TLSOption` resources as needed

- **CRD-native output**
    - Produces Traefik v3–compatible YAML
    - Uses `IngressRoute`, `Middleware`, and `TLSOption` CRDs
    - Avoids dynamic or runtime configuration hacks

- **Safe annotation conversion**
    - Converts only annotations with a well-defined Traefik equivalent
    - Explicitly warns and skips unsupported or non-equivalent annotations

---

### Supported annotation handling

- **HTTP behavior**
    - Path rewrites
    - HTTP → HTTPS redirects
    - CORS configuration
    - Rate limiting
    - Request and response header manipulation

- **Backend protocol handling**
    - `nginx.ingress.kubernetes.io/backend-protocol`
    - `nginx.ingress.kubernetes.io/grpc-backend`
    - Correct promotion from `Ingress` to `IngressRoute`
    - Supports HTTP, HTTPS, gRPC (h2c), and gRPCS backends

- **TLS and mTLS**
    - Converts `auth-tls-verify-client` to Traefik `TLSOption`
    - Correct TLS-layer handling (not middleware)
    - Clear warnings for CA certificate and static configuration requirements

- **cert-manager Certificate extraction** *(in progress)*
    - Extracts live `Certificate` resources from the cluster via the Kubernetes dynamic client
    - Falls back to generating a `Certificate` from `cert-manager.io/*` Ingress annotations when no live resource is found
    - Sanitizes extracted resources for GitOps (removes `ownerReferences`, `resourceVersion`, `managedFields`, `status`, etc.)
    - Prevents certificate loss when the original Ingress is deleted during migration

- **Configuration snippets**
    - Converts **header-only** `configuration-snippet` directives
    - Detects and warns on unsafe or NGINX-specific directives
    - Never injects raw configuration into Traefik

---

### Safety and transparency

- **No silent behavior changes**
    - Unsupported features are never auto-converted
    - All skipped annotations are reported via warnings

- **Heuristic conversions are opt-in**
    - Non-equivalent mappings (e.g. buffering heuristics) require explicit flags
    - Heuristics are clearly labeled and never applied by default

- **Ingress-scoped behavior**
    - Annotations apply only to the Ingresses that define them
    - No accidental global Traefik configuration

---

### Designed for real-world migrations

- Handles large clusters using paginated Kubernetes API access
- Produces deterministic output suitable for code review
- Intended as a **migration aid**, not a black-box replacement
- Aligns with Traefik v3 CRDs and best practices

---

## 📌 How this differs from Traefik’s native NGINX migration

Traefik provides an official guide for migrating from NGINX Ingress to Traefik here:

- https://doc.traefik.io/traefik/migrate/nginx-to-traefik/

That guide focuses on **native Ingress compatibility** and covers only the subset of NGINX annotations that Traefik can interpret directly.

**Ingress Traefik Converter is not an alternative to that migration path.**

Instead, this tool goes further and focuses on **CRD-native, production-safe migrations**:

- It analyzes NGINX annotations that Traefik does **not** support natively
- It generates Traefik CRDs (`IngressRoute`, `Middleware`, `TLSOption`) where a safe and explicit mapping exists
- It detects, warns about, and skips annotations that cannot be represented correctly
- It never relies on Traefik’s Ingress compatibility layer
- It never injects raw or unsafe configuration into Traefik

In particular, this tool helps with annotations listed as **unsupported by Traefik’s Ingress-NGINX compatibility layer**, documented here:

- https://doc.traefik.io/traefik-hub/api-gateway/reference/routing/kubernetes/ref-ingress-nginx#unsupported-nginx-annotations

Where possible, the converter:
- Maps these annotations to **Traefik-native CRDs**
- Or emits **explicit warnings** when no safe equivalent exists

### When should I use this tool?

- Use Traefik’s native migration when your setup fits its supported subset of annotations
- Use **Ingress Traefik Converter** when you:
  - Rely on complex or unsupported NGINX annotations
  - Want CRD-native Traefik configuration instead of Ingress compatibility mode
  - Need explicit, reviewable, production-safe migration output

### What this tool will not do

- It will not attempt lossy or unsafe conversions
- It will not silently change behavior
- It will not pretend NGINX-specific features exist in Traefik

The goal is **correctness, transparency, and safety**, not “make it work at any cost”.

---
## Installation

* Recommend installing released versions. Release binaries are available on the [releases](https://github.com/nikhilsbhat/nginx-traefik-converter/releases) page.

#### Homebrew

Install latest version on `nginx-traefik-converter` on `macOS`

```shell
brew tap nikshilsbhat/stable git@github.com:nikhilsbhat/homebrew-stable.git
# for latest version
brew install nikshilsbhat/stable/nginx-traefik-converter
# for specific version
brew install nikshilsbhat/stable/nginx-traefik-converter@0.0.3
```

Check [repo](https://github.com/nikhilsbhat/homebrew-stable) for all available versions of the formula.

#### Docker

Latest version of docker images are published to [ghcr.io](https://github.com/nikhilsbhat/nginx-traefik-converter/pkgs/container/nginx-traefik-converter), all available images can be found there. </br>

```bash
docker pull ghcr.io/nikhilsbhat/nginx-traefik-converter:latest
docker pull ghcr.io/nikhilsbhat/nginx-traefik-converter:<github-release-tag>
```

#### Build from Source

1. Clone the repository:
    ```sh
    git clone https://github.com/nikhilsbhat/nginx-traefik-converter.git
    cd nginx-traefik-converter
    ```
2. Build the project:
    ```sh
    make local.build
    ```

---

## Usage

### Basic Usage

To convert the existing Nginx ingress to Traefik configurations:

```sh
nginx-traefik-converter convert -a                                   #should convert ingress present in all the namespace.
nginx-traefik-converter convert -c kube-context-one                  #when you have multiple contexts in same kubeconfig file.
nginx-traefik-converter convert -c kube-context-one -n namespace-one #adding to above, operations limited to namespace 'namespace-one'  
```

## Documentation

Updated documentation on all available commands and flags can be
found [here](https://github.com/nikhilsbhat/nginx-traefik-converter/blob/main/docs/doc/nginx-traefik-converter.md).