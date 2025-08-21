package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/soatok/freon/client/internal"
)

// Entrypoint for the command line program
func main() {
	flag.Usage = func() { fmt.Fprintf(os.Stderr, "%s\n", usage) }

	if len(os.Args) == 1 {
		flag.Usage()
		os.Exit(1)
	}

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	// This is where commands are processed.
	// Note that the first verb after `freon` is case insensitive.
	command := strings.ToLower(args[0])
	subArgs := args[1:]
	switch command {
	case "keygen":
		if len(subArgs) == 0 {
			fmt.Fprintf(os.Stderr, "Error: keygen requires a subcommand\n\n")
			fmt.Fprintf(os.Stderr, "%s\n", keygenUsage)
			os.Exit(1)
		}

		subcommand := subArgs[0]
		switch subcommand {
		case "create":
			FreonKeygenCreate(subArgs[1:])
		case "join":
			FreonKeygenJoin(subArgs[1:])
		case "list":
			FreonKeygenList(subArgs[1:])
		default:
			fmt.Fprintf(os.Stderr, "Error: unknown keygen subcommand: %s\n\n", subcommand)
			fmt.Fprintf(os.Stderr, "%s\n", keygenUsage)
			os.Exit(1)
		}

	case "sign":
		if len(subArgs) == 0 {
			fmt.Fprintf(os.Stderr, "Error: sign requires a subcommand\n\n")
			fmt.Fprintf(os.Stderr, "%s\n", signUsage)
			os.Exit(1)
		}

		subcommand := subArgs[0]
		switch subcommand {
		case "create":
			FreonSignCreate(subArgs[1:])
		case "join":
			FreonSignJoin(subArgs[1:])
		default:
			fmt.Fprintf(os.Stderr, "Error: unknown sign subcommand: %s\n\n", subcommand)
			fmt.Fprintf(os.Stderr, "%s\n", signUsage)
			os.Exit(1)
		}

	case "terminate":
		FreonTerminate(subArgs)

	case "help":
		if len(subArgs) == 0 {
			flag.Usage()
		} else {
			// Handle help for specific commands
			switch subArgs[0] {
			case "keygen":
				fmt.Fprintf(os.Stderr, "%s\n", keygenUsage)
			case "sign":
				fmt.Fprintf(os.Stderr, "%s\n", signUsage)
			case "terminate":
				fmt.Fprintf(os.Stderr, "%s\n", terminateUsage)
			default:
				fmt.Fprintf(os.Stderr, "No help available for: %s\n", subArgs[0])
				os.Exit(1)
			}
		}

	default:
		fmt.Fprintf(os.Stderr, "Error: unknown command: %s\n\n", command)
		flag.Usage()
		os.Exit(1)
	}
}

// Handle file inputs for shell scripting arguments.
//
// If filename is non-empty, read that file.
// If filename is empty, attempt to read STDIN.
func readInput(filename string) ([]byte, error) {
	if filename != "" {
		// Read from file
		return os.ReadFile(filename)
	}

	// No filename: check if STDIN has data
	stat, err := os.Stdin.Stat()
	if err != nil {
		return nil, err
	}

	if (stat.Mode() & os.ModeCharDevice) != 0 {
		// STDIN is a terminal, no data
		return nil, fmt.Errorf("no input provided")
	}

	// Read from STDIN
	return io.ReadAll(os.Stdin)
}

// CMD: `freon keygen create ...`
func FreonKeygenCreate(args []string) {
	fs := flag.NewFlagSet("keygen create", flag.ExitOnError)
	fs.Usage = func() { fmt.Fprintf(os.Stderr, "%s\n", keygenCreateUsage) }
	host := fs.String("h", "", "Coordinator hostname:port")
	hostLong := fs.String("host", "", "Coordinator hostname:port")
	participants := fs.Int("n", 0, "Number of participants")
	participantsLong := fs.Int("participants", 0, "Number of participants")
	threshold := fs.Int("t", 0, "Minimum shares required for signing")
	thresholdLong := fs.Int("threshold", 0, "Minimum shares required for signing")
	recipient := fs.String("r", "", "Age/SSH public key to encrypt share")
	recipientLong := fs.String("recipient", "", "Age/SSH public key to encrypt share")
	fs.Parse(args)

	// Merge short/long flags
	if *hostLong != "" {
		*host = *hostLong
	}
	if *participantsLong != 0 {
		*participants = *participantsLong
	}
	if *thresholdLong != 0 {
		*threshold = *thresholdLong
	}
	if *recipientLong != "" {
		*recipient = *recipientLong
	}

	// Validate required flags
	if *host == "" {
		fmt.Fprintf(os.Stderr, "Error: -h/--host is required\n")
		fs.Usage()
		os.Exit(1)
	}
	if *participants == 0 {
		fmt.Fprintf(os.Stderr, "Error: -n/--participants is required\n")
		fs.Usage()
		os.Exit(1)
	}
	if *threshold == 0 {
		fmt.Fprintf(os.Stderr, "Error: -t/--threshold is required\n")
		fs.Usage()
		os.Exit(1)
	}
	if *participants < 2 || *participants > 255 {
		fmt.Fprintf(os.Stderr, "Error: participants must be between 2 and 255\n")
		os.Exit(1)
	}
	if *threshold > *participants {
		fmt.Fprintf(os.Stderr, "Error: threshold cannot exceed participants\n")
		os.Exit(1)
	}

	// Now that we have a configuration, let's initialize the ceremony
	internal.InitKeyGenCeremony(*host, uint16(*participants), uint16(*threshold))
}

