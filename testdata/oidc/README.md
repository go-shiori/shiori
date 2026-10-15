# OIDC test server

Start a local OIDC server to test OIDC authentication.

## Setup

```bash
npm install
```

## Run

```bash
npm run oidc
```

## OIDC Config

```
SHIORI_OIDC_ENABLED=true
SHIORI_OIDC_USERNAME_CLAIM=sub
SHIORI_OIDC_CLIENT_ID=dev-client
SHIORI_OIDC_CLIENT_SECRET=secret
SHIORI_OIDC_ISSUER=http://localhost:4000
SHIORI_OIDC_REDIRECT_URL=http://localhost:8080/api/v1/auth/oidc/callback
```