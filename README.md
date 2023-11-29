# Songomatic

Songomatic is midi tune / drum loop generator to help with music composition inspiration. What songomatic generates is almost never useful by itself, but is intended to be something you use and edit - maybe to try something you would never have thought to try. It helps get you over the "blank page" syndrome.

Songomatic currently does not use AI, it is akin to rolling a set of dice to create melodies and drum loops.

![screen shot](docs/screen.png)

## Running

Using a different environment variable set for prod

```bash
docker run --env-file=.env.production -p 8080:3000 robrohan/songomatic
```
