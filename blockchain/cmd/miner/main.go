package main

import (
	"fmt"
	"log"

	"github.com/oksmith/home-server/blockchain/pkg/chain"
	"github.com/oksmith/home-server/blockchain/pkg/mempool"
	"github.com/oksmith/home-server/blockchain/pkg/transaction"
	"github.com/oksmith/home-server/blockchain/pkg/wallet"
)

func main() {
	fmt.Println("=== BLOCKCHAIN WITH TRANSACTIONS ===")

	// Create blockchain
	difficulty := 3
	miningReward := 50.0
	bc := chain.New(difficulty, miningReward)
	mp := mempool.New()

	// Create wallets
	fmt.Println("Creating wallets...")
	alice, err := wallet.New()
	if err != nil {
		log.Fatal(err)
	}
	bob, err := wallet.New()
	if err != nil {
		log.Fatal(err)
	}
	charlie, err := wallet.New()
	if err != nil {
		log.Fatal(err)
	}
	miner, err := wallet.New()
	if err != nil {
		log.Fatal(err)
	}

	aliceAddr := alice.Address()
	bobAddr := bob.Address()
	charlieAddr := charlie.Address()
	minerAddr := miner.Address()

	fmt.Printf("Alice:   %s\n", aliceAddr[:16]+"...")
	fmt.Printf("Bob:     %s\n", bobAddr[:16]+"...")
	fmt.Printf("Charlie: %s\n", charlieAddr[:16]+"...")
	fmt.Printf("Miner:   %s\n\n", minerAddr[:16]+"...")

	// Register public keys with the blockchain
	bc.RegisterPublicKey(aliceAddr, alice.PublicKey)
	bc.RegisterPublicKey(bobAddr, bob.PublicKey)
	bc.RegisterPublicKey(charlieAddr, charlie.PublicKey)
	bc.RegisterPublicKey(minerAddr, miner.PublicKey)

	// Block 1: Mine to give Alice some coins
	fmt.Println("Block 1: Mining reward to Alice...")
	if err := bc.AddBlock([]*transaction.Transaction{}, aliceAddr); err != nil {
		log.Fatal(err)
	}
	printBalances(bc, aliceAddr, bobAddr, charlieAddr, minerAddr)

	// Block 2: Alice sends to Bob
	fmt.Println("\nBlock 2: Alice sends 15 coins to Bob...")
	tx1 := transaction.New(aliceAddr, bobAddr, 15.0)
	if err := tx1.Sign(alice.PrivateKey); err != nil {
		log.Fatal(err)
	}
	mp.Add(tx1)

	if err := bc.AddBlock(mp.GetAll(), minerAddr); err != nil {
		log.Fatal(err)
	}
	mp.Clear()
	printBalances(bc, aliceAddr, bobAddr, charlieAddr, minerAddr)

	// Block 3: Multiple transactions
	fmt.Println("\nBlock 3: Multiple transactions...")
	tx2 := transaction.New(bobAddr, charlieAddr, 5.0)
	if err := tx2.Sign(bob.PrivateKey); err != nil {
		log.Fatal(err)
	}

	tx3 := transaction.New(aliceAddr, charlieAddr, 10.0)
	if err := tx3.Sign(alice.PrivateKey); err != nil {
		log.Fatal(err)
	}

	mp.Add(tx2)
	mp.Add(tx3)

	if err := bc.AddBlock(mp.GetAll(), minerAddr); err != nil {
		log.Fatal(err)
	}
	mp.Clear()
	printBalances(bc, aliceAddr, bobAddr, charlieAddr, minerAddr)

	// Display blockchain
	fmt.Println("\n=== BLOCKCHAIN SUMMARY ===")
	for _, block := range bc.Blocks {
		fmt.Printf("\nBlock #%d (Hash: %s...)\n", block.Index, block.Hash[:16])
		fmt.Printf("  Transactions: %d\n", len(block.Transactions))
		for i, tx := range block.Transactions {
			if tx.IsCoinbase() {
				fmt.Printf("    %d. COINBASE -> %s: %.2f coins\n", i+1, tx.To[:16]+"...", tx.Amount)
			} else {
				fmt.Printf("    %d. %s... -> %s...: %.2f coins\n",
					i+1, tx.From[:16], tx.To[:16], tx.Amount)
			}
		}
	}

	// Validate blockchain
	fmt.Printf("\nBlockchain valid? %v\n", bc.IsValid())

	// Test: Try to spend more than you have
	fmt.Println("\n=== TESTING INSUFFICIENT FUNDS ===")
	invalidTx := transaction.New(charlieAddr, bobAddr, 1000.0)
	if err := invalidTx.Sign(charlie.PrivateKey); err != nil {
		log.Fatal(err)
	}
	mp.Add(invalidTx)

	fmt.Printf("Attempting to send 1000 coins (Charlie only has %.2f)...\n", bc.GetBalance(charlieAddr))
	if err := bc.AddBlock(mp.GetAll(), minerAddr); err != nil {
		fmt.Printf("Transaction rejected: %v\n", err)
	} else {
		fmt.Println("ERROR: Invalid transaction was accepted!")
	}

	mp.Clear()
	printBalances(bc, aliceAddr, bobAddr, charlieAddr, minerAddr)

	// Test: what happens if you don't sign the transaction before adding it to mempool
	fmt.Println("\n=== TESTING UNSIGNED TRANSACTION ===")
	unsignedTx := transaction.New(aliceAddr, bobAddr, 5.0)

	if err := mp.Add(unsignedTx); err != nil {
		fmt.Printf("Mempool rejected: %v\n", err)
	} else {
		if err := bc.AddBlock(mp.GetAll(), minerAddr); err != nil {
			fmt.Printf("Blockchain rejected: %v\n", err)
		}
	}
	mp.Clear()

	// Test: Double spending
	fmt.Println("\n=== TESTING DOUBLE SPENDING ===")
	doubleSpendTx1 := transaction.New(aliceAddr, bobAddr, 20.0)
	if err := doubleSpendTx1.Sign(alice.PrivateKey); err != nil {
		log.Fatal(err)
	}

	doubleSpendTx2 := transaction.New(aliceAddr, charlieAddr, 20.0)
	if err := doubleSpendTx2.Sign(alice.PrivateKey); err != nil {
		log.Fatal(err)
	}

	mp.Add(doubleSpendTx1)
	mp.Add(doubleSpendTx2)

	fmt.Printf("Attempting double spend (Alice has %.2f, trying to spend 40.0)...\n", bc.GetBalance(aliceAddr))
	if err := bc.AddBlock(mp.GetAll(), minerAddr); err != nil {
		fmt.Printf("Double spending rejected: %v\n", err)
	} else {
		fmt.Println("ERROR: Double spending was accepted!")
	}
	mp.Clear()

}

func printBalances(bc *chain.Chain, addresses ...string) {
	fmt.Println("Balances:")
	names := []string{"Alice", "Bob", "Charlie", "Miner"}
	for i, addr := range addresses {
		fmt.Printf("  %s: %.2f coins\n", names[i], bc.GetBalance(addr))
	}
}
