# Distributed greeter tasks

## Manuals

- Basic HTTP server: https://tutorialedge.net/golang/creating-simple-web-server-with-golang/
- REST + SQL service: https://blog.logrocket.com/how-to-build-a-rest-api-with-golang-using-gin-and-gorm/ 
- JWT auth: https://dev.to/omnisyle/simple-jwt-authentication-for-golang-part-1-3kfo

## Tasks

- **(POSTPONED)** Use refresh tokens
  - Needs too much work on the UI side
- **(DONE)** Fix UI to use the new tokens/rest endpoints
- **(DONE)** Use GIN for the REST layer
- **(DONE)** Extract the jwt auth as a shared module
- Store jwt tokens in a cookie
  - Will change how /greet works
  - Will change how /login, /logout works
    - How to communicate back to the browser the cookie is invalid?
- **(DONE)** Add "favorite language" to greeter
- **(DONE)** Add "delete user" and "get user details" to login service
- Add messaging communication to greeter to sync state
  - Portable protocol: AMQP (supported by AWS and Azure services)
  - Quick solution:
    - KubeMQ?
    - **(DONE)** Redis?
  - Event Log (e.g. Kafka)
    - Likely not needed
  - *Q*: Guarantee that events are not missed
    - When is a pub/sub topic cleared of stored events? (so that a service can re-boot and re-consume them)
    - Perhaps Kafka is needed after all?
- Add persistence DB's to the greeter and login
  - SQL + transactions
  - db init containers (?)
- **(DONE)** Add UI for login service to delete the user account
- Add readiness probes
  - Ready once Redis is available
- Add some debug logs:
  - Gin logs the requests, but there's need for more
  - *Q*: How to mix the Gin logs which are structured in a particular way with my logs?
- Fine grained handling of JWT token parsing errors
  - E.g. expired tokens must not fail a call to `/logout`
  - *Note*: A chance to learn modern-day error handling in Go
- **(PROGRESS)** Add a Helm chart
  - **(DONE)** Customizable images/replica count
  - Customizable ingress: annotations to select something that's not "nginx"
  - Customizable secrets for the security token
  - Optional deployment of redis in case one is provided by the cloud is used
  - Optional debug services
- Move Redis from the Helm chart into it's own chart
- Setup a JS dev environment for the UI
- **(DONE)** Fix the REST API
  - At login return: `{ token: opaque, loginId: UUID, userId: int }`
  - To logout `DELETE /login/<UUID>`
  - Remove `/users/current`
  - `PUT /users/<id>` to store preferences
  - *Q*: What happens to the refresh token?
    - *A*: Nothing. It's not used to represent the primary login
      - It is used to `POST /refresh`
      - ... similarly to `POST /logins`
      - ... to create a new resource `/logins/<UUID>`
- Fix the Go builds to use Make:
  - Make outside of container for dev
  - Dockerfile to call Make inside gobuild container for prod/k8s
  - [https://danishpraka.sh/2019/12/07/using-makefiles-for-go.html]
- Fix the UI build to use Make
  - Perhaps make it into a Go server of static content?

## Create an Azure version with the "research security architecture"

- *NOTE*: Architecture
  - Use azure mysql
  - Use azure redis
  - Use Managed Identities to talk to mysql and redis
  - Use K8S Secure Store CSI driver to store JWT secret in Azure Key Vault

- Create a greeter environment with all components internal like in the dev version
- Modify it to have a manged db and plain password
- Modify it to use manged redis (may have to create a TF module)
- Modify db to use MSI
- (?) Modify redis to use MSI
- Modify to load the JWT secret from a file mount, not an env var
- Modify to store the JWT secret in Azure Key Vault
- Look for ways to make the JWT secret generated and inaccessible to users
