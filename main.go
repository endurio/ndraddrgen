// Copyright (c) 2015 The Decred Developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/decred/dcrwallet/walletseed"
)

// The hierarchy described by BIP0043 is:
//  m/<purpose>'/*
// This is further extended by BIP0044 to:
//  m/44'/<coin type>'/<account>'/<branch>/<address index>
//
// The branch is 0 for external addresses and 1 for internal addresses.

// maxCoinType is the maximum allowed coin type used when structuring
// the BIP0044 multi-account hierarchy.  This value is based on the
// limitation of the underlying hierarchical deterministic key
// derivation.
const maxCoinType = hdkeychain.HardenedKeyStart - 1

// MaxAccountNum is the maximum allowed account number.  This value was
// chosen because accounts are hardened children and therefore must
// not exceed the hardened child range of extended keys and it provides
// a reserved account at the top of the range for supporting imported
// addresses.
const MaxAccountNum = hdkeychain.HardenedKeyStart - 2 // 2^31 - 2

// ExternalBranch is the child number to use when performing BIP0044
// style hierarchical deterministic key derivation for the external
// branch.
const ExternalBranch uint32 = 0

// InternalBranch is the child number to use when performing BIP0044
// style hierarchical deterministic key derivation for the internal
// branch.
const InternalBranch uint32 = 1

var curve = btcec.S256()

var params = chaincfg.MainNetParams

// Flag arguments.
var getHelp = flag.Bool("h", false, "Print help message")
var testnet = flag.Bool("testnet", false, "")
var simnet = flag.Bool("simnet", false, "")
var seed = flag.Bool("seed", false, "Generate an HD extended seed instead of a single keypair")
var verify = flag.Bool("verify", false, "Verify a seed by generating the first "+
	"address")
var key = flag.Bool("key", false, "Import private key")
var verbose = flag.Bool("v", false, "Print dev data")
var uncompressed = flag.Bool("u", false, "Print uncompressed keys instead")
var showPrivateKey = flag.Bool("p", false, "Print private key to console")

func setupFlags(msg func(), f *flag.FlagSet) {
	f.Usage = msg
}

var newLine = "\n"

// writeNewFile writes data to a file named by filename.
// Error is returned if the file does exist. Otherwise writeNewFile creates the file with permissions perm;
// Based on ioutil.WriteFile, but produces an err if the file exists.
func writeNewFile(filename string, data string, perm os.FileMode) error {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_EXCL, perm)
	if err != nil {
		return err
	}
	n, err := f.WriteString(data)
	if err == nil && n < len(data) {
		// There was no error, but not all the data was written, so report an error.
		err = io.ErrShortWrite
	}
	if err != nil {
		// There was an error, so close file (ignoring any further errors) and return the error.
		f.Close()
		return err
	}
	return f.Close()
}

// generateKeyPair generates and stores a secp256k1 keypair in a file.
func generateKeyPair(generate, verbose, uncompressed, showPrivateKey bool) (string, error) {
	var priv *btcec.PrivateKey
	if generate {
		key, err := ecdsa.GenerateKey(curve, rand.Reader)
		if err != nil {
			return "", err
		}
		priv = &btcec.PrivateKey{
			PublicKey: key.PublicKey,
			D:         key.D,
		}
	} else {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter Private Key: ")
		prvKey, _ := reader.ReadString('\n')
		prvKey = strings.TrimSpace(prvKey)
		privWif, err := btcutil.DecodeWIF(prvKey)
		if err != nil {
			return "", err
		}
		priv = privWif.PrivKey
	}
	pub := priv.PubKey()

	var buf bytes.Buffer

	writeKeyData(&buf, priv, pub, !uncompressed, verbose, showPrivateKey)

	return buf.String(), nil
}

func bytesToString(bytes []byte) (str string) {
	for i, b := range bytes {
		if i%8 == 0 {
			str += newLine
		}
		str += fmt.Sprintf(" 0x%02X,", b)
	}
	return str
}

