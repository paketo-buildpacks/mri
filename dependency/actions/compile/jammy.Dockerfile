FROM paketobuildpacks/build-jammy-full

ENV DEBIAN_FRONTEND noninteractive

COPY entrypoint /entrypoint

ENTRYPOINT ["/entrypoint"]
