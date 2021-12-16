FROM golang:1.17-alpine as builder

WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o cmd/seeker/app cmd/seeker/*.go


FROM scratch
WORKDIR /app
COPY --from=builder /app/cmd/seeker .
ENTRYPOINT [ "./app" ]