module github.com/N0F1X3d/todo/db-service

go 1.25.3

require (
	github.com/golang-migrate/migrate/v4 v4.19.1
	github.com/lib/pq v1.11.2
	github.com/redis/go-redis/v9 v9.18.0
	github.com/stretchr/testify v1.11.1
	google.golang.org/grpc v1.78.0
)

require (
	github.com/alicebob/miniredis/v2 v2.36.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/yuin/gopher-lua v1.1.1 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

require (
	github.com/BurntSushi/toml v1.2.1 // indirect
	github.com/N0F1X3d/todo/pkg v0.0.0
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/ilyakaznacheev/cleanenv v1.5.0
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/text v0.31.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251029180050-ab9386a59fda // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	olympos.io/encoding/edn v0.0.0-20201019073823-d3554ca0b0a3 // indirect
)

replace github.com/N0F1X3d/todo/pkg => ../pkg
