# Trakt-netflix

Hacky program to sync Netlfix with Trakt.

## Usage

This is meant to be run as a cron job.

On the first run, the script will mark as watched the last 20 movies/episodes you watched on Netflix. To avoid that, you can create a file named `data` at the root of the project / next to the binary, and put the last thing you watched in it. For a movie, just add the name of the movie in it much like it is on [Netflix's watchlist](https://www.netflix.com/settings/viewed) `<movie name>`. If it's a show, the format is `<show name>: <episode name>`. For example `The Stranded: The gate`. Unlike the list on netflix, there are no double-quotes around the episode name, and the name of the show is only marked once.

### Environment variables
| ENV | Required | Format | Info |
| --- | --- | --- | --- |
| NETFLIX_COOKIE | required |  | Value of the `NetflixId` cookie |
| NETFLIX_ACCOUNT_ID | optional |  | Can be found everywhere in the local storage, usually in a `MDX_*` object. If not set, it will use the last account used with the provided cookie. |
| TRAKT_COOKIE | required |  | Value of the `_traktsession` cookie |
| TRAKT_CSRF | required | | Value of the `x-csrf-token` token. You can get it by opening the Network tab in Chrome, triggering Trakt private API, and looking at the headers of a request such as `episodes.json`. `X-Csrf-Token` should be at the bottom of the headers list. You can trigger their API by adding an item to your watchlist by clicking on Add To List on a poster, after doing so you'll see a `watchlist` request in the requests list. [Click here](https://trakt.tv/search/movies/?query=lord+of+the+ring) for quick access to a list of movies to add to your watchlist.
| SLACK_WEBHOOKS | optional | webhook1,webhook2 | |
