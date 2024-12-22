FROM golang:1.23 AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/nutriPrice ./main.go

FROM scratch
COPY --from=build /bin/nutriPrice /bin/nutriPrice
CMD ["/bin/nutriPrice"]