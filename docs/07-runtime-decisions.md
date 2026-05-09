# Runtime Decisions

This document records the early runtime choices for Kelompok.

## Product Name

Use `Kelompok` as the product and repository name.

The project is a clean open-source reset, not a continuation of an unreleased overbuilt version.

Tagline:

```text
The Solutions of Movement
```

## Backend

Use Go for the backend API, CLI, workers, and service integrations.

Reasons:

- Small deployment footprint
- Fast CLI development
- Efficient background jobs
- Good fit for self-hosting
- Clear module boundaries
- Easy static binary distribution

## Frontend

Use SvelteKit for the frontend.

Reasons:

- Fast development for public profiles and CRM screens
- Less boilerplate than React-heavy stacks
- Good SSR for public organization, event, and donation pages
- Vite-based local development
- Good self-hosting path through adapter-node

Recommended stack:

- SvelteKit
- TypeScript
- Tailwind CSS
- shadcn-svelte or a small internal component library
- OpenAPI-generated client when the backend API stabilizes

Next.js is still acceptable if the project later needs deeper React ecosystem compatibility, but SvelteKit is the default recommendation for speed and simplicity.

## Default Ports

Use ports that are uncommon across the user's other projects:

```text
API: 4621
Web app: 4622
Worker metrics/internal diagnostics: 4623
PostgreSQL Docker port: 54621
```

## License

Use Apache-2.0.

This is the best fit for broad adoption by communities, NGOs, universities, public-interest groups, and companies that may want to self-host or contribute without heavy license friction.

## Product Positioning

Kelompok should first work as a CRM:

- Create organization
- Manage organization profile
- Publish public organization page
- Claim organization
- Create event
- Create donation campaign
- Publish impact report
- Access API and CLI

The public project should stay focused on CRM, public profiles, events, donation campaigns, impact reporting, API, and CLI workflows.
