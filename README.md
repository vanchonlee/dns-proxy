# AZ-Specific DNS Proxy

This DNS proxy is designed to filter DNS responses based on specific AWS Availability Zones (AZs). It's particularly useful for controlling traffic distribution in AWS environments, ensuring that an ingress only forwards requests to services within its own AZ.

## Features

- Acts as a DNS proxy, listening on port 53 (UDP)
- Filters IP addresses based on AWS Availability Zones
- Uses a configuration file to determine AZ IP ranges
- No AWS SDK dependency, improving performance and reducing API calls

## Prerequisites

- Go 1.15 or later

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

## Configuration

Create a file named `az_config.json` in the same directory as the executable. This file should contain the IP ranges for each Availability Zone. Here's an example:

```json
{
  "AZs": {
    "us-west-2a": ["10.0.0.0/16"],
    "us-west-2b": ["10.1.0.0/16"],
    "us-west-2c": ["10.2.0.0/16"]
  }
}
```

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
