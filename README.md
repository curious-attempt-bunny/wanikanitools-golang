
# wanikanitools-golang

## Deploying to Dokku

### On your Dokku server

  dokku apps:create wanikanitools-golang

### Locally

  git remote add dokku dokku@YOUR_HOST_IP:wanikanitools-golang
  git push dokku head