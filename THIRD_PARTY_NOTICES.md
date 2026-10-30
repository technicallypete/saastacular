# Third-party and separately licensed material

SaaStacular is Copyright 2026 Vitruvian Technology, Corp. Material owned by
Vitruvian Technology, Corp. is licensed under the Apache License, Version 2.0
(see `LICENSE` and `NOTICE`). That grant is subject to the licenses that already
apply to particular files and does not replace their notices. Third-party
material is not relicensed; its own license and notices apply. Where this file
and a license file in the repository differ, the license file governs.

This file uses three categories:

1. **Bundled artifacts.** Files shipped in this repository.
2. **Downloaded build and runtime dependencies.** Not in this repository; fetched
   by package managers, image pulls or source builds when you build or run.
3. **Optional components.** Downloaded only for an optional feature, platform or
   workflow.

Licence statements below come only from files in this repository. Where the
repository holds no licence file or licence metadata for a component, the entry
says **not established offline**. That component's own upstream licence applies,
and you must review it before redistributing or using it in production. Notes
marked *adopter review* describe obligations you may have when you build,
distribute or operate the software. They are not waivers, and this file does not
claim that any such obligation has been met.

## 1. Bundled artifacts (shipped files)

### shadcn/ui-derived modules (MIT)

These files come from the shadcn/ui registry (Base UI, Rhea style) and keep the
shadcn/ui MIT licence (Copyright (c) 2023 shadcn):

- the 54 modules under `web/src/components/ui/`;
- `web/src/hooks/use-mobile.ts`;
- `web/src/lib/utils.ts`;
- `web/src/styles/rhea-leftovers.css`.

The full licence text is in
[`web/THIRD_PARTY_NOTICES.md`](web/THIRD_PARTY_NOTICES.md). That file also records
the npm packages these modules use (section 2).

### UI kit source under `web/src` (first-party)

The 69 files under `web/src` are the company's UI kit source. They ship unaltered,
and their integrity is checked against `web/tests/reference/kit-source.sha256`.
The kit is first-party material of Vitruvian Technology, Corp., licensed under the
root Apache License 2.0, except that the shadcn/ui-derived modules listed above
keep their MIT notice. The kit is built on shadcn/ui and Base UI. Its eight
composition modules under `web/src/components/` (outside `ui/`) and the other
stylesheets under `web/src/styles/` are company-authored.

**No branded image, logo or mark file is shipped.** The kit's internal component
names remain in file paths, import paths, code comments and CSS names, along with
the attribution above. One kit header component that the application does not use
keeps a default wordmark text value. That default prop does not mean the
application displays a brand mark: the application header shows the configured
product name or the tenant's own label. None of these internal names grants
trademark rights. Apache License 2.0, section 6, grants no permission to use trade
names, trademarks, service marks or product names.

### Bundled fonts (SIL Open Font License 1.1)

The font files are unmodified upstream files. Each family's licence ships next to
it, byte-exact.

| Family | Files | Licence file | Copyright (from the licence file) |
|---|---|---|---|
| Geist | `web/public/fonts/geist/geist-wght.ttf` | `web/public/fonts/geist/OFL.txt` | Copyright 2024 The Geist Project Authors |
| Geist Mono | `web/public/fonts/geist-mono/geist-mono-wght.ttf` | `web/public/fonts/geist-mono/OFL.txt` | Copyright 2024 The Geist Project Authors |
| Barlow Condensed | `web/public/fonts/barlow-condensed/barlow-condensed-400.ttf`, `barlow-condensed-600.ttf` | `web/public/fonts/barlow-condensed/OFL.txt` | Copyright 2017 The Barlow Project Authors |
| Oswald | `web/public/fonts/oswald/oswald-wght.ttf` | `web/public/fonts/oswald/OFL.txt` | Copyright 2016 The Oswald Project Authors |

These fonts are licensed under the SIL Open Font License, Version 1.1 only. They
are not covered by the Apache License 2.0.

### Imagery

`web/public/blueprint/overview.svg` is original, project-authored geometry
(first-party, Apache-2.0). The test-only images under `web/tests/fixtures/` are
synthetic, project-authored inputs (first-party, Apache-2.0).

### Company material under its own MIT licence

These directories keep their own licence files. The root Apache-2.0 licence does
not replace them.

| Component | Path | Licence | Licence file |
|---|---|---|---|
| Restruct SDLC (default development submodule; the application build and runtime do not use it) | `.agents/skills/restruct-sdlc/` | MIT, Copyright (c) 2026 Vitruvian Technology, Corp. | `.agents/skills/restruct-sdlc/LICENSE`, `.agents/skills/restruct-sdlc/sdlc-management/LICENSE` |

