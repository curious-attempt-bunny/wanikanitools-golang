FROM alpine:latest

MAINTAINER Edward Muller <edward@heroku.com>

WORKDIR "/opt"

ADD .docker_build/wanikanitools-golang /opt/bin/wanikanitools-golang
ADD ./templates /opt/templates
ADD ./static /opt/static

CMD ["/opt/bin/wanikanitools-golang"]

