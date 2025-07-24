# wordfeud-helper

Een simpele en snelle Wordfeud helper die woorden matcht op basis van opgegeven letters en een patroon.

## Functies

- Matcht Nederlandse woorden op patroon met wildcard-ondersteuning:
  - `%` - nul of meer letters uit je letterset
  - `?` - precies één letter uit je letterset
  - `*` - nul of meer willekeurige letters
  - `&` - precies één willekeurige letter
- Scoreberekening op basis van Scrabble/Wordfeud-punten
- Webinterface (HTML + Go backend)
- Ondersteuning voor jokers (via `?` in letterset)
- Geschikt voor mobiel

## Import `wordlist.txt`

```sh
$ sh import.sh
```

## Running

```sh
$ go build -v
$ wordfeud-helper wordlist.txt
```

## Licentie

Zlib