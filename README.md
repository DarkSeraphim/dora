# Dora the web server 

Dora is a web server that serves static content from a
github repository.

You supply it with a github repository url and a deploy key. On
start-up Dora clones the repo and starts serving files.

Optionally you can configure a github webhook to get 
update-on-push.

By default Dora only serves files from a sub-directory named
"public", but you can override this.

## Configuration

Dora is configured via the following environment variables:

 - REPO_URL: the repo url in the format git@github.com/...

 - BRANCH: the branch to checkout, defaults to "main"

 - DEPLOY_KEY: private key for repo access

 - HOOK_SECRET: secret for securing the git webhook. If not set any hook request is accepted.
   
 - DOCUMENT_ROOT: directory to serve from, defaults to 'public'

 - HOOK_ENDPOINT: the path on which Dora listens for github webhooks, defaults to "/__pull"
