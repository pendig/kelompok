# Product Vision

Kelompok is a free and open-source platform for organizations that need public visibility, basic CRM features, event workflows, donor transparency, and impact reporting.

Tagline: The Solutions of Movement.

The first audience is organizations. Events and donors are important, but they should orbit the organization profile instead of becoming separate products too early.

## Mission

Make organization infrastructure free, transparent, and self-hostable.

Kelompok should help organizations:

- Be discoverable through public profiles
- Keep basic public data accurate
- Claim and manage their own profile
- Show impact and SDGS alignment
- Run simple events
- Open donor campaigns
- Publish fund usage updates
- Publish articles, news, announcements, and activity updates
- Integrate with external tools through API and CLI

## Target Users

- Local community organizations
- NGOs
- Foundations
- Student organizations
- Volunteer networks
- Social impact programs
- Public-interest communities
- International nonprofits looking for lightweight profile and reporting tools

## Product Pillars

### 1. Data and Public Profile

Each organization should get an auto-generated public page.

Public profile content may include:

- Name
- Slug
- Logo
- Banner
- History
- Description
- Leadership
- Contact email
- Official website
- Social links
- Location
- SDGS mapping
- Impact report
- Public posts or updates
- Public events
- Donation campaigns
- Source evidence from imports and external references

The first useful product is a CRM and public profile system that works through manual creation, imports, claims, and API/CLI workflows.

### 2. Posts, Articles, and Public Updates

Organizations should be able to publish public content from their profile.

Early post features:

- Organization news
- Activity updates
- Articles
- Announcements
- Impact stories
- Draft and published status
- Tags and categories

Posts should be simple at first. They are not a full CMS replacement, but they let organizations keep their public profile alive and communicate progress.

### 3. Event Management and Ticketing

Events should begin as simple public pages connected to organizations.

Early event features:

- Create public pages for discovered events
- Let registered organizations create events manually
- Support free or paid ticket options later
- Keep the first version simple enough to ship quickly

Future features can include:

- Registration forms
- QR tickets
- Attendance and check-in
- Invoices
- Payment gateway integration
- Organizer dashboard
- Event reports

### 4. Donor Management

Donor management should begin with transparent donation campaigns and reports.

Early donor features:

- Organization can open a donation campaign
- Campaign has goal, description, timeline, and public status
- Organization can publish updates on fund usage
- Public visitors can see donation reports and impact updates

Future features:

- Donor CRM
- Donor receipts
- Recurring donation records
- International NGO recommendations
- Related organization recommendation engine
- SDGS-based donor matching

## Product Boundary

Kelompok should avoid becoming too large too early.

In the first release, do not rebuild every feature from older repositories. The early platform should prioritize:

- Organization CRUD
- Public organization pages
- Claim flow
- Profile editing
- Post and article publishing
- Event creation
- Public event pages
- Donation campaign creation
- Public donor reporting
- API and CLI foundations

Donor recommendations, AI workflows, QR check-in, and complex ticketing should come later.

## Product Principles

- Public profiles should be useful before login
- Registered organizations should get control without complex onboarding
- Every source record should keep evidence
- Claim flow must be auditable
- Manual edits should not destroy original source evidence
- Dynamic data is allowed, but stable data should stay queryable
- API and CLI should support future AI workflows from day one