func writeKeyData(buf *bytes.Buffer, priv *btcec.PrivateKey, pub *btcec.PublicKey, compressed, verbose, showPrivateKey bool) error {
	var serializedPK []byte
	if compressed {
		buf.WriteString(newLine)
		buf.WriteString("[COMPRESSED]")
		serializedPK = pub.SerializeCompressed()
	} else {
		buf.WriteString("[UNCOMPRESSED]")
		serializedPK = pub.SerializeUncompressed()
	}

	hash := btcutil.Hash160(serializedPK)

	addr, err := btcutil.NewAddressPubKeyHash(hash, &params)
	if err != nil {
		return err
	}

	signature, err := btcec.SignCompact(curve, priv, hash, false)
	if err != nil {
		return err
	}

	pk, _, err := btcec.RecoverCompact(curve, signature, hash)
	if err != nil {
		return err
	}

	if !pk.IsEqual(pub) {
		return errors.New("Sum Ting Wong")
	}

	buf.WriteString(newLine)
	buf.WriteString("Address: ")
	buf.WriteString(addr.EncodeAddress())
	buf.WriteString(newLine)
	if verbose {
		buf.WriteString("Hash:")
		buf.WriteString(bytesToString(hash))
		buf.WriteString(newLine)
		buf.WriteString("Serialized PK:")
		buf.WriteString(bytesToString(serializedPK))
		buf.WriteString(newLine)
	}

	privWif, err := btcutil.NewWIF(priv, &params, compressed)
	if err != nil {
		return err
	}

	if showPrivateKey {
		buf.WriteString("Private key: ")
		buf.WriteString(privWif.String())
		buf.WriteString(newLine)
	} else {
		writeNewFile(addr.EncodeAddress(), privWif.String(), 0600)
	}

	return nil
}

// deriveCoinTypeKey derives the cointype key which can be used to derive the
// extended key for an account according to the hierarchy described by BIP0044
// given the coin type key.
//
// In particular this is the hierarchical deterministic extended key path:
// m/44'/<coin type>'
func deriveCoinTypeKey(masterNode *hdkeychain.ExtendedKey,
	coinType uint32) (*hdkeychain.ExtendedKey, error) {
	// Enforce maximum coin type.
	if coinType > maxCoinType {
		return nil, fmt.Errorf("bad coin type")
	}

	// The hierarchy described by BIP0043 is:
	//  m/<purpose>'/*
	// This is further extended by BIP0044 to:
	//  m/44'/<coin type>'/<account>'/<branch>/<address index>
	//
	// The branch is 0 for external addresses and 1 for internal addresses.

	// Derive the purpose key as a child of the master node.
	purpose, err := masterNode.Child(44 + hdkeychain.HardenedKeyStart)
	if err != nil {
		return nil, err
	}

	// Derive the coin type key as a child of the purpose key.
	coinTypeKey, err := purpose.Child(coinType + hdkeychain.HardenedKeyStart)
	if err != nil {
		return nil, err
	}

	return coinTypeKey, nil
}

// deriveAccountKey derives the extended key for an account according to the
// hierarchy described by BIP0044 given the master node.
//
// In particular this is the hierarchical deterministic extended key path:
//   m/44'/<coin type>'/<account>'
func deriveAccountKey(coinTypeKey *hdkeychain.ExtendedKey,
	account uint32) (*hdkeychain.ExtendedKey, error) {
	// Enforce maximum account number.
	if account > MaxAccountNum {
		return nil, fmt.Errorf("account num too high")
	}

	// Derive the account key as a child of the coin type key.
	return coinTypeKey.Child(account + hdkeychain.HardenedKeyStart)
}

// checkBranchKeys ensures deriving the extended keys for the internal and
// external branches given an account key does not result in an invalid child
// error which means the chosen seed is not usable.  This conforms to the
// hierarchy described by BIP0044 so long as the account key is already derived
// accordingly.
//
// In particular this is the hierarchical deterministic extended key path:
//   m/44'/<coin type>'/<account>'/<branch>
//
// The branch is 0 for external addresses and 1 for internal addresses.
func checkBranchKeys(acctKey *hdkeychain.ExtendedKey) error {
	// Derive the external branch as the first child of the account key.
	if _, err := acctKey.Child(ExternalBranch); err != nil {
		return err
	}

	// Derive the external branch as the second child of the account key.
	_, err := acctKey.Child(InternalBranch)
	return err
}

