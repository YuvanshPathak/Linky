FROM golang:1.20.4-alpine AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY ./ ./

RUN go build -o main.exe main.go

FROM scratch
WORKDIR /app
COPY --from=build /app/main.exe ./
COPY --from=build /app/go-templates/ ./go-templates/
EXPOSE 4000
CMD [ "./main.exe" ]