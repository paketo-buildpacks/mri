FROM ubuntu:jammy

RUN apt update && apt install -y curl

COPY run.rb /test/run.rb

COPY entrypoint /entrypoint

ENTRYPOINT ["/entrypoint"]

WORKDIR /test
