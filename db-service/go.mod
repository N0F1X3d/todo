module github.com/N0F1X3d/todo/db-service

go 1.25.3

require (
    google.golang.org/grpc v1.58.3
    github.com/lib/pq v1.10.9
    github.com/N0F1X3d/todo/gen/go v0.0.0
)

replace github.com/N0F1X3d/todo/gen/go => ../gen/go