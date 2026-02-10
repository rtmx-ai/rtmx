# REQ-BRAND-001: GitHub Organization Avatar

## Requirement

GitHub organization avatar shall use RTMX icon.

## Description

Set the GitHub organization avatar (profile picture) for the rtmx-ai organization to match the RTMX icon used on rtmx.ai. This ensures brand consistency across all RTMX touchpoints.

## Acceptance Criteria

- [ ] Avatar uploaded to github.com/rtmx-ai organization settings
- [ ] Uses 256x256 PNG version of RTMX icon
- [ ] Visually matches favicon on rtmx.ai
- [ ] Appears correctly in GitHub UI (org page, repo listings, member lists)

## Icon Specifications

The RTMX icon consists of:
- Dark background (#1e293b) with rounded corners
- Three vertical bars representing RTM columns:
  - Left bar: Sky blue (#0ea5e9)
  - Center bar: Green (#22c55e) top, Amber (#f59e0b) bottom
  - Right bar: Sky blue (#0ea5e9)

## Assets Created

PNG icons generated from favicon.svg:
- `public/icons/rtmx-icon-256x256.png` (for GitHub)
- `public/icons/rtmx-icon-128x128.png`
- `public/icons/rtmx-icon-64x64.png`

## Manual Steps

1. Navigate to https://github.com/organizations/rtmx-ai/settings/profile
2. Click the organization avatar
3. Upload `rtmx-icon-256x256.png`
4. Save changes

## Verification

- [ ] Visit https://github.com/rtmx-ai and confirm icon displays
- [ ] Check organization appears in search with correct avatar
