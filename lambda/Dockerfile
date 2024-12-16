FROM golang:1.22 as build
WORKDIR /log-monitor
# Copy dependencies list
COPY go.mod go.sum ./
# Build with optional lambda.norpc tag
COPY lambda.go .
RUN go build -tags lambda.norpc -o main lambda.go
# Copy artifacts to a clean image
FROM public.ecr.aws/lambda/provided:al2023
COPY --from=build /log-monitor/main ./main
ENTRYPOINT [ "./main" ]