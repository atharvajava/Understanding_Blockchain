package main

import (

	// converting Bytes to Hex String
	"crypto/sha256" // SHA256 --
	"encoding/hex"  // for Hexadecimal encoding
	"encoding/json" // for generating bytes
	"fmt"           // debug
	"math/rand"     // Generate random numbers
	// Timestamp to generate random seed
)

// We will be using function to create a hash function to create a fingerprint for each transactions-
// This hash function links each of our blocks to each other. Below is the helper class to wrap our GO
// Hash function

func hashme(s interface{}) string {
	bytes, _ := json.Marshal(s)
	h := sha256.New()
	h.Write(bytes)
	return hex.EncodeToString(h.Sum(nil))
}

// Helper function to generate sign
func sign() int {
	s := -1 + rand.Intn(2)
	if s == 0 {
		s++
	}
	return s
}

// Next, we want to create a function to generate exchanges between Atharva and Monish.
// We will indicate withdrawals with negative numbers, and deposits with positive numbers.
// we will construct our transaction to always be between the two users of our system, and make sure
// that the deposit is the same magnitude as the withdrawal- i.e that we are neither creating nor
// nor destroying any money
func makeTransaction(maxValue int32) map[string]int {
	rand.Seed(0)
	sign := sign()
	amount := rand.Intn(3) + 1
	arveePays := sign * amount
	moniPays := -1 * arveePays
	return map[string]int{"Atharva": arveePays, "Monish": moniPays}
}

//Next step: making our very own blocks! We’ll take the first k
//transactions from the transaction buffer, and turn them into a block
// Before we do that we need to define method for checking validity of the transactions we ve pulled
// into the block. For bitcoin the validation function checks that the input values are valid
// unspent transaction output. ie input is no greater then the output. In, Ethereum the validation
//function checks that if the smart contracts are successfully executed or not
// We wont build that kind of complex system we will just build a simple set of rules
// for basic token system.
//The sum of deposits and withdrawals must be 0 (tokens are neither created nor destroyed)
// A user’s account must have sufficient funds to cover any withdrawals

func updateState(txn map[string]int, state map[string]int) map[string]int {
	for key, _ := range txn {
		if _, ok := state[key]; ok {
			state[key] += txn[key]
		} else {
			state[key] = txn[key]
		}
	}
	return state
}

func isValidTxn(txn map[string]int, state map[string]int) bool {
	//Assume that the transaction is a dictionary keyed by account names
	var sum int
	for _, val := range txn {
		sum += val
	}
	//Check that the sum of the deposits and withdrawals is 0
	if sum != 0 {
		return false
	}
	//Check that the transaction does not cause an overdraft
	var acctBalance int
	for key, _ := range txn {
		if val, ok := state[key]; ok {
			acctBalance = val
		} else {
			acctBalance = 0
		}
		if acctBalance+txn[key] < 0 {
			return false
		}
	}
	return true
}

//For each block, we want to collect a set of
//transactions, create a header, hash it, and add it to the chain

func makeBlock(txns []map[string]int, chain []map[string]interface{}) map[string]interface{} {
	parentBlock := chain[len(chain)-1]
	parentHash := parentBlock["hash"]
	contents := parentBlock["contents"].(map[string]interface{})
	blockNumber := contents["blockNumber"].(int)
	txnCount := len(txns)
	blockContents := map[string]interface{}{"blockNumber": blockNumber + 1, "parentHash": parentHash, "txnCount": txnCount, "txns": txns}
	blockhash := hashme(blockContents)
	block := map[string]interface{}{"hash": blockhash, "contents": blockContents}
	return block
}

func checkBlockHash(block map[string]interface{}) {
	contents := block["contents"].(map[string]interface{})
	blockNumber := contents["blockNumber"].(int)
	expectedHash := hashme(contents)
	if block["hash"] != expectedHash {
		fmt.Println("Hash does not match %d", blockNumber)
	}
	return
}

