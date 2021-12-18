FROM golang:1.17-alpine as builder

WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o bin/seeker cmd/seeker/*.go


FROM scratch
WORKDIR /app
COPY --from=builder /app/bin .
COPY --from=builder /app/migrations ./migrations
ENTRYPOINT [ "./seeker" ]