# Instruction
This project is learning quic server

## User Guide
```bash
go init quic-server
go get github.com/quic-go/quic-go

# Generate random data
dd if=/dev/urandom of=random_1mb.bin bs=1M count=1
```