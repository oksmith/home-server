# Blockchain Network Node

A peer-to-peer blockchain node that can mine blocks, relay transactions, and synchronize with other nodes.

## What It Does

This program runs a blockchain node with the following capabilities:

- **Stores a full blockchain** - Validates and stores all blocks and transactions
- **Mines blocks** - Can mine new blocks with proof-of-work
- **Peer networking** - Connects to other nodes and exchanges data
- **Transaction relay** - Receives and broadcasts transactions
- **Chain synchronization** - Automatically adopts the longest valid chain
- **HTTP API** - Exposes endpoints for interaction

## Quick Start

### Running a Single Node

```bash
go run main.go -port 8080
```

### Running a 3-Node Network

**Terminal 1 (Bootstrap Node):**
```bash
go run main.go -port 8080
```

**Terminal 2:**
```bash
go run main.go -port 8081 -peers localhost:8080
```

**Terminal 3:**
```bash
go run main.go -port 8082 -peers localhost:8080,localhost:8081
```

## Command Line Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-port` | 8080 | Port to run the node on |
| `-peers` | "" | Comma-separated list of peer addresses |
| `-difficulty` | 3 | Mining difficulty (number of leading zeros) |
| `-reward` | 50.0 | Mining reward in coins |

## API Endpoints

### GET /chain
Returns the full blockchain.

```bash
curl http://localhost:8080/chain
```

### GET /peers
Lists connected peers.

```bash
curl http://localhost:8080/peers
```

### POST /peers
Manually add a peer.

```bash
curl -X POST http://localhost:8080/peers \
  -H "Content-Type: application/json" \
  -d '{"peer":"localhost:8083"}'
```

### GET /balance?address=ADDRESS
Get the balance for an address.

```bash
curl "http://localhost:8080/balance?address=abc123..."
```

### POST /mine
Mine a new block (includes mining reward).

```bash
curl -X POST http://localhost:8080/mine
```

### POST /transaction
Submit a transaction (used internally by nodes, transactions must be signed).

```bash
curl -X POST http://localhost:8080/transaction \
  -H "Content-Type: application/json" \
  -d '{"from":"...","to":"...","amount":10}'
```

### POST /block
Receive a block from a peer (used internally by nodes).

## Experiments

### Experiment 1: Basic Mining

Start two nodes and mine blocks on each:

```bash
# Terminal 1
go run main.go -port 8080

# Terminal 2
go run main.go -port 8081 -peers localhost:8080

# Terminal 3 - Mine on Node 1
curl -X POST http://localhost:8080/mine

# Check both nodes have the block
curl http://localhost:8080/chain | jq '.blocks | length'
curl http://localhost:8081/chain | jq '.blocks | length'
```

Both should show the same chain length. Block propagation works!

### Experiment 2: Late Joiner

See how a new node catches up with an existing network:

```bash
# Terminal 1
go run main.go -port 8080

# Mine several blocks
curl -X POST http://localhost:8080/mine
curl -X POST http://localhost:8080/mine
curl -X POST http://localhost:8080/mine

# Check chain length
curl http://localhost:8080/chain | jq '.blocks | length'
# Output: 4 (genesis + 3 mined)

# Terminal 2 - Start new node
go run main.go -port 8081 -peers localhost:8080

# Check it synced automatically
curl http://localhost:8081/chain | jq '.blocks | length'
# Output: 4 (synced on startup)
```

### Experiment 3: Fork and Resolution

Create competing chains and watch consensus emerge:

```bash
# Start two nodes
# Terminal 1
go run main.go -port 8080

# Terminal 2
go run main.go -port 8081 -peers localhost:8080

# Disconnect them (restart Node 2 without peers)
# In Terminal 2: Ctrl+C, then:
go run main.go -port 8081

# Mine on both (creates fork)
curl -X POST http://localhost:8080/mine  # Terminal 3
curl -X POST http://localhost:8081/mine  # Terminal 4

# Check they have different blocks at position 1
curl http://localhost:8080/chain | jq '.blocks[1].hash'
curl http://localhost:8081/chain | jq '.blocks[1].hash'
# Different hashes = fork!

# Reconnect them
curl -X POST http://localhost:8080/peers \
  -H "Content-Type: application/json" \
  -d '{"peer":"localhost:8081"}'

# Mine one more block on Node 1 (makes it longer)
curl -X POST http://localhost:8080/mine

# Check Node 2's chain - should match Node 1 now
curl http://localhost:8081/chain | jq '.blocks[1].hash'
curl http://localhost:8080/chain | jq '.blocks[1].hash'
# Same hash = consensus!
```

