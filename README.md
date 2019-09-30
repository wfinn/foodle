# Foodle
Maybe this helps to decide what to eat...

Foodle is a simple app, tho there are some things which need documentation.
## "Accounts"
When a user makes her first vote, her name will be tied to a secret cookie.
Deleting the Cookie will result in losing the "account".
## CSRF
Foodle generates a random token for every browser.
The token will be set as Cookie and appear as parameter in votes.
Only if those 2 match, a vote is valid.
