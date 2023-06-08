# Gunakan image golang:alpine sebagai base image
FROM golang:alpine

# Set working directory di dalam container
WORKDIR /crm

# Salin file go.mod dan go.sum ke dalam container
COPY go.mod go.sum ./

# Download dependensi Go
RUN go mod download

# Salin kode sumber aplikasi ke dalam container
COPY . .

# Build aplikasi Go
RUN go build -o crm

# Menjalankan aplikasi saat container dijalankan
CMD ["./crm"]
