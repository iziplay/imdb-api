# Iziplay IMDB API

## Concepts

### Storage

For more simplicity, this uses a Postgres database (and Gorm as ORM under the hood)

### API

This API provides some routes to let newer projects use the fully-typed response when querying an IMDB title

#### OMDB emulation

This API is able to provide the same type of response as OMDB but there is some caveats:

- it misses some fields
  - writers
  - plot
  - posters
  - ...
- is not fully compliant
  - types
  - ...
- it misses some queries
  - cannot get by title
  - cannot search

### Auto-sync

This app auto-sync 1 time per day (at launch or 24 hours after launch if it was sync before) by downloading the TSV on IMDB datasource

## Why?

### Auto sync (no more invalid IMDB title for newer titles)

When using some other projects, the synchronization with IMDB was sometimes outdated because:

- the sync was started manually by the owner
- the sync was less than 1/month

### Typings

There were no good typing for others APIs and it was awful to test any IMDB type to get all response types

### Statistics

There were no API with statistics about the current synchronization state or counts to validate data