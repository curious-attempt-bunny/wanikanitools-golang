
# wanikanitools-golang

Experimental alternate implementation of https://github.com/curious-attempt-bunny/wanikanitools.

## Running locally

  go get github.com/curious-attempt-bunny/wanikanitools-golang
  cd $GOPATH/src/github.com/curious-attempt-bunny/wanikanitools-golang
  go build main.go pages.go assignments.go review_statistics.go subjects.go url.go dashboard.go leech.go summary.go cors.go
  WANIKANI_V2_API_KEY=xxx PORT=5000 ./main

## Deploying to Dokku

### On your Dokku server

  dokku apps:create wanikanitools-golang
  dokku config:set --no-restart wanikanitools-golang WANIKANI_V2_API_KEY=xxx
  dokku config:set --no-restart wanikanitools-golang GIN_MODE=release

Optionally:

  dokku config:set --no-restart wanikanitools-golang NEW_RELIC_LICENSE_KEY=xxx NEW_RELIC_APP_NAME=yyy

### Locally

  git remote add dokku dokku@YOUR_HOST_IP:wanikanitools-golang
  git push dokku head

### Adding HTTPS (on Dokku server)

  sudo dokku plugin:install https://github.com/dokku/dokku-letsencrypt.git
  dokku config:set --no-restart wanikanitools-golang DOKKU_LETSENCRYPT_EMAIL=yourregistrationemail.com
  dokku letsencrypt wanikanitools-golang
  dokku letsencrypt:cron-job --add    