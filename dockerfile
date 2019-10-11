FROM golang:1.12.7-alpine

# install os dependencies
RUN apk update && apk add --no-cache bash=5.0.0-r0 curl=7.65.1-r0 openssh=8.0_p1-r0 make=4.2.1-r2 git=2.22.0-r0 g++=8.3.0-r0 musl-dev=1.1.22-r3 pkgconf=1.6.1-r1 busybox-extras wget
# python python-dev py-pip build-base

LABEL Name="payment-service"
LABEL Version="0.5.0"

ENV APPDIR=app
RUN mkdir ${APPDIR}
ENV WORKDIR=/go/${APPDIR}

WORKDIR ${WORKDIR}
COPY . .

# datadog agent installation
#RUN DD_API_KEY=9b70a596c9045c496384fcd812222a32 bash -c "$(curl -L https://raw.githubusercontent.com/DataDog/datadog-agent/master/cmd/agent/install_script.sh)"
# COPY datadog.yaml /etc/datadog-agent/datadog.yaml
# COPY process.conf /etc/datadog-agent/conf.d/
# start datadog-agent
# add user deployer
RUN adduser -D deployer deployer
# set directory & file owner
RUN chown -R deployer:deployer .
RUN chmod +x payment-service
# set user as deploy

USER deployer
#EXPOSE 8100
EXPOSE 9200
#EXPOSE 8090

ENTRYPOINT ["/go/app/payment-service"]
#ENTRYPOINT ["./start.sh"]
