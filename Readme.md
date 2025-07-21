# Instruction
This project is learning quic server

## User Guide
```bash
# Generate random data
mkdir datasource
dd if=/dev/urandom of=datasource/random_1mb.bin bs=1M count=1

# Run server and serve specific folder as data source
go run main.go -dir datasource
```