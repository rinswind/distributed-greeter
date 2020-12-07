# Distributed greeter tasks

- (POSTPONED) Use refresh tokens
  - Needs too much work on the UI side
- (DONE) Fix UI to use the new tokens/rest endpoints
- Use DB for users in login service
- Use GIN for the REST layer
- Extract jwt cache as a shared module
- Store jwt tokens in a cookie
  - Will change how /greet works
  - Will change how /login, /logout works
    - How to communicate back to the browser the cookie is invalid?
- Add "favorite language" to greeter
- Add "delete user" and "get user details" to login service
  - Add messaging communication to greeter to sync state
- Add UI for logger to list and delete users
- Split UI into login and greeter
- Add readiness probes
  - Ready once Redis is available