### Experiment 4: Peer Discovery

Watch automatic peer discovery in action:

```bash
# Terminal 1
go run main.go -port 8080

# Terminal 2
go run main.go -port 8081 -peers localhost:8080

# Check Node 1's peers (should be empty initially)
curl http://localhost:8080/peers
# Output: []

# Mine a block on Node 2 (triggers peer discovery)
curl -X POST http://localhost:8081/mine

# Check Node 1's peers again
curl http://localhost:8080/peers
# Output: ["localhost:8081"]

# Node 1 learned about Node 2 automatically!
```

### Experiment 5: Chain Validation

See what happens with invalid blocks:

```bash
# Start a node and mine blocks
go run main.go -port 8080
curl -X POST http://localhost:8080/mine
curl -X POST http://localhost:8080/mine

# Get the chain
curl http://localhost:8080/chain > chain.json

# Edit chain.json manually (change a transaction amount)
# Try to load it back (would fail validation)
```

### Experiment 6: Network Mesh

Build a mesh network:

```bash
# Terminal 1
go run main.go -port 8080

# Terminal 2
go run main.go -port 8081 -peers localhost:8080

# Terminal 3
go run main.go -port 8082 -peers localhost:8081

# Terminal 4
go run main.go -port 8083 -peers localhost:8082

# Mine on any node
curl -X POST http://localhost:8083/mine

# Check all nodes got the block
curl http://localhost:8080/chain | jq '.blocks | length'
curl http://localhost:8081/chain | jq '.blocks | length'
curl http://localhost:8082/chain | jq '.blocks | length'
curl http://localhost:8083/chain | jq '.blocks | length'
```

### Experiment 7: Balance Tracking

Watch balances change as blocks are mined:

```bash
# Start nodes
go run main.go -port 8080
go run main.go -port 8081 -peers localhost:8080

# Get Node 1's wallet address from startup output
# Look for: "Wallet Address: abc123..."

# Check initial balance (should be 0)
curl "http://localhost:8080/balance?address=abc123..."

# Mine a block (gives mining reward to Node 1)
curl -X POST http://localhost:8080/mine

# Check balance again (should be 50.0)
curl "http://localhost:8080/balance?address=abc123..."

# Mine another block
curl -X POST http://localhost:8080/mine

# Balance should be 100.0
curl "http://localhost:8080/balance?address=abc123..."
```

## Understanding the Output

When you start a node, you'll see:

```
[localhost:8080] Added peer: localhost:8081
[localhost:8080] Syncing with peers...

=== NODE INFO ===
Address: localhost:8080
Wallet Address: a720089a51dc4beeab0d10ac50111c8ae4b467062c47bf6f5f1933429546e946
Chain Length: 1 blocks
Balance: 0.00 coins
Peers: [localhost:8081]

[localhost:8080] Starting server...
```

When mining:
```
[localhost:8080] Mining block with 0 transactions...
Mined block 1 with 1 transactions (nonce: 1922)
[localhost:8080] Mined block 1!
```

When receiving blocks from peers:
```
[localhost:8081] Replacing chain with longer chain (length: 2)
```

## Tips

- **Use `jq` for JSON formatting:** Install with `brew install jq` (Mac) or `apt-get install jq` (Linux)
- **Watch logs in real-time:** Keep terminal windows visible to see peer interactions
- **Chain length indicates sync:** All nodes should have the same chain length after sync
- **Mining takes time:** Difficulty 3 mines in ~1-5 seconds, difficulty 4 takes ~30 seconds
- **Genesis block is block 0:** Chain length 1 means only genesis block exists

## Troubleshooting

**"Empty reply from server" when mining:**
- The node panicked. Check the terminal running the node for error messages.
- Common cause: Synced chain has nil maps. Restart the node.

**Blocks not propagating:**
- Check peer lists: `curl http://localhost:8080/peers`
- Peers might not be bidirectional. Mine a block to trigger peer discovery.

**Nodes have different chain lengths:**
- They're forked or not connected. Check peer lists.
- Mine one more block on the longer chain to trigger sync.

**Node won't start:**
- Port already in use. Use a different port or kill the existing process.

## What's Next

This is a learning implementation. Production blockchains add:
- Transaction fees and mempool prioritization
- Difficulty adjustment algorithms
- Merkle trees for efficient verification
- SPV (light) clients
- Network security (DDoS protection, peer scoring)
- Persistent storage (databases instead of in-memory)
- Proper transaction signing via wallet API