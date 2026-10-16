FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
RUN go build -o /out/releaserail ./cmd/releaserail

FROM alpine:3.20
RUN adduser -D -g '' releaserail
USER releaserail
COPY --from=build /out/releaserail /usr/local/bin/releaserail
ENTRYPOINT ["/usr/local/bin/releaserail"]
CMD ["healthcheck"]
