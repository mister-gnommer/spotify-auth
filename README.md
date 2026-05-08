# spotify-auth

A tiny CLI tool that performs the Spotify OAuth 2.0 Authorization Code flow and prints a refresh token. Run it once on your laptop to get a token, then paste it into whatever app needs it.

---

## Why this exists

Spotify's OAuth requires a browser redirect. That's fine on a laptop, but awkward for headless servers. This tool handles the local OAuth dance so the actual server-side app never needs a browser — it just uses the refresh token you get from here.

---

## How it works

```
1. You run: spotify-auth --client-id=... --client-secret=... --scopes=...
2. A temporary HTTP server starts on http://127.0.0.1:8888
3. Your browser opens the Spotify authorization page
4. You log in and approve access
5. Spotify redirects to http://127.0.0.1:8888/callback
6. The tool exchanges the code for tokens and prints the refresh token
7. Paste it into your app's config — done
```

The refresh token has no expiry (unless you revoke it). You only need to run this once per account.

---

## Usage

```bash
spotify-auth \
  --client-id YOUR_CLIENT_ID \
  --client-secret YOUR_CLIENT_SECRET \
  --scopes "user-read-playback-state,user-read-currently-playing,playlist-read-private,playlist-read-collaborative"
```

Output:

```
Opening browser for Spotify authorization...
Waiting for callback on http://127.0.0.1:8888/callback

Refresh token:
AQD3y...your_token_here...Xk2

Paste this into your app config.
```

---

## Spotify app setup

Before running, register the redirect URI in your Spotify app:

1. Go to [developer.spotify.com/dashboard](https://developer.spotify.com/dashboard)
2. Open your app → **Edit Settings**
3. Under **Redirect URIs**, add: `http://127.0.0.1:8888/callback`
4. Save

`http://127.0.0.1` (loopback) is explicitly allowed by Spotify even though HTTP is otherwise required to be HTTPS.

---

## Scopes

Pass whatever scopes your app needs. Common ones:

| Scope | What it allows |
|---|---|
| `user-read-playback-state` | Read current playback state (track, shuffle, device) |
| `user-read-currently-playing` | Read currently playing track |
| `playlist-read-private` | Read private playlists |
| `playlist-read-collaborative` | Read collaborative playlists |
| `user-library-read` | Read saved tracks (Liked Songs) |

---

## Building

```bash
go build -o spotify-auth .
```

Requires Go 1.21+.

---

## Notes

- This tool is intentionally single-purpose — it produces a refresh token and exits. No persistent state, no config file, no daemon mode.
- Port `8888` is hardcoded. If it's in use, kill whatever's on it first.
- Run this on your local machine only, not on a server.
