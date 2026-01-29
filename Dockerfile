### Description: Dockerfile for ingress-traefik-converter
FROM alpine:3.23

COPY ingress-traefik-converter /

# Starting
ENTRYPOINT [ "/ingress-traefik-converter" ]