# llmtraining

This repository contains Go examples demonstrating how to optimize handling of Certificate Revocation Lists (CRLs) in enterprise PKI environments.

## CRL Partitioning Example

`crl_partitioning.go` shows a basic approach to partitioning CRLs by certificate serial number prefix. The program fetches the appropriate CRL segment based on the serial number and checks whether the certificate is revoked.

### Usage

1. Build the program:

   ```bash
   go build crl_partitioning.go
   ```

2. Run the program with the base URL where CRL partitions are hosted and the certificate serial number:

   ```bash
   ./crl_partitioning https://crl.example.com/ ab12cd34
   ```

The program outputs whether the certificate is revoked according to the fetched CRL partition.

