# AZ-Specific DNS Proxy

This DNS proxy is designed to filter DNS responses based on a specific AWS Availability Zone (AZ). It's particularly useful for controlling traffic distribution in AWS environments, ensuring that an ingress only forwards requests to services within its own AZ.

## Features

- Acts as a DNS proxy, listening on port 53 (UDP)
- Filters IP addresses based on a specified AWS Availability Zone
- Uses AWS SDK to query EC2 network interfaces for AZ information
- Configurable via environment variables

## Prerequisites

- Go 1.15 or later
- AWS account with appropriate permissions
- AWS CLI configured with the necessary credentials

## Installation

1. Clone the repository:
   ```
   git clone https://github.com/vanchonlee/dns-proxy.git
   cd dns-proxy
   ```

2. Install dependencies:
   ```
   go mod tidy
   ```

3. Build the application:
   ```
   go build -o az-dns-proxy
   ```

## Usage

1. Set the target Availability Zone:
   ```
   export TARGET_AZ=us-west-2b
   ```

2. Run the DNS proxy:
   ```
   sudo ./az-dns-proxy
   ```

   Note: Running on port 53 requires root privileges.

## Configuration

The application is configured using the following environment variables:

- `TARGET_AZ`: The target AWS Availability Zone (e.g., "us-west-2b")

## AWS Permissions

The application requires AWS permissions to describe EC2 network interfaces. Ensure that the AWS credentials used have the following permissions:

- `ec2:DescribeNetworkInterfaces`

## Integrating with Ingress Controllers

To use this DNS proxy with your ingress controllers:

1. Deploy the DNS proxy in your cluster or on a dedicated instance with access to EC2 metadata.
2. Configure each ingress controller to use the DNS proxy for name resolution.
3. Ensure that the DNS proxy has the necessary AWS permissions.

## How It Works

1. When an ingress controller sends a DNS query, the proxy determines the AZ of the ingress controller based on its IP address.
2. The proxy resolves the requested domain name to IP addresses.
3. It then filters the IP addresses to only include those in the same AZ as the requesting ingress controller.
4. The filtered IP addresses are returned in the DNS response.

This ensures that each ingress controller only receives IP addresses for services in its own AZ.

## Limitations

- The proxy currently only handles A records (IPv4 addresses).
- It assumes it's running in an AWS environment with access to EC2 metadata.
- Performance may be impacted by API rate limits when querying EC2 information.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
