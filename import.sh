#!/bin/sh

# update wordlist-repository
git submodule init
git submodule update

# try renaming old wordlist
mv wordlist.txt wordlist.txt.old 2> /dev/null || true

# write new wordlist
# sort basiswoorden-gekeurd, otherwise `comm` will complain
# grep only letter-words which are longer than 2 characters
# remove all roman-digits
# uniq-sort and write
sort opentaal-wordlist/elements/basiswoorden-gekeurd.txt |
    grep -E '^[a-zA-Z]{2,}$' |
    comm -23 - opentaal-wordlist/elements/romeinse-cijfers.txt |
    tr '[:upper:]' '[:lower:]' |
    sort -u > wordlist.txt
