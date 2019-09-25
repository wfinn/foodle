# Foodle
Maybe this helps to decide what to eat...

Foodle is a simple app, tho there are some things which need documentation.
## "Accounts"
When a user makes her first vote, her name will be tied to a secret cookie.
Deleting the Cookie will result in losing the "account".
## CSRF
Each text/html response contains a random token.
The token will be set as Cookie and appear as parameter in votes.
Only if those 2 match, a vote is valid.
