FROM golang:1.22.1 as builder

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 make api
RUN CGO_ENABLED=0 make consumer

FROM alpine

COPY --from=builder /app/build/api /app/api
COPY --from=builder /app/build/consumer /app/consumer

EXPOSE 1313

CMD ["/app/api"]
