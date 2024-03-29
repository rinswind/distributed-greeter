# Distributed greeter tasks

## Manuals

- [Basic HTTP server](https://tutorialedge.net/golang/creating-simple-web-server-with-golang)
- [REST + SQL service](https://blog.logrocket.com/how-to-build-a-rest-api-with-golang-using-gin-and-gorm)
- [JWT auth](https://dev.to/omnisyle/simple-jwt-authentication-for-golang-part-1-3kfo)

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
- **(DONE)** Add a Helm chart
  - **(DONE)** Customizable images/replica count
  - Customizable ingress: annotations to select something that's not "nginx"
  - Customizable secrets for the security token
  - Optional debug services
- **(DONE)** Move Redis from the Helm chart into it's own chart
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
- Secure Ingress
  - Fix helm charts to allow TLS configuration to be defined in the Ingress
  - Create self-signed certificate
  - Put certificate in AWS SSL
  - Configure Secrets Store CSI driver to load pub/priv certs and sync to a Secret resource
  - Configure Ingress to use TLS with that resource
  - *Q*: Does anything have to be done with domains?
  - Generate a self-signed cert for `example.org`
    `openssl req -x509 -newkey rsa:4096 -nodes -subj '/CN=greeter.saglive.cloud' -keyout tls.key -out tls.crt -sha256 -days 365`
  - Generate a TLS Secret
    `kubectl create secret tls -n greeter-aws greeter-saglive-cloud-secret --cert=tls.crt --key=tls.key`

## Secrets Management

- Add a "secrets mounter" pod to the CSI SecretsProviderClass to allow the export of Secret resources when needed

## Create an Azure version with the "suggested security architecture"

- *NOTE*: The suggested security architecture
  - Use azure mysql
  - Use azure redis
  - Use Managed Identities to talk to mysql and redis
  - Use K8S Secure Store CSI driver to store JWT secret in Azure Key Vault
  - Rotate secrets

- **(DONE)** Create a greeter environment with all components internal like in the dev version
- **(DONE)** Create a base-line case
  - Use Secret resources provisioned off-band in etcd
  - In PODs mount secrets as files
  - Separate DB initialization from DB creation to prepare for the use of a Managed DB
- **(DONE)** Try to secure the secrets in the dev-like version.
  - E.g. in a Azure KeyVault with the CSI driver to supply them.
  - *NOTE*: Setting a secret to access the MySQL schemas is an issue.
- **(DONE)** Modify it to have a manged db and plain password
- **(DONE)** Modify it to use manged redis (may have to create a TF module)
  - [https://docs.microsoft.com/en-us/azure/azure-cache-for-redis/cache-overview]
- **(DONE)** Modify db to use MSI
  - [https://docs.microsoft.com/en-us/azure/mysql/howto-configure-sign-in-azure-ad-authentication]
  - [https://docs.microsoft.com/en-us/azure/active-directory/managed-identities-azure-resources/how-to-use-vm-token#get-a-token-using-go]
  - [https://docs.microsoft.com/en-us/azure/mysql/howto-connect-with-managed-identity]
  - [https://pkg.go.dev/golang.org/x/oauth2]
- **(CANCELLED)** Modify redis to use MSI
  - Not supported by Azure, must use a secret
- **(DONE)** Modify to load the JWT secret from a file mount, not an env var
- **(DONE)** Modify to store the JWT secret in Azure Key Vault
- Look for ways to make the JWT secret generated and inaccessible to users
- Automate the setup setup for the MySQL
  - Requires modifications to the terraform config
  - Required sequence:
      1. Enable an AAD admin user for the MySQL
      2. Get the K8S agent-pool MSI
      3. Get the AAD user to
          - Create the app user identified by the MSI client ID
          - Create a DB and add admin right to the user

## Create an AWS version with the "suggested security architecture"

- Rewrite `mysql-init` in Go
  - That will generate random passwords
  - Or make a Terraform config to make Dbs?
  - That will create the given Db if it doesn't exist and generate a user/pass and store them in SSM

- *NOTE*: The suggested security architecture
  - Use RDS (rather than MySQL pods)
  - Use Elasticache for Redis (rather than Redis pods)
  - Use IAM to talk to RDS and MemoryDB
  - Use K8S Secure Store CSI driver to store JWT secret in AWS Secrets Manager
  - Rotate secrets

- Create a greeter environment with all components internal like in the dev version
- **(DONE)** Create a base-line case
  - Use Secret resources provisioned off-band in etcd
  - In PODs mount secrets as files
- **(DONE)** Secure the secrets in the dev-like version.
  - Load the CSI driver provider
  - Create a Helm chart to make CSI provider classes (for the respective secrets)
  - Populate it with the secrets
  - Give the CSI driver permissions to access the Secrets Manager
  - E.g. in a Azure KeyVault with the CSI driver to supply them.