Restruct is an unmodified Git submodule from
https://github.com/technicallypete/restruct-sdlc, pinned by this repository's Git
link and registered in `.gitmodules`. Its MIT licenses remain in the initialized
submodule; initializing it downloads the pinned upstream contents. Updates are
explicit reviewed dependency pin changes. No locally transformed Restruct files
are shipped by this repository.

### Lock and module files

The lock and module files ship as files: `web/package-lock.json`, `core/go.mod`
and `core/go.sum`, `intelligence/uv.lock`, `infra/validation/tools/go.mod` and the
Terraform `.terraform.lock.hcl`. They pin the dependencies in sections 2 and
3; the dependencies themselves are not bundled.

## 2. Downloaded build and runtime dependencies (not in this repository)

### npm (`web/package.json`, `web/package-lock.json`)

The licence values below are the package-declared `license` fields recorded in
`web/package-lock.json` (lockfile version 3). They were not checked against each
package's own licence file. The lock file holds 759 package entries: 149 runtime
and 610 development-only. Every entry declares a licence.

| Licence (as declared) | All | Runtime |
|---|---|---|
| MIT | 618 | 96 |
| ISC | 42 | 14 |
| Apache-2.0 | 40 | 20 |
| MPL-2.0 | 14 | 0 |
| BSD-2-Clause | 12 | 0 |
| LGPL-3.0-or-later | 10 | 10 (optional, see section 3) |
| BSD-3-Clause | 9 | 2 |
| Apache-2.0 AND LGPL-3.0-or-later | 3 | 3 (optional, see section 3) |
| BlueOak-1.0.0 | 3 | 0 |
| 0BSD | 2 | 1 |
| Apache-2.0 AND LGPL-3.0-or-later AND MIT | 1 | 1 (optional, see section 3) |
| MIT AND ISC | 1 | 1 |
| CC-BY-4.0 | 1 | 1 |
| Python-2.0 | 1 | 0 |
| CC0-1.0 | 1 | 0 |
| (MIT OR CC0-1.0) | 1 | 0 |

*Adopter review:* the CC-BY-4.0 entry is `caniuse-lite`, browser-support data
used by the build tooling. CC-BY-4.0 requires attribution if you redistribute it.

Direct dependencies (versions pinned in `web/package.json`; licences as declared in
the lock file):

| Package | Version | Licence | Scope |
|---|---|---|---|
| `@base-ui/react` | 1.9.0 | MIT | runtime |
| `class-variance-authority` | 0.7.1 | Apache-2.0 | runtime |
| `clsx` | 2.1.1 | MIT | runtime |
| `cmdk` | 1.1.1 | MIT | runtime |
| `embla-carousel-react` | 8.6.0 | MIT | runtime |
| `input-otp` | 1.5.0 | MIT | runtime |
| `lucide-react` | 1.52.0 | ISC | runtime |
| `next` | 16.4.1 | MIT | runtime |
| `next-themes` | 0.4.6 | MIT | runtime |
| `react` | 19.3.0 | MIT | runtime |
| `react-day-picker` | 9.14.0 | MIT | runtime |
| `react-dom` | 19.3.0 | MIT | runtime |
| `react-resizable-panels` | 4.15.0 | MIT | runtime |
| `recharts` | 3.10.1 | MIT | runtime |
| `server-only` | 0.0.1 | MIT | runtime |
| `sonner` | 2.0.8 | MIT | runtime |
| `tailwind-merge` | 3.7.0 | MIT | runtime |
| `tw-animate-css` | 1.4.0 | MIT | runtime |
| `@axe-core/playwright` | 4.13.0 | MPL-2.0 | development |
| `@playwright/test` | 1.63.0 | Apache-2.0 | development |
| `@tailwindcss/postcss` | 4.3.3 | MIT | development |
| `@types/node` | 24.19.1 | MIT | development |
| `@types/react` | 19.3.0 | MIT | development |
| `@types/react-dom` | 19.3.0 | MIT | development |
| `eslint` | 9.39.5 | MIT | development |
| `eslint-config-next` | 16.4.1 | MIT | development |
| `shadcn` | 4.22.1 | MIT | development (stylesheet only) |
| `tailwindcss` | 4.3.3 | MIT | development |
| `typescript` | 5.9.3 | Apache-2.0 | development |

The resolved closures of Base UI and the kit components' other packages (MIT, ISC,
BSD-3-Clause, 0BSD, `MIT AND ISC`) are listed with versions in
`web/THIRD_PARTY_NOTICES.md`.

### Go (`core/go.mod`, `core/go.sum`)

`core/go.mod` requires 13 direct and 24 indirect modules, 37 in total. Module files
hold no licence metadata. The licences below are as stated in
`docs/foundation/VENDORS.md`. Every other module is **not established offline**.

