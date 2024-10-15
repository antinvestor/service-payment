# Payment Service

This repository contains the code for the Payment Service built with **Golang** and **gRPC**. The service handles transactions, including sending and receiving payments. This guide will walk you through setting up, running, and testing the Payment Service locally.

---

## Prerequisites

Ensure you have the following installed on your system:

- **Docker** (for running services and databases in containers)
- **Golang** (for running or building the service locally)
- **grpcurl** (for testing the gRPC service)

You can install grpcurl by following the instructions [here](https://github.com/fullstorydev/grpcurl).

---

## Building and Running the Payment Service

### Building the Docker Image

```bash
docker build -t service-payment:latest .
```

This command builds a Docker image named `service-payment:latest` from the current directory.

### Starting the Service Using Docker Compose

```bash
docker-compose up
```

This command starts the Docker Compose environment, which includes the Payment Service and its dependencies (e.g., PostgreSQL). The service will be available on `localhost:8020` for HTTP and `localhost:50051` for gRPC.

### Testing the Payment Service

#### Preparing the gRPC Request

```bash
grpcurl -plaintext -d '{
  "data": {
    "id": "123",
    "transaction_id": "tx_001",
    "reference_id": "ref_001",
    "batch_id": "batch_001",
    "amount": {
      "currency_code": "USD",
      "units": 100,
      "nanos": 0
    }
  }
}' localhost:8020 payment.v1.PaymentService/Send
```

This command sends a gRPC request to the `Send` method with the provided payment data. The expected response will confirm whether the payment was successfully sent.

### Configuration

The Payment Service can be configured using environment variables or by modifying the `config.yml` file. Key configurations include:

* `GRPC_SERVER_PORT`: The port on which the gRPC server runs (default: 50051)
* `HTTP_SERVER_PORT`: The port on which the HTTP server runs (default: 8020)
* `POSTGRES_URI`: The URI for connecting to the PostgreSQL database

Make sure to configure these properly in the `docker-compose.yml` or environment files before running the service.

### Troubleshooting

* **Port conflicts:** If you encounter an error like `bind: address already in use`, ensure no other services are using port 8020 or 50051. You can stop conflicting services or adjust the port mappings in the `docker-compose.yml` file.
* **Invalid token error:** If the service uses authentication and you receive token errors, make sure to properly configure your JWT token or disable authentication for local testing by adjusting the configuration in `config.yml`.

### License

This project is licensed under the terms of the MIT License.

### Contact

For any issues or questions, feel free to reach out to the maintainers or open an issue in the repository.


