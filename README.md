# Songomatic

Songomatic is midi tune / drum loop generator to help with music composition inspiration. What songomatic generates is almost never useful by itself, but is intended to be something you use and edit - maybe to try something you would never have thought to try. It helps get you over the "blank page" syndrome.

Songomatic currently does not use AI, it is akin to rolling a set of dice to create melodies and drum loops.

![screen shot](docs/screen.png)

## Running

You can either check the code out and build it yourself (see the `Makefile`), or you can just run the [docker container](https://hub.docker.com/repository/docker/robrohan/songomatic/general) if you like.

```bash
docker run -p 8080:3000 robrohan/songomatic
```

then browse to http://localhost:3000

### Ansible Example

```yml
- name: Start the songomatic container
  community.docker.docker_container:
    name: songomatic
    image: robrohan/songomatic
    state: started
    restart: true
    env:
      WB_DB_DRIVER: sqlite3
      WB_DB_CONNECTION: main.db
      WB_DB_POST: "PRAGMA synchronous = OFF;PRAGMA journal_mode = MEMORY;PRAGMA cache_size = -16000"
    ports:
      - "8080:3000"
```
