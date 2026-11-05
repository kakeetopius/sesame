FROM golang:alpine AS builder
RUN mkdir /build 
ADD . /build/
WORKDIR /build 
RUN CGO_ENABLED=0 GOOS=linux go build -a  -o sesame ./main.go 

FROM scratch
COPY --from=builder /build/sesame /app/
WORKDIR /app
CMD ["./sesame"]