func checkBlockValidity(block map[string]interface{}, parent map[string]interface{}, state map[string]int) map[string]int {
	/*We want to check the following conditions:
	    - Each of the transactions are valid updates to the system state
	    - Block hash is valid for the block contents
	    - Block number increments the parent block number by 1
		- Accurately references the parent block's hash
	*/
	parentContent := parent["contents"].(map[string]interface{})
	parentNumber := parentContent["blockNumber"].(int)
	parentHash := parent["hash"]
	blockContent := block["contents"].(map[string]interface{})
	blockNumber := blockContent["blockNumber"].(int)
	//Check transaction validity; throw an error if an invalid transaction was found.
	blockTxn := blockContent["txns"].([]map[string]int)
	for _, txn := range blockTxn {
		if isValidTxn(txn, state) {
			state = updateState(txn, state)
		} else {
			fmt.Println("Invalid Transaction in the block ", blockNumber, txn)
		}
	}
	checkBlockHash(block)
	if blockNumber != (parentNumber + 1) {
		fmt.Println("Hash does not match contents of the block ", blockNumber)
	}

	if blockContent["parentHash"] != parentHash {
		fmt.Println("Hash does not match contents of the block", parentHash)
	}

	return state
}

func checkChain(chain []map[string]interface{}) map[string]int {
	/*
			  Work through the chain from the genesis block (which gets special treatment),
		      checking that all transactions are internally valid,
		      that the transactions do not cause an overdraft,
		      and that the blocks are linked by their hashes.
		      This returns the state as a dictionary of accounts and balances,
			  or returns False if an error was detected
	*/
	//Data input processing: Make sure that our chain is a array of Maps
	state := map[string]int{}
	//  Prime the pump by checking the genesis block
	//  We want to check the following conditions:
	//  - Each of the transactions are valid updates to the system state
	//  - Block hash is valid for the block contents
	chainContents := chain[0]["contents"].(map[string]interface{})
	txns := chainContents["txns"].([]map[string]int)
	for _, txn := range txns {
		state = updateState(txn, state)
	}
	parent := chain[0]
	//  Checking subsequent blocks: These additionally need to check
	//     - the reference to the parent block's hash
	//     - the validity of the block number
	for _, block := range chain[1:] {
		state = checkBlockValidity(block, parent, state)
		parent = block
	}
	return state
}

func main() {
	txnBuffer := []map[string]int{}
	for i := 1; i <= 30; i++ {
		txnBuffer = append(txnBuffer, makeTransaction(3))
	}
	state := map[string]int{"Atharva": 100, "Monish": 100}
	//fmt.Println(isValidTxn(map[string]int{"Atharva": -4, "Monish": 2, "Arpit": 2}, state))
	genesisBlockTxns := []map[string]int{state}
	genesisBlockContents := map[string]interface{}{"blockNumber": 0, "parentHash": "", "txnCount": 1, "txns": genesisBlockTxns}
	genesisHash := hashme(genesisBlockContents)
	genesisBlock := map[string]interface{}{"hash": genesisHash, "contents": genesisBlockContents}
	//genesisBlockStr, _ := json.Marshal(genesisBlock)
	chain := []map[string]interface{}{genesisBlock}
	//Let’s use this to process our transaction buffer into a set of blocks:
	blockSizeLimit := 5 //Arbitrary number of transactions per block- this is chosen by the block miner, and can vary between blocks!
	for len(txnBuffer) > 0 {
		//bufferStartSize := len(txnBuffer)
		var txnList []map[string]int
		for len(txnBuffer) > 0 && len(txnList) < blockSizeLimit {
			newTxn := txnBuffer[len(txnBuffer)-1]
			txnBuffer = txnBuffer[:len(txnBuffer)-1]
			validTxn := isValidTxn(newTxn, state)

			if validTxn {
				txnList = append(txnList, newTxn)
				state = updateState(newTxn, state)
			} else {
				continue
			}
		}

		myBlock := makeBlock(txnList, chain)
		chain = append(chain, myBlock)
	}
	fmt.Println(checkChain(chain))
	/*
		Now that we know how to create new blocks and link them together into a chain,
		let’s define functions to check that new blocks are valid- and that the whole chain is valid.
		On a blockchain network, this becomes important in two ways:
		When we initially set up our node, we will download the full blockchain history.
		After downloading the chain, we would need to run through the blockchain to compute
		the state of the system. To protect against somebody inserting invalid transactions
		in the initial chain, we need to check the validity of the entire chain in this initial download.
		Once our node is synced with the network (has an up-to-date copy of the blockchain and a
		representation of system state) it will need to check the validity of new blocks that
		are broadcast to the network.
		We will need three functions to facilitate in this:
		checkBlockHash: A simple helper function that makes sure that the block contents match the hash
		checkBlockValidity: Checks the validity of a block, given its parent and the current system state.
		We want this to return the updated state if the block is valid, and raise an error otherwise.
		checkChain: Check the validity of the entire chain, and compute the system state beginning at
		the genesis block. This will return the system state if the chain is valid, and raise an
		error otherwise.
	*/

}
