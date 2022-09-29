FROM ubuntu:jammy

ENV DEBIAN_FRONTEND noninteractive

RUN apt-get -y update && \
  apt-get -y install openssl1.1 libffi-dev libssl-dev autoconf bison gperf ruby zlib1g-dev libyaml-dev curl build-essential

COPY entrypoint /entrypoint

ENTRYPOINT ["/entrypoint"]
