# Service Payment Repository

This repository contains the Service Payment application, a robust platform built using **Golang** and **gRPC**. The application is designed to handle various payment transactions efficiently, including sending and receiving payments.

## Key Features
- **Transaction Management**: Supports sending and receiving payments with detailed transaction tracking.
- **gRPC Integration**: Utilizes gRPC for efficient communication between services.
- **Docker Support**: Easily deployable using Docker and Docker Compose.
- **Modular Architecture**: Designed for scalability and maintainability.

## Getting Started

### Prerequisites
- **Docker**: Required for containerized deployment.
- **Golang**: Necessary for local development and running the service.
- **grpcurl**: Useful for testing gRPC endpoints.

### Installation
1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/service-payment.git
   ```
2. Navigate to the project directory:
   ```bash
   cd service-payment
   ```
3. Build the Docker image:
   ```bash
   docker build -t service-payment:latest .
   ```
4. Start the services using Docker Compose:
   ```bash
   docker-compose up
   ```

### Usage
- The service is accessible via HTTP on `localhost:8020` and gRPC on `localhost:50051`.
- Use `grpcurl` to test gRPC endpoints.

## Contributing
Contributions are welcome! Please fork the repository and submit a pull request for any enhancements or bug fixes.

## License
This project is licensed under the MIT License.

## Contact
For questions or support, please contact the maintainers or open an issue in the repository.