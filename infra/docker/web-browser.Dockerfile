# syntax=docker/dockerfile:1
# Browser-test image for web (docs/foundation/WEB-SHELL.md). Build context: web/.
# The fixture stage builds the same component source with one extra, test-only route; that build
# exists only in this image, never in the production web image.
FROM node:24-slim AS fixture
ENV NEXT_TELEMETRY_DISABLED=1
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci --no-audit --no-fund
COPY . .
RUN mkdir -p app/token-fixtures app/account-fixtures \
 && cp tests/fixtures/token-scopes/page.tsx app/token-fixtures/page.tsx \
 && cp tests/fixtures/account-mcp/page.tsx app/account-fixtures/page.tsx \
 && npm run build

# Playwright's official image; the version must equal @playwright/test in package.json.
FROM mcr.microsoft.com/playwright:v1.63.0-noble@sha256:eff16c30e6f3f4af0a03fa4b706120d5e9b0891c344a27d64559aff5900a4a27 AS browser
ENV NEXT_TELEMETRY_DISABLED=1 CI=1
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci --no-audit --no-fund
COPY . .
COPY --from=fixture /app/.next/standalone /fixture
COPY --from=fixture /app/.next/static /fixture/.next/static
COPY --from=fixture /app/public /fixture/public
# This image's fontconfig resolves the generic sans-serif, system-ui and monospace families to a
# CJK face with no bold weight, so headings render at regular weight. Map them to the image's own
# Liberation faces, a typical system fallback (test environment only; nothing is downloaded).
RUN printf '%s\n' '<?xml version="1.0"?>' '<!DOCTYPE fontconfig SYSTEM "fonts.dtd">' '<fontconfig>' \
      '<alias binding="strong"><family>sans-serif</family><prefer><family>Liberation Sans</family></prefer></alias>' \
      '<alias binding="strong"><family>system-ui</family><prefer><family>Liberation Sans</family></prefer></alias>' \
      '<alias binding="strong"><family>monospace</family><prefer><family>Liberation Mono</family></prefer></alias>' \
      '</fontconfig>' >/etc/fonts/local.conf \
 && fc-cache -f >/dev/null
# The development hydration server writes /app/.next; tests write only under /tmp.
RUN chown -R pwuser:pwuser /app
USER pwuser
CMD ["npx", "playwright", "test"]