// generateSeed derives an address from an HDKeychain for use in wallet. It
// outputs the seed, address, and extended public key to the file specified.
func generateSeed() (string, error) {
	seed, err := hdkeychain.GenerateSeed(hdkeychain.RecommendedSeedLen)
	if err != nil {
		return "", err
	}

	// Derive the master extended key from the seed.
	root, err := hdkeychain.NewMaster(seed, &params)
	if err != nil {
		return "", err
	}
	defer root.Zero()

	// Derive the cointype key according to BIP0044.
	coinTypeKeyPriv, err := deriveCoinTypeKey(root, params.HDCoinType)
	if err != nil {
		return "", err
	}
	defer coinTypeKeyPriv.Zero()

	// Derive the account key for the first account according to BIP0044.
	acctKeyPriv, err := deriveAccountKey(coinTypeKeyPriv, 0)
	if err != nil {
		// The seed is unusable if the any of the children in the
		// required hierarchy can't be derived due to invalid child.
		if err == hdkeychain.ErrInvalidChild {
			return "", fmt.Errorf("the provided seed is unusable")
		}

		return "", err
	}

	// Ensure the branch keys can be derived for the provided seed according
	// to BIP0044.
	if err := checkBranchKeys(acctKeyPriv); err != nil {
		// The seed is unusable if the any of the children in the
		// required hierarchy can't be derived due to invalid child.
		if err == hdkeychain.ErrInvalidChild {
			return "", fmt.Errorf("the provided seed is unusable")
		}

		return "", err
	}

	// The address manager needs the public extended key for the account.
	acctKeyPub, err := acctKeyPriv.Neuter()
	if err != nil {
		return "", fmt.Errorf("failed to convert private key for account 0")
	}

	index := uint32(0)  // First address
	branch := uint32(0) // External

	// The next address can only be generated for accounts that have already
	// been created.
	acctKey := acctKeyPub
	defer acctKey.Zero()

	// Derive the appropriate branch key and ensure it is zeroed when done.
	branchKey, err := acctKey.Child(branch)
	if err != nil {
		return "", err
	}
	defer branchKey.Zero() // Ensure branch key is zeroed when done.

	key, err := branchKey.Child(index)
	if err != nil {
		return "", err
	}
	defer key.Zero()

	addr, err := key.Address(&params)
	if err != nil {
		return "", err
	}

	// Require the user to write down the seed.
	seedStr := walletseed.EncodeMnemonic(seed)
	seedStrSplit := strings.Split(seedStr, " ")
	fmt.Println("WRITE DOWN THE SEED GIVEN BELOW. YOU WILL NOT BE GIVEN " +
		"ANOTHER CHANCE TO.\n")
	fmt.Printf("Your wallet generation seed is:\n\n")
	for i := 0; i < hdkeychain.RecommendedSeedLen+1; i++ {
		fmt.Printf("%v ", seedStrSplit[i])

		if (i+1)%6 == 0 {
			fmt.Printf("\n")
		}
	}

	fmt.Printf("\n\nHex: %x\n", seed)
	fmt.Println("\nIMPORTANT: Keep the seed in a safe place as you\n" +
		"will NOT be able to restore your wallet without it.")
	fmt.Println("Please keep in mind that anyone who has access\n" +
		"to the seed can also restore your wallet thereby\n" +
		"giving them access to all your funds, so it is\n" +
		"imperative that you keep it in a secure location.\n")

	var buf bytes.Buffer
	buf.WriteString("First address: ")
	buf.WriteString(addr.EncodeAddress())
	buf.WriteString(newLine)
	buf.WriteString("Extended public key: ")
	buf.WriteString(acctKey.String())
	buf.WriteString(newLine)

	// Zero the seed array.
	copy(seed[:], bytes.Repeat([]byte{0x00}, 32))

	return buf.String(), nil
}

// promptSeed is used to prompt for the wallet seed which maybe required during
// upgrades.
func promptSeed(seedA *[32]byte) error {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter existing wallet seed: ")
		seedStr, err := reader.ReadString('\n')
		if err != nil {
			return err
		}

		seedStrTrimmed := strings.TrimSpace(seedStr)

		seed, err := walletseed.DecodeUserInput(seedStrTrimmed)
		if err != nil || len(seed) < hdkeychain.MinSeedBytes ||
			len(seed) > hdkeychain.MaxSeedBytes {
			if err != nil {
				fmt.Printf("Input error: %v\n", err.Error())
			}

			fmt.Printf("Invalid seed specified.  Must be "+
				"the words of the seed and at least %d bits and "+
				"at most %d bits\n", hdkeychain.MinSeedBytes*8,
				hdkeychain.MaxSeedBytes*8)
			continue
		}

		copy(seedA[:], seed[:])

		// Zero the seed slice.
		copy(seed[:], bytes.Repeat([]byte{0x00}, 32))
		return nil
	}
}

