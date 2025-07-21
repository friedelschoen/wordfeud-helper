## Update `wordlist.txt`

```sh
wget -O- https://github.com/OpenTaal/opentaal-wordlist/raw/refs/heads/master/wordlist.txt | grep -E '^[a-zA-Z]{2,}$' | tr '[:upper:]' '[:lower:]' | sort -u > wordlist.txt
```