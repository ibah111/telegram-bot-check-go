FROM golang:1.25 AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/bot ./main.go


FROM gcr.io/distroless/static:nonroot

WORKDIR /app
COPY --from=build /out/bot /app/bot

USER nonroot:nonroot
ENTRYPOINT ["/app/bot"]
