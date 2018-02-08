FROM golang:latest as builder
WORKDIR /home/stephan/src/trigger-proxy/
COPY app.go .
COPY example.csv mapping.csv
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM alpine:latest  
WORKDIR /root/
COPY --from=builder /home/stephan/src/trigger-proxy/app .
COPY --from=builder /home/stephan/src/trigger-proxy/mapping.csv .
CMD ["./app"]  
