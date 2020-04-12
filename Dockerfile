FROM golang:latest as build
LABEL stage=builder
WORKDIR /app
COPY . .

# Build the Go app
RUN go get -d ./...
RUN CGO_ENABLED=0 GOGC=off go build -o main .


FROM scratch
COPY --from=build /app/main /app/main
COPY ./static /app/static
# Expose port 8080 to the outside world
EXPOSE 8080

WORKDIR /app
# Command to run the executable
CMD ["./main"]


#FROM golang:latest as build
#WORKDIR /home/ethohampton/go/src/ethohampton.com/Shortener
#COPY . ./test
#RUN CGO_ENABLED=0 GOGC=off go build -o shortener -ldflags -s

#FROM scratch
#COPY --from=build /home/ethohampton/go/src/ethohampton.com/Shortener/shortener /usr/local/bin/shortener
#ENTRYPOINT [ "/usr/local/bin/shortener" ]
