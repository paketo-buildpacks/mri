FROM paketobuildpacks/build-bionic-full

ENV DEBIAN_FRONTEND noninteractive


COPY entrypoint /entrypoint

ENTRYPOINT ["/entrypoint"]
