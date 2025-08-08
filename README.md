# Trakt-netflix

Cron job to sync Netflix with Trakt.

## Usage

This is meant to be run as a cron job.

On the first run, the script will mark as watched the last 20 movies/episodes you watched on Netflix.

###  Getting Started

You first need to create a Trakt app at https://trakt.tv/oauth/applications/new. There isn't really a way out of it, trying to parse the website directly causes *a ton* of mismatches.

The first time you start the service, it will prompt you to authenticate with Trakt. Follow the instructions to complete the authentication process. This will create a `trakt_auth.json` file in the current directory.

### Environment variables
| ENV | Required | Format | Info |
| --- | --- | --- | --- |
| NETFLIX_COOKIE | required |  | Value of the `NetflixId` cookie |
| NETFLIX_ACCOUNT_ID | optional |  | Can be found everywhere in the local storage, usually in a `MDX_*` object. If not set, it will use the last account used with the provided cookie. |
| TRAKT_REDIRECT_URI | required |  | Value of redirect URL of your trakt app, it won't be used but we still need to provide it to trakt. You can use http://localhost |
| TRAKT_CLIENT_ID | required |  | Client ID of your trakt app |
| TRAKT_CLIENT_SECRET | required | | Client Secret of your trakt app |
| SLACK_WEBHOOKS | optional | webhook1,webhook2 | |
| CRON_SPECS | optional | | Defaults to @hourly see [Wikipedia](https://en.wikipedia.org/wiki/Cron) for format, Non-standard format are also accepted |

### setup with Docker Compose

```yaml
version: "3"
services:
  trakt-netflix:
    image: ghcr.io/nivl/trakt-netflix
    container_name: "trakt-netflix"
    environment:
    - TRAKT_CLIENT_ID=xxx
    - TRAKT_CLIENT_SECRET=yyy
    - TRAKT_REDIRECT_URI=http://localhost
    - NETFLIX_ACCOUNT_ID=zzz
    - NETFLIX_COOKIE=aaa
    - SLACK_WEBHOOKS=bbb
    volumes:
      - /path/to/config:/config
```