| Module | Version | Licence (repository-stated) |
|---|---|---|
| `github.com/aws/aws-sdk-go-v2` (+ `credentials`, `service/s3`) | 1.47.1 / 1.20.6 / 1.114.0 | Apache-2.0 |
| `github.com/MicahParks/jwkset` | 0.11.3 | Apache-2.0 |
| `github.com/MicahParks/keyfunc/v3` | 3.8.2 | Apache-2.0 |
| `github.com/golang-jwt/jwt/v5` | 5.3.1 | MIT |
| `github.com/jackc/pgx/v5` | 5.11.0 | MIT |
| `github.com/pressly/goose/v3` | 3.28.0 | MIT |
| `github.com/workos/workos-go/v10` | 10.5.0 | MIT |
| `golang.org/x/time` | 0.15.0 | BSD-3-Clause |
| `github.com/google/go-querystring` (indirect) | 1.2.0 | BSD-3-Clause |
| `github.com/modelcontextprotocol/go-sdk`, `golang.org/x/net`, `golang.org/x/text` and the remaining 23 indirect modules | as pinned | not established offline |

The validation and contracts tooling module (`infra/validation/tools/go.mod`)
requires `github.com/santhosh-tekuri/jsonschema/v6` 6.0.3 and one indirect
module. Licence: not established offline.

### Container images and source-built services

Dockerfiles and Compose files refer to public base and service images. Images are
pulled when you build or run, and MinIO and `mc` are built from upstream source at
image build time. None of these images or binaries is in this repository. Where
`docs/foundation/VENDORS.md` states a licence, it is repeated here. That statement
is repository documentation; it was not checked against upstream licence files
offline.

| Component | Licence as stated in `docs/foundation/VENDORS.md` | Note |
|---|---|---|
| PostgreSQL (`pgvector/pgvector` image) | PostgreSQL licence | — |
| Redis (`redis` image) | RSALv2/SSPLv1 for Redis 7.4+, with AGPLv3 added in Redis 8 | *Adopter review:* review licence fit before production use |
| MinIO server and `mc` (built from source; local development and validation only) | AGPL-3.0 | *Adopter review:* AGPL-3.0 obligations can apply if you distribute images containing it or offer it over a network; the community edition is archived. Re-evaluate before any non-local use |
| Go toolchain (`golang` images) | BSD-3-Clause | — |
| `osv-scanner`, `govulncheck` (dependency scan) | Apache-2.0; BSD-3-Clause | — |
| Other base and tool images (for example `node`, `alpine`, distroless, `docker`, `jq`, `gitleaks`, Playwright) | not established offline | — |

## 3. Optional components

These are fetched only when you enable the feature, platform or workflow named.

| Component | When fetched | Licence | Note |
|---|---|---|---|
| `@img/sharp-libvips-*`, `@img/sharp-wasm32`, `@img/sharp-win32-*` (optional npm packages of `next`'s image library) | only on a matching platform | LGPL-3.0-or-later; `Apache-2.0 AND LGPL-3.0-or-later`; `Apache-2.0 AND LGPL-3.0-or-later AND MIT` (as declared in the lock file) | *Adopter review:* LGPL obligations can apply if you distribute builds or images that contain these binaries |
| Intelligence service Python packages (`intelligence/pyproject.toml`, `intelligence/uv.lock`): `fastapi`, `uvicorn`, `jsonschema`, `httpx`; development `pytest`, `ruff`; 30 locked packages in total | intelligence-enabled configurations | not established offline (`uv.lock` records no licence metadata) | — |
| Python and `uv` images (`python`, `ghcr.io/astral-sh/uv`) | intelligence-enabled configurations | not established offline | — |
| WorkOS emulator (`ghcr.io/workos/emulate`) and nginx (test identity overlay) | identity browser and probe checks | MIT; BSD-2-Clause (as stated in `docs/foundation/VENDORS.md`) | never part of a production deployment |
| Terraform provider `registry.terraform.io/cloudflare/cloudflare` 5.28.1 | deployment tooling (`terraform init`) | not established offline (provider's own licence) | pinned in `infra/terraform/environments/example/.terraform.lock.hcl` (shipped unchanged) |
| Terraform provider `registry.terraform.io/hashicorp/google` 8.8.1 | deployment tooling (`terraform init`) | not established offline (provider's own licence) | as above |
| Terraform CLI image `hashicorp/terraform:1.16.7` | deployment tooling | not established offline | *Adopter review:* review the CLI's licence terms before use |
| Deployment and edge images (`caddy`, Google Cloud CLI image) | staging/preview deployment tooling | not established offline | — |
