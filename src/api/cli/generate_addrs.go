package cli

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	gcli "github.com/urfave/cli"

	"github.com/skycoin/skycoin/src/cipher"
	"github.com/skycoin/skycoin/src/wallet"
)

func generateAddrsCmd(cfg Config) gcli.Command {
	name := "generateAddresses"
	return gcli.Command{
		Name:      name,
		Usage:     "Generate additional addresses for a wallet",
		ArgsUsage: " ",
		Description: fmt.Sprintf(`The default wallet (%s) will
		be used if no wallet and address was specified.

		Use caution when using the "-p" command. If you have command
		history enabled your wallet encryption password can be recovered from the
		history log. If you do not include the "-p" option you will be prompted to
		enter your password after you enter your command.`, cfg.FullWalletPath()),
		Flags: []gcli.Flag{
			gcli.UintFlag{
				Name:  "n",
				Value: 1,
				Usage: `[numberOfAddresses]	Number of addresses to generate`,
			},
			gcli.StringFlag{
				Name:  "f",
				Value: cfg.FullWalletPath(),
				Usage: `[wallet file or path] Generate addresses in the wallet`,
			},
			gcli.BoolFlag{
				Name:  "json,j",
				Usage: "Returns the results in JSON format",
			},
		},
		OnUsageError: onCommandUsageError(name),
		Action:       generateAddrs,
	}
	// Commands = append(Commands, cmd)
}

func generateAddrs(c *gcli.Context) error {
	cfg := ConfigFromContext(c)

	// get number of address that are need to be generated.
	num := c.Uint64("n")
	if num == 0 {
		return errors.New("-n must > 0")
	}

	jsonFmt := c.Bool("json")

	w, err := resolveWalletPath(cfg, c.String("f"))
	if err != nil {
		return err
	}

	addrs, err := GenerateAddressesInFile(w, num)

	switch err.(type) {
	case nil:
	case WalletLoadError:
		errorWithHelp(c, err)
		return nil
	case WalletSaveError:
		return errors.New("save wallet failed")
	default:
		return err
	}

	if jsonFmt {
		s, err := FormatAddressesAsJSON(addrs)
		if err != nil {
			return err
		}
		fmt.Println(s)
	} else {
		fmt.Println(FormatAddressesAsJoinedArray(addrs))
	}

	return nil
}

// GenerateAddressesInFile generates addresses in given wallet file
func GenerateAddressesInFile(walletFile string, num uint64) ([]cipher.Address, error) {
	wlt, err := wallet.Load(walletFile)
	if err != nil {
		return nil, WalletLoadError(err)
	}

	addrs, err := wlt.GenerateAddresses(num)
	if err != nil {
		return nil, err
	}

	dir, err := filepath.Abs(filepath.Dir(walletFile))
	if err != nil {
		return nil, err
	}

	if err := wlt.Save(dir); err != nil {
		return nil, WalletSaveError(err)
	}

	return addrs, nil
}

// FormatAddressesAsJSON converts []cipher.Address to strings and formats the array into a standard JSON object wrapper
func FormatAddressesAsJSON(addrs []cipher.Address) (string, error) {
	d, err := formatJSON(struct {
		Addresses []string `json:"addresses"`
	}{
		Addresses: AddressesToStrings(addrs),
	})

	if err != nil {
		return "", err
	}

	return string(d), nil
}

// FormatAddressesAsJoinedArray converts []cipher.Address to strings and concatenates them with a comma
func FormatAddressesAsJoinedArray(addrs []cipher.Address) string {
	return strings.Join(AddressesToStrings(addrs), ",")
}

// AddressesToStrings converts []cipher.Address to []string
func AddressesToStrings(addrs []cipher.Address) []string {
	if addrs == nil {
		return nil
	}

	addrsStr := make([]string, len(addrs))
	for i, a := range addrs {
		addrsStr[i] = a.String()
	}

	return addrsStr
}
