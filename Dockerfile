FROM golang:1.10 as build

WORKDIR /go/src/github.com/charlieegan3/calendars

COPY . .

RUN CGO_ENABLED=0 go build -o calendars main.go


FROM gcr.io/distroless/base
COPY --from=build /go/src/github.com/charlieegan3/calendars/calendars /

CMD ["/calendars"]
