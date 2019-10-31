FROM alpine:latest

EXPOSE 9000/tcp

COPY build/bin/jenkins-hook-proxy /root/
WORKDIR /root

ENTRYPOINT ["./jenkins-hook-proxy"]
