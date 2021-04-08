# Dora: a git backed web server

Dora is a web server that serves static content from a
git repository. 

It is known to work with GitHub, although it probably also works with Gitlab or Gitea.

You supply it with a repository url and a deploy key. On
start-up Dora clones the repo and starts serving files.

Optionally you can configure a webhook to get update-on-push functionality.

By default, Dora only serves files from a sub-directory named
"public". You can override this, see Configuration below.

## Configuration

Dora is configured via the following environment variables:

 - REPO_URL: the repo url in the format git@github.com:myaccount/myrepo.git

 - BRANCH: the branch to checkout, defaults to "main"

 - DEPLOY_KEY: private key for repo access, should be a one-line string containing the base64 encoded private 
   ssh key. Configure the public part as a deploy key for your git repo.

 - HOOK_SECRET: secret for securing the git webhook. If not set any hook request is accepted.

 - HOOK_ENDPOINT: the path on which Dora listens for github webhooks, defaults to "/__pull"
  
 - DOCUMENT_ROOT: directory to serve from, defaults to "public"

 - CLONE_DIR: directory to clone repo in, defaults to "/var/www"

 - SITE_ROOT: path to serve the site from, defaults to "/"

 - PORT: port to listen on, defaults to 8080

 - BASICAUTH: optional, if set basic authentication is enabled for the entire site, except for the git webhook requests.
   Use the HOOK_SECRET to protect that endpoint. Should be set to a comma-separated list of user:password pairs, 
   e.g. "user1:pass1,user2:pass2"
   

## TLS support

Dora currently has no support for TLS out of the box. It is meant to be put behind a load balancer / proxy that 
takes care of TLS termination.
