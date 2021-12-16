FROM golang:1.17-alpine as builder

WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o cmd/gatekeeper/app cmd/gatekeeper/*.go


FROM scratch
WORKDIR /app
COPY --from=builder /app/cmd/gatekeeper .
ENTRYPOINT [ "./app" ]