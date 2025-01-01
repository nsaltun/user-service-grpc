# user-service-grpc

# Generate Key for JWT
Generate a private key and encode it to base64 and set the environment variable JWT_PRIVATE_KEY:
```bash
export JWT_PRIVATE_KEY=$(openssl ecparam -name prime256v1 -genkey -noout | base64 -w 0)
```

To see the output of the private key directly in the terminal:
```bash
openssl ecparam -name prime256v1 -genkey -noout | base64
```
