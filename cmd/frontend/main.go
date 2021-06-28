package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

type Config struct {
	VIPs              []string `default:"20.0.0.1/32" desc:"Virtual IP address" envconfig:"VIPS"`
	Gateways          []string `default:"169.254.100.254,fe80::beef" desc:"Next-hop addresses of external gateways"`
	VRRPs             []string `default:"" "desc:"VRRP IP addresses to be used as next-hops for static default routes" envconfig:"VRRPS"`
	ExternalInterface string   `default:"ext-vlan" desc:"External interface to start BIRD on" split_words:"true"`
	BirdConfigPath    string   `default:"/etc/bird" desc:"Path to place bird config files" split_words:"true"`
	LocalAS           string   `default:"8103" desc:"Local BGP AS number" envconfig:"LOCAL_AS"`
	RemoteAS          string   `default:"4248829953" desc:"Local BGP AS number" envconfig:"REMOTE_AS"`
	BGPLocalPort      string   `default:"10179" desc:"Local BGP server port" envconfig:"BGP_LOCAL_PORT"`
	BGPRemotePort     string   `default:"10179" desc:"Remote BGP server port" envconfig:"BGP_REMOTE_PORT"`
	BGPHoldTime       string   `default:"3" desc:"Seconds to wait for a Keepalive message from peer before considering the connection stale" envconfig:"BGP_HOLD_TIME"`
	TableID           int      `default:"4096" desc:"OS Kernel routing table ID BIRD syncs the routes with" envconfig:"TABLE_ID"`
	BFD               bool     `default:"false" desc:"Enable BFD for BGP" envconfig:"BGP"`
	ECMP              bool     `default:"false" desc:"Enable ECMP towards next-hops of avaialble gateways" envconfig:"ECMP"`
	DropIfNoPeer      bool     `default:"false" desc:"Install default blackhole route with high metric into routing table TableID" split_words:"true"`
	LogBird           bool     `default:"false" desc:"Add important bird log snippets to our log" split_words:"true"`
}

func main() {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGHUP,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer cancel()

	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&nested.Formatter{})

	config := &Config{}
	if err := envconfig.Usage("nfe", config); err != nil {
		logrus.Fatal(err)
	}
	if err := envconfig.Process("nfe", config); err != nil {
		logrus.Fatalf("%v", err)
	}
	logrus.Infof("rootConf: %+v", config)

	fe := NewFrontEndService(config)
	defer fe.RemoveVipRules()

	fe.WriteConfig(ctx, cancel)
	if err := fe.AddVipRules(cancel); err != nil {
		logrus.Fatalf("Failed to setup src routes for VIPs: %v", err)
	}

	feErrCh := fe.Start(ctx)
	exitOnErrCh(ctx, cancel, feErrCh)

	/* time.Sleep(1 * time.Second)
	if err := fe.VerifyConfig(ctx); err != nil {
		logrus.Errorf("Failed to verify config")
	} */

	fe.Monitor(ctx)

	logrus.Infof("FE running")
	<-ctx.Done()
	logrus.Warnf("FE shutting down")
}

func exitOnErrCh(ctx context.Context, cancel context.CancelFunc, errCh <-chan error) {
	// If we already have an error, log it and exit
	select {
	case err, ok := <-errCh:
		if ok {
			logrus.Errorf("exitOnErrCh(0): %v", err)
		}
	default:
	}
	// Otherwise wait for an error in the background to log and cancel
	go func(ctx context.Context, errCh <-chan error) {
		if err, ok := <-errCh; ok {
			logrus.Errorf("exitOnErrCh(1): %v", err)
		}
		cancel()
	}(ctx, errCh)
}
