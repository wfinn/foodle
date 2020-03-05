# Foodle
Simple app for daily votes, (hopefully) useful for deciding where to eat.

## Installation
```
go get github.com/wfinn/foodle
```
## Features
**Comments**

You can add comments to your vote, statistics will still work (if you use the correct syntax).

Example: Beef Plz! (Burger)

**Static Files**

Static files will get compiled into the binary.

If you want to edit static files, edit them, run _go generate_ and commit _static.go_ with your changes.
To add a static file edit _filenames_ in _static/genstatic.go_

**Accounts**

When a user makes her first vote, her name will be tied to a secret cookie.
Deleting the Cookie will result in losing the account.

## TODO
- nicer errors
- confirm use of cookies and storage of data
- locks on the .json files
- character whitelist for names
