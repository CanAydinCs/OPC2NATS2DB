# OPC2NATS2DB

This GitHub project is a pipeline designed for an automated system. Its primary goal is to organize data produced by automated systems, send it to a NATS queue, and subsequently save it.

## Software Used in the System

### `opc_server.py`
**Purpose**: Simulate an OPC server by generating and broadcasting data.

**Pre-run Adjustments**:
- Update the IP address in line 6: `url = "opc.tcp://ip_address:4840"`. The default port for OPC servers is 4840.

**Execution**:
- Run the command: `python 'opc_server.py'`.

### `golang_service.go`
**Purpose**: Transmit data from the OPC server to the NATS queue. This code is adaptable and can work with different OPC servers without modification.

**Pre-run Adjustments**:
- Update the server IP address in line 20: `opcServerURL := "opc.tcp://ip_address:4840"`.
- Adjust lines 15 and 16 as needed: `'connectAndReadOPCUAAndPublish("ns=2;i=3")'`. Modify the `ns` and `i` values based on the ID of the generated data, which can be verified using the UaExpert application.
- Ensure both the OPC server and NATS server are running.
- To start the NATS server, execute the `nats-server.exe` application found in the directory starting with "nats-server" and wait for it to start.

**Execution**:
- Run the command: `go run 'golang_service.go'`.

### `golang_consumer.go`
**Purpose**: Retrieve data from the NATS queue and transfer it to a database. Currently, it reads the data and prints it to the console. Future updates will include writing to TimescaleDB.

**Pre-run Adjustments**:
- Ensure the NATS server is running.

**Execution**:
- Run the command: `go run 'golang_consumer.go'`.

---

By following these instructions, you can set up and run the components of the OPC2NATS2DB project, enabling the automated processing and storage of data from OPC servers through NATS queues.