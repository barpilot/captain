FROM golang:1.11 as build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o /captain ./cmd/captain/
RUN chmod +x /captain

FROM alpine:3.8
RUN apk add --update-cache docker git openssh
COPY --from=build /captain /bin/captain
ENTRYPOINT [ "/bin/captain" ]