func verifySeed() error {
	seed := new([32]byte)
	err := promptSeed(seed)
	if err != nil {
		return err
	}

	// Derive the master extended key from the seed.
	root, err := hdkeychain.NewMaster(seed[:], &params)
	if err != nil {
		return err
	}
	defer root.Zero()

	// Derive the cointype key according to BIP0044.
	coinTypeKeyPriv, err := deriveCoinTypeKey(root, params.HDCoinType)
	if err != nil {
		return err
	}
	defer coinTypeKeyPriv.Zero()

	// Derive the account key for the first account according to BIP0044.
	acctKeyPriv, err := deriveAccountKey(coinTypeKeyPriv, 0)
	if err != nil {
		// The seed is unusable if the any of the children in the
		// required hierarchy can't be derived due to invalid child.
		if err == hdkeychain.ErrInvalidChild {
			return fmt.Errorf("the provided seed is unusable")
		}

		return err
	}

	// Ensure the branch keys can be derived for the provided seed according
	// to BIP0044.
	if err := checkBranchKeys(acctKeyPriv); err != nil {
		// The seed is unusable if the any of the children in the
		// required hierarchy can't be derived due to invalid child.
		if err == hdkeychain.ErrInvalidChild {
			return fmt.Errorf("the provided seed is unusable")
		}

		return err
	}

	// The address manager needs the public extended key for the account.
	acctKeyPub, err := acctKeyPriv.Neuter()
	if err != nil {
		return fmt.Errorf("failed to convert private key for account 0")
	}

	index := uint32(0)  // First address
	branch := uint32(0) // External

	// The next address can only be generated for accounts that have already
	// been created.
	acctKey := acctKeyPub
	defer acctKey.Zero()

	// Derive the appropriate branch key and ensure it is zeroed when done.
	branchKey, err := acctKey.Child(branch)
	if err != nil {
		return err
	}
	defer branchKey.Zero() // Ensure branch key is zeroed when done.

	key, err := branchKey.Child(index)
	if err != nil {
		return err
	}
	defer key.Zero()

	addr, err := key.Address(&params)
	if err != nil {
		return err
	}

	fmt.Printf("First derived address of given seed: \n%v\n",
		addr.EncodeAddress())

	// Zero the seed array.
	copy(seed[:], bytes.Repeat([]byte{0x00}, 32))

	return nil
}

func main() {
	if runtime.GOOS == "windows" {
		newLine = "\r\n"
	}
	helpMessage := func() {
		fmt.Println(
			"Usage: dcraddrgen [-testnet] [-simnet] [-noseed] [-verify] [-h] [-v] filename")
		fmt.Println("Generate a Endurio private and public key or wallet seed. \n" +
			"These are output to the file 'filename'.\n")
		fmt.Println("  -h \t\tPrint this message")
		fmt.Println("  -testnet \tGenerate a testnet key instead of mainnet")
		fmt.Println("  -simnet \tGenerate a simnet key instead of mainnet")
		fmt.Println("  -seed \tGenerate a seed instead of a single keypair")
		fmt.Println("  -verify \tVerify a seed by generating the first address")
		fmt.Println("  -key \tImport a private key")
		fmt.Println("  -v \tPrint dev data")
		fmt.Println("  -u \tPrint uncompressed keys instead")
		fmt.Println("  -p \tPrint private key to console")
	}

	setupFlags(helpMessage, flag.CommandLine)
	flag.Parse()

	if *getHelp {
		helpMessage()
		return
	}

	if *verify {
		err := verifySeed()
		if err != nil {
			fmt.Printf("Error verifying seed: %v\n", err.Error())
			return
		}
		return
	}

	fn := flag.Arg(0)

	// Alter the globals to specified network.
	if *testnet {
		if *simnet {
			fmt.Println("Error: Only specify one network.")
			return
		}
		params = chaincfg.TestNet3Params
	}
	if *simnet {
		params = chaincfg.SimNetParams
	}

	// Single keypair generation.
	if !*seed {
		str, err := generateKeyPair(!*key, *verbose, *uncompressed, *showPrivateKey)
		if err != nil {
			fmt.Printf("Error generating key pair: %v\n", err.Error())
			return
		}
		if len(fn) > 0 {
			writeNewFile(fn, str, 0600)
		} else {
			fmt.Print(str)
		}
		return
	}

	// Derivation of an address from an HDKeychain for use in wallet.
	str, err := generateSeed()
	if err != nil {
		fmt.Printf("Error generating seed: %v\n", err.Error())
		return
	}
	if len(fn) > 0 {
		writeNewFile(fn, str, 0600)
	} else {
		fmt.Print(str)
	}
}
