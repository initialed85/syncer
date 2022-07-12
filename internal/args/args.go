package args

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Args struct {
	Send       bool
	Receive    bool
	LocalPath  string
	RemotePath string
	RemoteHost string
	Rate       time.Duration
	Debounce   time.Duration
}

func ParseArgs() Args {
	args := Args{}

	flag.BoolVar(&args.Send, "send", false, "Should this node send")
	flag.BoolVar(&args.Receive, "receive", false, "Should this node receive")
	flag.StringVar(&args.LocalPath, "localPath", "", "Local path to sync")
	flag.StringVar(&args.RemotePath, "remotePath", "", "Remote path to sync")
	flag.StringVar(&args.RemoteHost, "remoteHost", "", "Remote host to sync with")

	flag.DurationVar(&args.Rate, "rate", time.Millisecond*100, "Rate to update at")
	flag.DurationVar(&args.Debounce, "debounce", time.Millisecond*2000, "Duration to wait for filesystem to settle")

	flag.Parse()

	return args
}

func ValidateArgs(args Args) Args {
	var err error

	if !(args.Send || args.Receive) {
		log.Fatal("-send xor -receive must be set")
	}

	if args.Send && args.Receive {
		log.Fatal("-send and -receive cannot be set")
	}

	args.LocalPath = strings.TrimSpace(args.LocalPath)
	if args.LocalPath == "" {
		log.Fatal("-localPath must be set")
	}

	args.LocalPath, err = filepath.Abs(args.LocalPath)
	if err != nil {
		log.Fatalf("-localPath %#+v could not be converted to an absolute path (stating %v)", args.LocalPath, err)
	}

	stat, err := os.Stat(args.LocalPath)
	if err != nil {
		log.Fatalf("-localPath %#+v failed os.Stat (stating %v)", args.LocalPath, err)
	}

	if !stat.IsDir() {
		log.Fatalf("-localPath %#+v is not a directory", args.LocalPath)
	}

	args.RemotePath = strings.TrimSpace(args.RemotePath)
	if args.RemotePath == "" {
		log.Fatal("-remotePath must be set")
	}

	args.RemoteHost = strings.TrimSpace(args.RemoteHost)
	if args.RemotePath == "" {
		log.Fatal("-remoteHost must be set")
	}

	if args.Rate < time.Duration(0) {
		log.Fatal("-rate cannot be negative")
	}

	if args.Debounce < time.Duration(0) {
		log.Fatal("-debounce cannot be negative")
	}

	if args.Debounce <= args.Rate {
		log.Printf("warning: debounce <= rate; every update will result in handling")
	}

	return args
}
