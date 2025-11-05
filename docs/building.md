# Building Breakglass

Instructions for building the breakglass application from source.

## Docker Image Build

### OSS Flavour (Recommended for Non-Telekom)

```bash
docker build -t breakglass:oss .
```

This builds the standard OSS-neutral UI suitable for all organizations.

### ⚠️ Telekom Branded Flavour

```bash
docker build --build-arg UI_FLAVOUR=telekom -t breakglass:telekom .
```

## IMPORTANT: TELEKOM BRANDED UI DISCLAIMER

The Telekom branded UI flavour is **PROPRIETARY TO DEUTSCHE TELEKOM** and is intended exclusively for use within Deutsche Telekom entities.

**Terms of Use:**

- The Telekom UI flavour contains proprietary Deutsche Telekom branding, designs, and customizations
- Unauthorized use outside of Deutsche Telekom is strictly prohibited
- Non-compliance with this restriction violates Deutsche Telekom's intellectual property rights
- Organizations not affiliated with Deutsche Telekom must use the OSS flavour

**Which version should I use?**

- **Deutsche Telekom employees and authorized entities:** May use the Telekom flavour
- **All other organizations:** MUST use the OSS flavour

The OSS flavour is fully functional, feature-complete, and appropriate for all use cases.

## Go Binary Build

Build the backend binary:

```bash
make build
```

Output: `./bin/breakglass`

## Frontend Build

Build the frontend separately:

```bash
cd frontend
npm install
npm run build
```

Output: `./frontend/dist/`

## Building for Production

1. Use OSS flavour (unless you are Deutsche Telekom)

```bash
docker build -t breakglass:prod -f Dockerfile .
```

2. Configure for production:

```bash
# Use external OIDC provider
# Enable TLS
# Configure proper email service
# Set resource limits
```

3. Push to your registry:

```bash
docker tag breakglass:prod myregistry.example.com/breakglass:v1.0.0
docker push myregistry.example.com/breakglass:v1.0.0
```

## Development Build

Build with dev configuration:

```bash
make deploy_dev
```

This includes:

- Keycloak for OIDC testing
- Mailhog for email testing
- NodePort services for local access

## Build Flags

Available build arguments for `docker build`:

- `UI_FLAVOUR` - `oss` (default) or `telekom`

Example:

```bash
docker build --build-arg UI_FLAVOUR=telekom -t breakglass:telekom .
```

## Dockerfile Details

The Dockerfile:

1. Builds Go backend
2. Builds frontend (with selected UI flavour)
3. Packages into final image
4. Sets up health checks

## Cross-Platform Builds

Build for multiple architectures:

```bash
docker buildx build --platform linux/amd64,linux/arm64 \
  -t breakglass:latest .
```

## Compliance

When building for any environment:

- Ensure you comply with licensing terms (REUSE standard)
- Non-Telekom organizations: use OSS flavour only
- Keep dependencies up to date
- Scan for vulnerabilities

## License Compliance

This project uses REUSE standard licensing. Each file contains license information.

View licenses:

```bash
ls ./LICENSES/
cat REUSE.toml
```

## Size Optimization

To reduce image size:

```bash
# Use alpine base (already minimal)
# Disable debug symbols in production
# Use multi-stage builds (automatic in Dockerfile)
```

Current image size: ~150MB (approximate, varies with flavour)
