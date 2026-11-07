# Jellyfin LoL-Esports Bridge

This Software is a lightweight server to provide a IPTV integration for the League of Legends Esports Schedule.

## AI Disclaimer

Cause of i focused on the Backend integration of the application I didn't really give a f*** for the Web UI in version
1.x therefore the components and App.tsx are AI generated.

This will change in a later version when i have more interest in web development again

## Sources

- esports-api.lolesports.com -> to retrieve event, schedule and team information
- twitch.tv -> twitch is the first provider integrated for channel streams.
- youtube.com -> (wip) this provider is not implemented jet but will be next.

## Getting Started
> [!NOTE]
> The Bridge is Configured through environment Variables.\
> The most important config is "ApplicationUrl" to tell Jellyfin where to find the Live Streams

### 1. Deploy the container image
```zsh
docker run ghcr.io/luca-naujoks/jellyfin-lolesports-streams -p 9090:80 -e 
```