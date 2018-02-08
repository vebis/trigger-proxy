FROM golang:latest as builder
WORKDIR /src/
COPY app.go .
COPY example.csv mapping.csv
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM alpine:latest  
WORKDIR /root/
COPY --from=builder /src/app .
COPY --from=builder /src/mapping.csv .
CMD ["./app"]  
