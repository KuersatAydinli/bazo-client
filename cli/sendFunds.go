package cli

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"github.com/bazo-blockchain/bazo-client/network"
	"github.com/bazo-blockchain/bazo-client/util"
	"github.com/bazo-blockchain/bazo-miner/crypto"
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/urfave/cli"
	"log"
)

type sendFundsArgs struct {
	header			int
	fromFile		string
	toFile			string
	toAddress		string
	multisigFile	string
	amount			int
	fee				int
	txcnt			int
}

func AddSendFundsCommand(app *cli.App, logger *log.Logger) {
	command := cli.Command {
		Name:	"sendFunds",
		Usage:	"send funds from one account to another",
		Action:	func(c *cli.Context) error {
			args := &sendFundsArgs {
				header: 		c.Int("header"),
				fromFile: 		c.String("from"),
				toAddress: 		c.String("toaddress"),
				toFile: 		c.String("tofile"),
				multisigFile: 	c.String("multisig"),
				amount: 		c.Int("amount"),
				fee: 			c.Int("fee"),
				txcnt: 			c.Int("txcnt"),
			}

			err := args.ValidateInput()
			if err != nil {
				return err
			}

			return sendFunds(args, logger)
		},
		Flags:	[]cli.Flag {
			cli.IntFlag {
				Name: 	"header",
				Usage: 	"header flag",
				Value:	0,
			},
			cli.StringFlag {
				Name: 	"from",
				Usage: 	"load the sender's private key from `FILE`",
			},
			cli.StringFlag {
				Name: 	"toaddress",
				Usage: 	"the recipient's 128 byze public address",
			},
			cli.StringFlag {
				Name: 	"tofile",
				Usage: 	"load the recipient's public key from `FILE`",
			},
			cli.IntFlag {
				Name: 	"amount",
				Usage:	"specify the amount to send",
			},
			cli.IntFlag {
				Name: 	"fee",
				Usage:	"specify the fee",
			},
			cli.IntFlag {
				Name: 	"txcount",
				Usage:	"the sender's current transaction counter",
			},
			cli.StringFlag {
				Name: 	"multisig",
				Usage: 	"load multi-signature server’s public key from `FILE`",
			},
		},
	}

	app.Commands = append(app.Commands, command)
}

func sendFunds(args *sendFundsArgs, logger *log.Logger) error {
	fromPrivKey, err := crypto.ExtractECDSAKeyFromFile(args.fromFile)
	if err != nil {
		return err
	}

	var toPubKey *ecdsa.PublicKey
	if len(args.toFile) == 0 {
		if len(args.toAddress) == 0 {
			return errors.New(fmt.Sprintln("No recipient specified"))
		} else {
			if len(args.toAddress) != 128 {
				return errors.New(fmt.Sprintln("Invalid recipient address"))
			}

			runes := []rune(args.toAddress)
			pub1 := string(runes[:64])
			pub2 := string(runes[64:])

			toPubKey, err = crypto.GetPubKeyFromString(pub1, pub2)
			if err != nil {
				return err
			}
		}
	} else {
		toPubKey, err = crypto.ExtractECDSAPublicKeyFromFile(args.toFile)
		if err != nil {
			return err
		}
	}

	var multisigPrivKey *ecdsa.PrivateKey
	if len(args.multisigFile) > 0 {
		multisigPrivKey, err = crypto.ExtractECDSAKeyFromFile(args.multisigFile)
		if err != nil {
			return err
		}
	}

	fromAddress := crypto.GetAddressFromPubKey(&fromPrivKey.PublicKey)
	toAddress := crypto.GetAddressFromPubKey(toPubKey)

	tx, err := protocol.ConstrFundsTx(
		byte(args.header),
		uint64(args.amount),
		uint64(args.fee),
		uint32(args.txcnt),
		protocol.SerializeHashContent(fromAddress),
		protocol.SerializeHashContent(toAddress),
		fromPrivKey,
		multisigPrivKey,
		nil)

	if err != nil {
		logger.Printf("%v\n", err)
		return err
	}

	if err := network.SendTx(util.Config.BootstrapIpport, tx, p2p.FUNDSTX_BRDCST); err != nil {
		logger.Printf("%v\n", err)
		return err
	} else {
		logger.Printf("Transaction successfully sent to network:\nTxHash: %x%v", tx.Hash(), tx)
	}

	return nil
}

func (args sendFundsArgs) ValidateInput() error {
	if len(args.fromFile) == 0 {
		return errors.New("argument missing: fromFile")
	}

	if len(args.toFile) == 0 && len(args.toAddress) == 0 {
		return errors.New("argument missing: toFile or toAddess")
	}

	if len(args.toFile) == 0 && len(args.toAddress) != 128 {
		return errors.New("invalid argument: toAddress")
	}

	if len(args.multisigFile) == 0 {
		return errors.New("argument missing: multisigFile")
	}

	if args.fee <= 0 {
		return errors.New("invalid argument: fee must be > 0")
	}

	if args.txcnt < 0 {
		return errors.New("invalid argument: txcnt must be >= 0")
	}

	if args.amount <= 0 {
		return errors.New("invalid argument: amount must be > 0")
	}

	return nil
}
