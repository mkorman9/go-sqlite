FROM debian:bullseye AS metadata
RUN apt update -qq && \
    apt install -y ca-certificates && \
    useradd -u 10001 app

FROM scratch
WORKDIR /
COPY --from=metadata /etc/passwd /etc/passwd
COPY --from=metadata /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY go-sqlite .
USER app
CMD ["./go-sqlite"]
