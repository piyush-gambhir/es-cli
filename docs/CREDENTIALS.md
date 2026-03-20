# Elasticsearch CLI - Authentication & Credentials Guide

This guide covers every authentication scenario supported by the Elasticsearch CLI (`es`). Whether you are running a single-node dev cluster, a production multi-node deployment, Elastic Cloud, or Amazon OpenSearch, this document walks you through obtaining credentials, configuring TLS, and connecting securely.

---

## Table of Contents

- [Quick Start](#quick-start)
- [Getting Your Credentials](#getting-your-credentials)
  - [Basic Auth (Username/Password)](#basic-auth-usernamepassword)
  - [API Keys (Recommended for Automation)](#api-keys-recommended-for-automation)
  - [Bearer Tokens](#bearer-tokens)
- [TLS / SSL Configuration](#tls--ssl-configuration)
- [Configuration](#configuration)
  - [Config File](#config-file)
  - [Environment Variables](#environment-variables)
  - [CLI Flags](#cli-flags)
  - [Multiple Profiles](#multiple-profiles)
  - [Resolution Order](#resolution-order)
- [Deployment Scenarios](#deployment-scenarios)
- [Edge Cases & Troubleshooting](#edge-cases--troubleshooting)
- [Security Best Practices](#security-best-practices)

---

## Quick Start

There are three ways to authenticate. Pick whichever fits your workflow.

### 1. Interactive Login (recommended for first-time setup)

```bash
es login
```

The CLI will prompt you for:

1. **Elasticsearch URL** -- e.g. `https://localhost:9200`
2. **Auth method** -- `basic`, `api-key`, or `bearer`
3. **Credentials** -- username/password, API key ID + secret, or bearer token
4. **CA certificate path** -- leave empty if using a public CA or Elastic Cloud
5. **Skip TLS verification** -- answer `y` only in development
6. **Profile name** -- defaults to `default`

The CLI tests the connection before saving. On success, the profile is written to `~/.config/es-cli/config.yaml` and set as the active profile.

### 2. Environment Variables (recommended for CI/CD)

```bash
# Basic auth
export ES_URL=https://elasticsearch.example.com:9200
export ES_USERNAME=elastic
export ES_PASSWORD=changeme

# Then run any command -- no login needed
es cluster health
```

### 3. CLI Flags (one-off commands)

```bash
es cluster health \
  --url https://elasticsearch.example.com:9200 \
  --username elastic \
  --password changeme
```

Flags override environment variables, which override profile config. See [Resolution Order](#resolution-order) for full details.

---

## Getting Your Credentials

### Basic Auth (Username/Password)

Basic auth sends a `username:password` pair encoded as a Base64 `Authorization: Basic` header on every request. It is the simplest auth method and works with all Elasticsearch versions.

#### Self-Managed Elasticsearch

**Default superuser (`elastic`)**

When you install Elasticsearch 8.x for the first time, it generates a random password for the built-in `elastic` superuser and prints it to stdout during startup:

```
-> Password for the elastic user (reset with `bin/elasticsearch-reset-password -u elastic`):
   Xk8mD2q5R7pYtFnG=
```

If you missed the output or need to reset:

```bash
# Run on the Elasticsearch node
bin/elasticsearch-reset-password -u elastic
```

This prints a new auto-generated password or lets you set one interactively with `-i`:

```bash
bin/elasticsearch-reset-password -u elastic -i
```

**Connect with the CLI:**

```bash
es login
# URL: https://localhost:9200
# Auth method: basic
# Username: elastic
# Password: <paste the password>
```

**Creating additional users (least-privilege)**

Using the superuser for day-to-day operations is discouraged. Create a dedicated user with only the roles it needs.

Via the Elasticsearch API:

```bash
curl -X POST "https://localhost:9200/_security/user/cli-user" \
  -H "Content-Type: application/json" \
  -u elastic:YOUR_ELASTIC_PASSWORD \
  -d '{
    "password": "s3cure-pa55w0rd",
    "roles": ["viewer"],
    "full_name": "CLI Read-Only User"
  }'
```

Via Kibana (step-by-step):

1. Open Kibana in your browser (e.g., `http://localhost:5601`).
2. Click the **hamburger menu** (top-left) and scroll down to **Management**.
3. Click **Stack Management**.
4. In the left sidebar under **Security**, click **Users**.
5. Click the **Create user** button.
6. Fill in the form:
   - **Username:** e.g., `cli-readonly` or `cli-admin`
   - **Full name:** e.g., `CLI Read-Only User`
   - **Email:** optional
   - **Password:** enter a strong password and confirm it
   - **Roles:** click the dropdown and select one or more roles (see table below)
7. Click **Create user**.
8. Use the new username and password with `es login` or environment variables.

Creating a custom role (for fine-grained access):

1. In **Stack Management > Security**, click **Roles** (instead of Users).
2. Click **Create role**.
3. Give it a name (e.g., `cli-logs-reader`).
4. Under **Index privileges**, add the indices this role can access:
   - **Indices:** e.g., `logs-*`
   - **Privileges:** e.g., `read`, `view_index_metadata`
5. Under **Cluster privileges**, add as needed:
   - `monitor` for cluster health/stats
   - `manage_index_templates` for template management
6. Click **Create role**.
7. Then create a user (above) and assign this custom role.

Common built-in roles:

| Role | Access Level |
|------|-------------|
| `superuser` | Full access to everything (avoid in production) |
| `viewer` | Read-only access to all Kibana features and data |
| `editor` | Read-write access to all Kibana features and data |
| `monitoring_user` | Read access to `.monitoring-*` indices |
| `ingest_admin` | Manage ingest pipelines |
| `manage_index_templates` | Manage index and component templates |

You can also create custom roles via the API or Kibana to restrict access to specific indices.

#### Elastic Cloud

When you create a deployment on Elastic Cloud, you are given credentials for the `elastic` superuser.

1. Log in to [cloud.elastic.co](https://cloud.elastic.co).
2. Navigate to your deployment.
3. Under **Security**, find the `elastic` user password. If you need to reset it, click **Reset password**.
4. Copy the **Elasticsearch endpoint URL** from the deployment overview page. It looks like: `https://my-deployment-abc123.es.us-east-1.aws.found.io:443`

```bash
es login
# URL: https://my-deployment-abc123.es.us-east-1.aws.found.io:443
# Auth method: basic
# Username: elastic
# Password: <paste from cloud console>
# CA certificate path: <leave empty -- Elastic Cloud uses public CAs>
# Skip TLS verification: N
```

---

### API Keys (Recommended for Automation)

API keys are the recommended auth method for scripts, CI/CD pipelines, and long-running automation. They can be scoped to specific indices and operations, and they can be revoked independently without changing any user's password.

The CLI uses Elasticsearch's native `ApiKey` authentication scheme. When you provide `--api-key-id` and `--api-key`, the CLI computes `base64(id:api_key)` and sends it as:

```
Authorization: ApiKey <base64-encoded-credentials>
```

#### Creating an API Key via Kibana

1. Open Kibana and go to **Stack Management > Security > API Keys**.
2. Click **Create API key**.
3. Give it a name (e.g., `es-cli-prod`).
4. Optionally restrict its privileges using role descriptors (see [API Key Permissions](#api-key-permissions)).
5. Optionally set an expiration.
6. Click **Create API key**.
7. Kibana shows the **ID** and **API key** values. Copy both immediately -- the API key value is shown only once.

The **ID** maps to `--api-key-id` (or `ES_API_KEY_ID`).
The **API key** maps to `--api-key` (or `ES_API_KEY`).

#### Creating an API Key via API

```bash
curl -X POST "https://localhost:9200/_security/api_key" \
  -H "Content-Type: application/json" \
  -u elastic:YOUR_PASSWORD \
  -d '{
    "name": "es-cli-automation",
    "expiration": "365d",
    "role_descriptors": {
      "cli_role": {
        "cluster": ["monitor", "manage_index_templates"],
        "indices": [
          {
            "names": ["logs-*", "metrics-*"],
            "privileges": ["read", "view_index_metadata"]
          }
        ]
      }
    }
  }'
```

Response:

```json
{
  "id": "VuaCfGcBCdbkQm-e5aOx",
  "name": "es-cli-automation",
  "expiration": 1735689600000,
  "api_key": "ui2lp2axTNmsyakw9tvNnw",
  "encoded": "VnVhQ2ZHY0JDZGJrUW0tZTVhT3g6dWkybHAyYXhUTm1zeWFrdzl0dk5udw=="
}
```

Use the `id` and `api_key` fields from the response:

```bash
es login
# URL: https://localhost:9200
# Auth method: api-key
# API Key ID: VuaCfGcBCdbkQm-e5aOx
# API Key: ui2lp2axTNmsyakw9tvNnw
```

Or with environment variables:

```bash
export ES_URL=https://localhost:9200
export ES_API_KEY_ID=VuaCfGcBCdbkQm-e5aOx
export ES_API_KEY=ui2lp2axTNmsyakw9tvNnw

es cluster health
```

Or with CLI flags:

```bash
es cluster health \
  --url https://localhost:9200 \
  --api-key-id VuaCfGcBCdbkQm-e5aOx \
  --api-key ui2lp2axTNmsyakw9tvNnw
```

> **Note:** The `encoded` field in the API response is `base64(id:api_key)`. The CLI computes this encoding automatically from the ID and API key you provide -- you do not need to provide the pre-encoded value.

#### API Key Permissions

API keys can have their own role descriptors that limit what they can do, independent of the creating user's roles. If no `role_descriptors` are provided, the key inherits the full privileges of the user who created it.

**Read-only key for monitoring:**

```json
{
  "name": "es-cli-monitor",
  "role_descriptors": {
    "monitor": {
      "cluster": ["monitor"],
      "indices": [
        {
          "names": ["*"],
          "privileges": ["read", "view_index_metadata", "monitor"]
        }
      ]
    }
  }
}
```

**Index management key (no document writes):**

```json
{
  "name": "es-cli-index-admin",
  "role_descriptors": {
    "index_admin": {
      "cluster": ["manage_index_templates", "monitor"],
      "indices": [
        {
          "names": ["app-*"],
          "privileges": [
            "create_index", "delete_index", "manage",
            "read", "view_index_metadata"
          ]
        }
      ]
    }
  }
}
```

**Full access key for a specific index pattern:**

```json
{
  "name": "es-cli-logs",
  "role_descriptors": {
    "logs_full": {
      "cluster": ["monitor"],
      "indices": [
        {
          "names": ["logs-*"],
          "privileges": ["all"]
        }
      ]
    }
  }
}
```

**Revoking an API key:**

```bash
curl -X DELETE "https://localhost:9200/_security/api_key" \
  -H "Content-Type: application/json" \
  -u elastic:YOUR_PASSWORD \
  -d '{"id": "VuaCfGcBCdbkQm-e5aOx"}'
```

---

### Bearer Tokens

Bearer tokens authenticate by sending a token in the `Authorization: Bearer <token>` header. Use bearer tokens when:

- You are connecting through Elastic Cloud and using an Elastic Cloud API key (which is a single token, not an id+secret pair).
- You have obtained an OAuth2 access token from Elasticsearch's token-based authentication.
- You are using Elasticsearch service tokens.

#### Elastic Cloud API Keys

Elastic Cloud deployments support API keys created through the Cloud console. These are different from the Elasticsearch-native API keys described above -- they are single opaque tokens.

1. Log in to [cloud.elastic.co](https://cloud.elastic.co).
2. Go to your deployment.
3. Under **Management > Elasticsearch API keys** (or the Security section), create a new key.
4. Copy the token value.

```bash
export ES_URL=https://my-deployment.es.us-east-1.aws.found.io:443
export ES_TOKEN=your-cloud-api-key-token

es cluster health
```

#### OAuth2 / Token-Based Auth

Elasticsearch supports the `_security/oauth2/token` endpoint for obtaining short-lived access tokens:

```bash
curl -X POST "https://localhost:9200/_security/oauth2/token" \
  -H "Content-Type: application/json" \
  -d '{
    "grant_type": "password",
    "username": "elastic",
    "password": "YOUR_PASSWORD"
  }'
```

Response:

```json
{
  "access_token": "dGhpcyBpcyBub3QgYSByZWFsIHRva2VuIGJ1dCBpdCBpcyBvbmx5IHRlc3Q=",
  "type": "Bearer",
  "expires_in": 1200
}
```

Use the `access_token`:

```bash
es cluster health --url https://localhost:9200 --token dGhpcyBpcyBub3QgYSByZWFsIHRva2VuIGJ1dCBpdCBpcyBvbmx5IHRlc3Q=
```

> **Note:** OAuth2 tokens expire (default 20 minutes). They are best for short-lived sessions. For long-running automation, use API keys instead.

#### Service Tokens

Elasticsearch service tokens are used for service-to-service authentication (e.g., Kibana connecting to Elasticsearch). You can generate one with:

```bash
bin/elasticsearch-service-tokens create elastic/kibana my-kibana-token
```

The output token can be used as a bearer token:

```bash
export ES_TOKEN=AAEAAWVsYXN0aWMva2liYW5hL215LWtpYmFuYS10b2tlbjp...
es cluster health --url https://localhost:9200
```

Service tokens are typically reserved for Elastic Stack components, but there is nothing preventing their use with the CLI for testing or debugging purposes.

---

## TLS / SSL Configuration

### Self-Managed with Default Security (Elasticsearch 8.x)

Elasticsearch 8.x enables TLS by default for both HTTP and transport layers. During first startup, it generates self-signed certificates and a Certificate Authority (CA).

**Finding the CA certificate:**

The CA certificate is located at:

```
$ES_HOME/config/certs/http_ca.crt
```

Common paths:

| Installation Method | CA Certificate Path |
|---|---|
| tar.gz / zip | `/path/to/elasticsearch-8.x/config/certs/http_ca.crt` |
| Debian / RPM | `/etc/elasticsearch/certs/http_ca.crt` |
| Docker | `/usr/share/elasticsearch/config/certs/http_ca.crt` |

**Using the CA certificate with the CLI:**

```bash
# Via CLI flag
es cluster health --url https://localhost:9200 --ca-cert /etc/elasticsearch/certs/http_ca.crt

# Via environment variable
export ES_CA_CERT=/etc/elasticsearch/certs/http_ca.crt
es cluster health

# Saved in a profile (via es login)
es login
# CA certificate path: /etc/elasticsearch/certs/http_ca.crt
```

**Skipping TLS verification (development only):**

```bash
# Via CLI flag
es cluster health --url https://localhost:9200 --insecure

# Short flag
es cluster health --url https://localhost:9200 -k

# Via environment variable
export ES_INSECURE=true
```

> **Warning:** Never use `--insecure` in production. It disables certificate verification entirely, making the connection vulnerable to man-in-the-middle attacks.

### Self-Signed Certificates

If your Elasticsearch cluster uses a self-signed certificate or a private CA, you have two options:

**Option 1: Point the CLI to the CA certificate**

```bash
es cluster health --url https://es.internal:9200 --ca-cert /path/to/my-ca.pem
```

The `--ca-cert` flag accepts a PEM-encoded file containing one or more CA certificates. The CLI loads this into a custom TLS root CA pool.

**Option 2: Add the CA to your system trust store**

On macOS:

```bash
sudo security add-trusted-cert -d -r trustRoot \
  -k /Library/Keychains/System.keychain /path/to/my-ca.pem
```

On Ubuntu/Debian:

```bash
sudo cp /path/to/my-ca.pem /usr/local/share/ca-certificates/my-ca.crt
sudo update-ca-certificates
```

On RHEL/CentOS/Fedora:

```bash
sudo cp /path/to/my-ca.pem /etc/pki/ca-trust/source/anchors/my-ca.pem
sudo update-ca-trust
```

Once added to the system trust store, the CLI (and all other programs) will trust certificates issued by that CA without needing `--ca-cert`.

### Elastic Cloud

Elastic Cloud deployments use certificates issued by well-known public Certificate Authorities. No `--ca-cert` is needed, and `--insecure` is not required. TLS works out of the box:

```bash
es cluster health --url https://my-deployment.es.us-east-1.aws.found.io:443
```

### Mutual TLS (mTLS) / Client Certificates

The CLI does not currently support client certificate authentication (mTLS). If your cluster requires client certificates, consider using a reverse proxy that handles the client certificate and forwards requests with basic auth or API key auth to Elasticsearch.

---

## Configuration

### Config File

The CLI stores configuration in YAML at:

```
~/.config/es-cli/config.yaml
```

If the `XDG_CONFIG_HOME` environment variable is set, the path becomes `$XDG_CONFIG_HOME/es-cli/config.yaml`.

The file is created automatically by `es login`. You can also create or edit it manually. The file is written with `0600` permissions (owner read/write only).

**Full config.yaml example with all auth methods:**

```yaml
current_profile: prod

profiles:
  # Basic auth profile
  dev:
    url: https://localhost:9200
    auth_method: basic
    username: elastic
    password: dev-password
    ca_cert: /etc/elasticsearch/certs/http_ca.crt
    insecure: false
    read_only: false

  # API key profile
  staging:
    url: https://staging-es.internal:9200
    auth_method: api_key
    api_key_id: VuaCfGcBCdbkQm-e5aOx
    api_key: ui2lp2axTNmsyakw9tvNnw
    ca_cert: /etc/pki/tls/certs/staging-ca.pem

  # Bearer token profile
  prod:
    url: https://prod-es.example.com:443
    auth_method: bearer
    token: dGhpcyBpcyBub3QgYSByZWFsIHRva2VuIGJ1dCBpdCBpcyBvbmx5IHRlc3Q=
    read_only: true

  # Elastic Cloud profile
  cloud:
    url: https://my-deploy-abc123.es.us-east-1.aws.found.io:443
    auth_method: basic
    username: elastic
    password: cloud-password

  # Minimal profile (insecure dev)
  local:
    url: https://localhost:9200
    auth_method: basic
    username: elastic
    password: changeme
    insecure: true

defaults:
  output: table
```

**Config file fields reference:**

| Field | Type | Description |
|-------|------|-------------|
| `current_profile` | string | Name of the active profile |
| `profiles.<name>.url` | string | Elasticsearch endpoint URL |
| `profiles.<name>.auth_method` | string | `basic`, `api_key`, or `bearer` (auto-detected if omitted) |
| `profiles.<name>.username` | string | Username for basic auth |
| `profiles.<name>.password` | string | Password for basic auth |
| `profiles.<name>.api_key_id` | string | API key ID |
| `profiles.<name>.api_key` | string | API key secret |
| `profiles.<name>.token` | string | Bearer token |
| `profiles.<name>.ca_cert` | string | Path to PEM-encoded CA certificate |
| `profiles.<name>.insecure` | bool | Skip TLS certificate verification |
| `profiles.<name>.read_only` | bool | Block all write operations |
| `defaults.output` | string | Default output format (`table`, `json`, `yaml`) |

### Environment Variables

Environment variables override profile values but are overridden by CLI flags.

| Variable | Description | Example |
|----------|-------------|---------|
| `ES_URL` | Elasticsearch URL | `https://localhost:9200` |
| `ES_USERNAME` | Username for basic auth | `elastic` |
| `ES_PASSWORD` | Password for basic auth | `changeme` |
| `ES_API_KEY_ID` | API key ID | `VuaCfGcBCdbkQm-e5aOx` |
| `ES_API_KEY` | API key secret | `ui2lp2axTNmsyakw9tvNnw` |
| `ES_TOKEN` | Bearer token | `dGhpcyBpcy...` |
| `ES_CA_CERT` | Path to CA certificate for TLS | `/path/to/ca.pem` |
| `ES_INSECURE` | Skip TLS verification (`true` or `1`) | `true` |
| `ES_READ_ONLY` | Block write operations (`true` or `1`) | `true` |
| `ES_NO_INPUT` | Disable interactive prompts | `true` |
| `ES_QUIET` | Suppress informational output | `true` |
| `ES_VERBOSE` | Enable verbose HTTP logging | `true` |

**Auth method auto-detection from environment variables:**

The CLI auto-detects the auth method based on which variables are set:

- If `ES_TOKEN` is set, uses bearer auth.
- Else if both `ES_API_KEY_ID` and `ES_API_KEY` are set, uses API key auth.
- Else if both `ES_USERNAME` and `ES_PASSWORD` are set, uses basic auth.

You do not need to set a separate auth method variable.

### CLI Flags

CLI flags have the highest priority and override everything else.

| Flag | Short | Description |
|------|-------|-------------|
| `--url` | | Elasticsearch URL |
| `--username` | `-u` | Username for basic auth |
| `--password` | `-p` | Password for basic auth |
| `--api-key-id` | | API key ID |
| `--api-key` | | API key secret |
| `--token` | | Bearer token |
| `--ca-cert` | | Path to CA certificate |
| `--insecure` | `-k` | Skip TLS verification |
| `--profile` | | Use a specific profile for this command |
| `--read-only` | | Block write operations |
| `--no-input` | | Disable interactive prompts |
| `--quiet` | `-q` | Suppress informational output |
| `--verbose` | `-v` | Enable verbose HTTP logging |

**Example: override the profile URL for a single command:**

```bash
es cluster health --profile prod --url https://other-node:9200
```

### Multiple Profiles

Profiles let you maintain credentials for multiple clusters and switch between them.

**Creating profiles:**

```bash
# Create a dev profile
es login
# ... fill in dev credentials ...
# Profile name: dev

# Create a staging profile
es login
# ... fill in staging credentials ...
# Profile name: staging

# Create a prod profile
es login
# ... fill in prod credentials ...
# Profile name: prod
```

**Managing profiles:**

```bash
# List all configured profiles
es config list-profiles

# Switch the default profile
es config use-profile staging

# Use a profile for a single command
es cluster health --profile prod

# View current configuration
es config view
```

**Example: multi-environment shell aliases**

Add to your `~/.bashrc` or `~/.zshrc`:

```bash
alias es-dev="es --profile dev"
alias es-staging="es --profile staging"
alias es-prod="es --profile prod --read-only"
```

Then:

```bash
es-dev index list
es-staging cluster health
es-prod index list   # read-only -- write commands are blocked
```

### Resolution Order

Configuration is resolved by layering sources. For each setting, the first non-empty value wins:

```
CLI flags  >  Environment variables  >  Profile config file  >  Defaults
```

Concretely:

1. **CLI flags** -- `--url`, `--username`, `--password`, `--api-key-id`, `--api-key`, `--token`, `--ca-cert`, `--insecure`
2. **Environment variables** -- `ES_URL`, `ES_USERNAME`, `ES_PASSWORD`, `ES_API_KEY_ID`, `ES_API_KEY`, `ES_TOKEN`, `ES_CA_CERT`, `ES_INSECURE`
3. **Profile values** from `~/.config/es-cli/config.yaml` (the active profile, or the one specified by `--profile`)
4. **Defaults** -- output format defaults to `table`

This means you can have a base profile in the config file and override individual values with environment variables or flags on a per-command basis.

---

## Deployment Scenarios

### Self-Managed (Single Node)

The most common development setup. Elasticsearch 8.x enables security by default.

```bash
# After starting Elasticsearch for the first time, note the password and CA cert
es login
# URL: https://localhost:9200
# Auth method: basic
# Username: elastic
# Password: <auto-generated password from startup output>
# CA certificate path: /path/to/elasticsearch-8.x/config/certs/http_ca.crt
# Skip TLS verification: N
# Profile name: local
```

Or via environment variables:

```bash
export ES_URL=https://localhost:9200
export ES_USERNAME=elastic
export ES_PASSWORD=your-password
export ES_CA_CERT=/path/to/elasticsearch-8.x/config/certs/http_ca.crt
```

### Self-Managed (Multi-Node Cluster)

In a multi-node cluster, all nodes share the same CA. Point the CLI at any node or a load balancer in front of the cluster.

```bash
es login
# URL: https://es-lb.internal:9200
# Auth method: api-key
# API Key ID: VuaCfGcBCdbkQm-e5aOx
# API Key: ui2lp2axTNmsyakw9tvNnw
# CA certificate path: /etc/pki/tls/certs/elasticsearch-ca.pem
# Profile name: prod-cluster
```

If your load balancer terminates TLS with a public certificate, you do not need `--ca-cert`.

### Elastic Cloud

Elastic Cloud handles TLS with public CAs and provides the Elasticsearch endpoint URL in the deployment overview.

```bash
es login
# URL: https://my-deployment-abc123.es.us-east-1.aws.found.io:443
# Auth method: basic
# Username: elastic
# Password: <from cloud console>
# CA certificate path: <leave empty>
# Skip TLS verification: N
# Profile name: cloud
```

Alternatively, create a Cloud API key and use bearer auth:

```bash
export ES_URL=https://my-deployment-abc123.es.us-east-1.aws.found.io:443
export ES_TOKEN=your-cloud-api-key

es cluster health
```

### Amazon OpenSearch (Compatibility)

Amazon OpenSearch Service provides an Elasticsearch-compatible API. Use basic auth with the master user credentials.

```bash
es login
# URL: https://search-my-domain-abc123.us-east-1.es.amazonaws.com
# Auth method: basic
# Username: admin
# Password: <your master user password>
# CA certificate path: <leave empty -- AWS uses public CAs>
# Profile name: aws-opensearch
```

> **Note:** OpenSearch has diverged from Elasticsearch. Some advanced Elasticsearch APIs (ILM, certain security endpoints) may not be available or may behave differently on OpenSearch. Basic operations like index management, search, and document CRUD work as expected.

### Docker

**Elasticsearch 8.x Docker (security enabled by default):**

```bash
# Start Elasticsearch
docker run -d --name es \
  -p 9200:9200 \
  -e "discovery.type=single-node" \
  -e "ELASTIC_PASSWORD=changeme" \
  docker.elastic.co/elasticsearch/elasticsearch:8.17.0

# Copy the CA certificate from the container
docker cp es:/usr/share/elasticsearch/config/certs/http_ca.crt /tmp/http_ca.crt

# Connect
export ES_URL=https://localhost:9200
export ES_USERNAME=elastic
export ES_PASSWORD=changeme
export ES_CA_CERT=/tmp/http_ca.crt

es cluster health
```

**Elasticsearch 8.x Docker (security disabled for local dev):**

```bash
docker run -d --name es \
  -p 9200:9200 \
  -e "discovery.type=single-node" \
  -e "xpack.security.enabled=false" \
  docker.elastic.co/elasticsearch/elasticsearch:8.17.0

# No auth needed
export ES_URL=http://localhost:9200
es cluster health
```

### Kubernetes

When running in Kubernetes, credentials are typically stored in Secrets and injected as environment variables.

**Pod spec example:**

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: es-cli-job
spec:
  containers:
    - name: es-cli
      image: your-image-with-es-cli
      env:
        - name: ES_URL
          value: "https://elasticsearch-master.elastic.svc.cluster.local:9200"
        - name: ES_USERNAME
          valueFrom:
            secretKeyRef:
              name: elasticsearch-credentials
              key: username
        - name: ES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: elasticsearch-credentials
              key: password
        - name: ES_CA_CERT
          value: "/etc/elasticsearch/certs/ca.crt"
        - name: ES_NO_INPUT
          value: "true"
      volumeMounts:
        - name: es-certs
          mountPath: /etc/elasticsearch/certs
          readOnly: true
  volumes:
    - name: es-certs
      secret:
        secretName: elasticsearch-ca-cert
```

If you are using the [Elastic Cloud on Kubernetes (ECK)](https://www.elastic.co/guide/en/cloud-on-k8s/current/index.html) operator, the CA certificate is in a Secret named `<cluster-name>-es-http-certs-public`, and the elastic user password is in `<cluster-name>-es-elastic-user`.

---

## Edge Cases & Troubleshooting

### Elasticsearch 7.x vs 8.x

| Aspect | Elasticsearch 7.x | Elasticsearch 8.x |
|--------|-------------------|-------------------|
| Security | Disabled by default (free tier) | Enabled by default |
| TLS (HTTP) | Disabled by default | Enabled by default (self-signed) |
| Default superuser | `elastic` (if security enabled) | `elastic` (always) |
| API key auth | Available (if security enabled) | Available |
| Password | Set by user during setup | Auto-generated on first start |

**Connecting to Elasticsearch 7.x with security disabled:**

```bash
# No auth needed, plain HTTP
export ES_URL=http://localhost:9200
es cluster health
```

**Connecting to Elasticsearch 7.x with security enabled (X-Pack Basic or higher):**

```bash
export ES_URL=https://localhost:9200
export ES_USERNAME=elastic
export ES_PASSWORD=changeme
export ES_CA_CERT=/path/to/ca.pem   # if TLS is enabled
es cluster health
```

### Elasticsearch Behind a Load Balancer

When Elasticsearch is behind a load balancer or reverse proxy:

- Use the load balancer's URL as `--url`.
- If the LB terminates TLS with a public certificate, no `--ca-cert` is needed.
- If the LB terminates TLS with a self-signed certificate, point `--ca-cert` to the LB's CA, not Elasticsearch's CA.
- If the LB passes TLS through to Elasticsearch, point `--ca-cert` to Elasticsearch's CA.

### Elasticsearch Behind a Kibana Proxy

Some setups proxy Elasticsearch through Kibana's server. This is sometimes done to enforce Kibana Spaces or additional access control. If your Elasticsearch is only accessible through a Kibana proxy:

- The URL may include a path prefix (e.g., `https://kibana.example.com/api/console/proxy`).
- This is not a standard setup for CLI access. Connect directly to the Elasticsearch endpoint instead when possible.

### Common Errors

| Error Message | Cause | Fix |
|---------------|-------|-----|
| `elasticsearch URL is required` | No URL provided via flags, env vars, or profile | Set `--url`, `ES_URL`, or run `es login` |
| `connection test failed: ... certificate signed by unknown authority` | TLS is enabled but the CA is not trusted | Provide `--ca-cert /path/to/ca.pem` or add the CA to your system trust store |
| `connection test failed: ... connection refused` | Elasticsearch is not running or wrong port | Check the Elasticsearch process and verify the URL and port |
| `connection test failed: ... x509: certificate is valid for ..., not ...` | URL hostname does not match the certificate's SAN | Use the hostname that matches the certificate, or use `--insecure` in dev |
| `401 Unauthorized` (basic auth) | Wrong username or password | Verify credentials; reset password with `elasticsearch-reset-password` |
| `401 Unauthorized` (API key) | Invalid, expired, or revoked API key | Create a new API key; check expiration |
| `401 Unauthorized` (bearer token) | Expired or invalid token | Obtain a new token; tokens expire (default 20 min for OAuth2) |
| `403 Forbidden` | Authenticated but lacking privileges | Check the user's roles or API key's role descriptors |
| `interactive input required but --no-input is set` | `es login` was run with `--no-input` or `ES_NO_INPUT=true` | Use environment variables instead of `es login`, or remove `--no-input` |
| `creating client: configuring TLS: reading CA certificate ...: no such file or directory` | The CA cert path in your config or `ES_CA_CERT` points to a file that does not exist | Verify the path; update your profile or environment variable |
| `creating client: configuring TLS: failed to parse CA certificate` | The file is not a valid PEM-encoded certificate | Ensure the file is PEM format (starts with `-----BEGIN CERTIFICATE-----`) |
| `command '...' is blocked in read-only mode` | `--read-only` or `ES_READ_ONLY=true` is active | Remove the read-only flag or set `--read-only=false` |

### Debugging Connection Issues

Use `--verbose` (or `-v`) to see the full HTTP request and response, including headers (auth headers are redacted for safety):

```bash
es cluster health --verbose
```

Output:

```
--> GET https://localhost:9200/_cluster/health
    User-Agent: es-cli/1.0.0
    Accept: application/json
    Authorization: [REDACTED]
<-- 200 OK (12ms)
    Content-Type: application/json; charset=UTF-8
```

If you suspect a TLS issue, you can also test with curl:

```bash
curl -v --cacert /path/to/ca.pem -u elastic:password https://localhost:9200/
```

---

## Security Best Practices

### Use API Keys Instead of Passwords

API keys can be scoped to specific indices and operations, rotated independently, and revoked without affecting other users. For any automated or CI/CD use case, prefer API keys over username/password.

### Apply Least-Privilege Principles

Create API keys or users with the minimum permissions needed:

- For monitoring scripts: cluster `monitor` privilege, index `read` + `view_index_metadata`.
- For index management: add `create_index`, `delete_index`, `manage` on specific index patterns.
- For data ingestion: index `create_doc`, `index` on specific indices.
- Avoid `superuser` for day-to-day operations.

### Set Expiration on API Keys

Always set an `expiration` when creating API keys. Even if the key is for long-running automation, a 1-year expiration is better than no expiration:

```json
{
  "name": "es-cli-automation",
  "expiration": "365d"
}
```

### Protect the Config File

The CLI writes `~/.config/es-cli/config.yaml` with `0600` permissions (owner read/write only). Do not loosen these permissions. Passwords and API keys are stored in plaintext in this file.

```bash
# Verify permissions
ls -la ~/.config/es-cli/config.yaml
# Should show: -rw-------
```

If your organization requires encrypted credential storage, use environment variables sourced from a secrets manager (Vault, AWS Secrets Manager, 1Password CLI, etc.) instead of the config file:

```bash
# Example with 1Password CLI
export ES_PASSWORD=$(op item get "Elasticsearch Prod" --fields password)
```

### Use Read-Only Mode as a Safety Net

When exploring a production cluster, enable read-only mode to prevent accidental writes:

```bash
# Per command
es index list --read-only

# Per session
export ES_READ_ONLY=true

# Per profile (in config.yaml)
profiles:
  prod:
    url: https://prod-es:9200
    read_only: true
```

### Do Not Skip TLS Verification in Production

`--insecure` / `ES_INSECURE=true` disables all certificate verification. This is acceptable only for local development with self-signed certificates. In staging and production, always use a proper CA certificate with `--ca-cert`.

### Use `--no-input` in Automation

In CI/CD pipelines and scripts, always set `--no-input` or `ES_NO_INPUT=true` to prevent the CLI from hanging on interactive prompts. Combine with `--confirm` on destructive commands:

```bash
export ES_NO_INPUT=true
es index delete old-index --confirm --if-exists
```

### Rotate Credentials Regularly

- Rotate the `elastic` superuser password periodically.
- Create API keys with expiration dates and replace them before they expire.
- Revoke API keys that are no longer needed.
- After rotating, update your profile with `es login` or edit the config file directly.

### Audit Your Connections

Use verbose mode sparingly during debugging, but note that the CLI redacts `Authorization`, `Cookie`, `Set-Cookie`, and `X-Api-Key` headers in verbose output to prevent accidental credential leakage in logs.
