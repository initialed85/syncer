# syncer

Here lies my failed attempt to make an amazing thing for live-ish syncing code changes from your dev laptop to your dev server.

## What went wrong

- `fsnotify/fsnotify` was unable to track the volume of files I needed to track (a monorepo's worth)
- Recursive directory walks are either too slow when single threaded or too heavy when multi threaded

## What I'll probably try

- Second guess the limitations `fsnotify/fsnotify` came up against and see if I can implement just a macOS filesystem watch for lots of
  files in a very non-portable way

## Usage

```shell
# TODO: I really don't recommend this but I'm here to party
sudo launchctl limit maxfiles 1048576 8388608
```
