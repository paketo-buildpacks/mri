FROM paketobuildpacks/build-jammy-full

ENV DEBIAN_FRONTEND noninteractive

ARG cnb_uid=0
ARG cnb_gid=0

USER ${cnb_uid}:${cnb_gid}

RUN apt-get -y update && \
  apt-get -y upgrade && \
  apt-get -y install rustc && \
  rm -rf /var/lib/apt/lists/* /tmp/* /etc/apt/preferences

COPY entrypoint /entrypoint

ENTRYPOINT ["/entrypoint"]
