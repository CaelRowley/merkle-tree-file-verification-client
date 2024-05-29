# Merkle Tree File Verification Client

A simple CLI tool built with Go for demonstrating the functionality of a Merkle tree. The tool allows users to create test files, generate a Merkle tree from these files, upload files to a server, download files from the server along with their Merkle proofs, and verify file integrity using the stored Merkle tree root hash.

This tool is written to be used in conjuction with the [Merkle Tree File Verification Backend](https://github.com/CaelRowley/merkel-tree-file-verification-backend).


## Commands

- **Create Test Files**
  - Deletes all existing locally created test files.
  - Creates a specified number of test files based on user input.

- **Generate Merkle Tree**
  - Generates a Merkle tree from the test files and stores the root hash in memory.

- **Upload Test Files**
  - Clears all files stored on the server.
  - Uploads all local test files to the server.

- **Download and Verify File**
  - Downloads a file from the server along with its Merkle proof.
  - Verifies the integrity of the downloaded file using the Merkle proof and the stored root hash.

- **Corrupt a File on Server**
  - Simulates file corruption on the server by modifying the data while keeping a reference to the original hash.
  - Demonstrates how the client's verification process detects file tampering using a Merkle proof.

- **Delete Test Files**
  - Deletes all locally created test files.

- **Delete Downloads**
  - Deletes all locally downloaded test files.

- **Exit**
  - Closes the CLI tool.

## Environment Variables

- **SERVER_URL**
  - This specifies the URL of the backend server that the client should communicate with.
  - Defaults to `http://localhost:8080` (points to the default URL of the [backend service]((https://gitlab.com/CaelRowley/merkel-tree-file-verification-backend)) locally).
  - If you want to connect to a deployed backend, set `SERVER_URL` to the URL of your backend:
    ```bash
    export SERVER_URL=http://your-backend.com
    ```

## Getting Started

After cloning the repository you can use this tool by running it locally or by using Docker.

### Running locally

**⚠️ Prerequisites:** Make sure you have [Go](https://go.dev/doc/install) installed on your machine.

```bash
go run ./app/main.go
```

### Running with Docker Compose

**⚠️ Prerequisites:** Make sure you have [Docker](https://docs.docker.com/desktop/) and [Docker Compose](https://docs.docker.com/compose/install/) installed on your machine.


**Step 1:** Build the Docker Image (if needed)

```bash
docker-compose -f docker-compose.yml build
```

**Step 2:** Run the client

```bash
docker-compose -f docker-compose.yml run client
```

### Example Workflow

1. **Create Test Files**
   - `Create Test Files` to generate test files.

2. **Generate Merkle Tree**
   - `Generate Merkle Tree` to compute the Merkle tree and store the root hash.

3. **Upload Test Files**
   - `Upload Test Files` to upload test files to the server.

4. **Download and Verify File**
   - `Download and Verify File` to download a file and its Merkle proof from the server.
   - Also verifies the file's integrity using the downloaded Merkle proof and the stored root hash.

5. **Corrupt a File on Server**
   - `Corrupt a File on Server` to modify a file on the server.

6. **Download and Verify File**
   - `Download and Verify File` to download the file you corrupted.
   - Note that the Merkle proof validation failed as the hash generated from the proof doesn't match with the stored root hash.

7. **Cleanup**
   - Clean up local files and downloads using `Delete Test Files` and `Delete Downloads` as needed.
