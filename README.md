# Foodle
Maybe this helps to decide what to eat...

## Static Files
Static Files will get compiled into the binary.
If you want to edit static files, edit them, run _go generate_ and commit _static.go_ with your changes.
To add a static file edit _filenames_ in _static/genstatic.go_
## "Accounts"
When a user makes her first vote, her name will be tied to a secret cookie.
Deleting the Cookie will result in losing the "account".
## Comments
You can add comments to your vote, statistics will still work (if you use the correct syntax).

Example: Beef Plz! (Burger)
## TODO
- locks on the .json files
- whitelist for names