// CMD: `freon keygen join ...`
func FreonKeygenJoin(args []string) {
	fs := flag.NewFlagSet("keygen join", flag.ExitOnError)
	fs.Usage = func() { fmt.Fprintf(os.Stderr, "%s\n", keygenJoinUsage) }
	host := fs.String("h", "", "Coordinator hostname:port")
	hostLong := fs.String("host", "", "Coordinator hostname:port")
	groupID := fs.String("g", "", "Group ID from ceremony creator")
	groupIDLong := fs.String("group", "", "Group ID from ceremony creator")
	recipient := fs.String("r", "", "Age/SSH public key to encrypt share")
	recipientLong := fs.String("recipient", "", "Age/SSH public key to encrypt share")
	fs.Parse(args)

	// Merge short/long flags
	if *hostLong != "" {
		*host = *hostLong
	}
	if *groupIDLong != "" {
		*groupID = *groupIDLong
	}
	if *recipientLong != "" {
		*recipient = *recipientLong
	}

	// Data validation
	if *host == "" {
		fmt.Fprintf(os.Stderr, "Error: -h/--host is required\n")
		fs.Usage()
		os.Exit(1)
	}
	if *groupID == "" {
		fmt.Fprintf(os.Stderr, "Error: -g/--group is required\n")
		fs.Usage()
		os.Exit(1)
	}
	if *recipient == "" {
		fmt.Fprintf(os.Stderr, "Error: -r/--recipient is required\n")
		fs.Usage()
		os.Exit(1)
	}

	internal.JoinKeyGenCeremony(*host, *groupID, *recipient)
}

// CMD: `freon keygen list ...`
func FreonKeygenList(args []string) {
	internal.ListKeyGen()
}

// CMD: `freon sign create ...`
func FreonSignCreate(args []string) {
	// Parse CLI arguments:
	fs := flag.NewFlagSet("sign create", flag.ExitOnError)
	fs.Usage = func() { fmt.Fprintf(os.Stderr, "%s\n", signCreateUsage) }
	groupID := fs.String("g", "", "Group ID from DKG ceremony")
	groupIDLong := fs.String("group", "", "Group ID from DKG ceremony")
	host := fs.String("h", "", "Coordinator hostname:port")
	hostLong := fs.String("host", "", "Coordinator hostname:port")
	openssh := fs.Bool("openssh", false, "Return OpenSSH-compatible signature format")
	namespace := fs.String("namespace", "", `Specify a namespace for OpenSSH (default: "file")`)
	fs.Parse(args)

	// Merge short/long flags
	if *groupIDLong != "" {
		*groupID = *groupIDLong
	}
	if *hostLong != "" {
		*host = *hostLong
	}

	// Data validation
	if *groupID == "" {
		fmt.Fprintf(os.Stderr, "Error: -g/--group is required\n")
		fs.Usage()
		os.Exit(1)
	}
	if *openssh {
		// Default to "file"
		if *namespace == "" {
			*namespace = "file"
		}
	} else if *namespace != "" {
		fmt.Printf("--namespace can only be used with --openssh")
		fs.Usage()
		os.Exit(1)
	}

	// Get message file from remaining args
	remainingArgs := fs.Args()
	var messageFile string = ""
	if len(remainingArgs) > 0 {
		messageFile = remainingArgs[0]
	}
	message, err := readInput(messageFile)
	if err != nil {
		fmt.Printf("A message file is required")
		fs.Usage()
		os.Exit(1)
	}
	internal.InitSignCeremony(*groupID, *host, message, *openssh, *namespace)
}

// CMD: `freon sign join ...`
func FreonSignJoin(args []string) {
	fs := flag.NewFlagSet("sign join", flag.ExitOnError)
	fs.Usage = func() { fmt.Fprintf(os.Stderr, "%s\n", signJoinUsage) }
	ceremonyID := fs.String("c", "", "Ceremony ID")
	ceremonyIDLong := fs.String("ceremony", "", "Ceremony ID")
	host := fs.String("h", "", "Coordinator hostname:port")
	hostLong := fs.String("host", "", "Coordinator hostname:port")
	identity := fs.String("i", "", "Path to age secret keys file")
	identityLong := fs.String("identity", "", "Path to age secret keys file")
	// autoConfirm := fs.Bool("auto-confirm", false, "Skip message confirmation prompt")
	fs.Parse(args)

	if *ceremonyIDLong != "" {
		*ceremonyID = *ceremonyIDLong
	}
	if *hostLong != "" {
		*host = *hostLong
	}
	if *identityLong != "" {
		*identity = *identityLong
	}
	remainingArgs := fs.Args()
	var messageFile string = ""
	if len(remainingArgs) > 0 {
		messageFile = remainingArgs[0]
	}
	message, err := readInput(messageFile)
	if err != nil {
		fmt.Printf("A message file is required")
		fs.Usage()
		os.Exit(1)
	}
	internal.JoinSignCeremony(*ceremonyID, *host, *identity, message)
	// , *autoConfirm)
}

// CMD: `freon sign terminate ...`
func FreonTerminate(args []string) {
	fs := flag.NewFlagSet("sign join", flag.ExitOnError)
	fs.Usage = func() { fmt.Fprintf(os.Stderr, "%s\n", terminateUsage) }
	ceremonyID := fs.String("c", "", "Ceremony ID")
	ceremonyIDLong := fs.String("ceremony", "", "Ceremony ID")
	fs.Parse(args)
	if *ceremonyIDLong != "" {
		*ceremonyID = *ceremonyIDLong
	}
	internal.TerminateSignCeremony(*ceremonyID)
}
