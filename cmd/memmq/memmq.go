package memmq

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/hellobchain/memmq/broker"
	mqclient "github.com/hellobchain/memmq/client"
	mqgrpc "github.com/hellobchain/memmq/client/grpc"
	mqresolver "github.com/hellobchain/memmq/client/resolver"
	mqselector "github.com/hellobchain/memmq/client/selector"
	"github.com/hellobchain/memmq/server"
	grpcsrv "github.com/hellobchain/memmq/server/grpc"
	httpsrv "github.com/hellobchain/memmq/server/http"
	"github.com/hellobchain/wswlog/wlogging"
)

var logger = wlogging.MustGetLoggerWithoutName()

var (
	// MQ server address
	address = flag.String("address", ":8081", "MQ server address")
	// TLS certificate
	cert = flag.String("cert_file", "", "TLS certificate file")
	// TLS key
	key = flag.String("key_file", "", "TLS key file")

	// server persist to file
	persist = flag.Bool("persist", false, "Persist messages to [topic].mq file per topic")

	// proxy flags
	proxy   = flag.Bool("proxy", false, "Proxy for an MQ cluster")
	retries = flag.Int("retries", 1, "Number of retries for publish or subscribe")
	servers = flag.String("servers", "", "Comma separated MQ cluster list used by Proxy")

	// client flags
	interactive = flag.Bool("i", false, "Interactive client mode")
	client      = flag.Bool("client", false, "Run the MQ client")
	publish     = flag.Bool("publish", false, "Publish via the MQ client")
	subscribe   = flag.Bool("subscribe", false, "Subscribe via the MQ client")
	topic       = flag.String("topic", "", "Topic for client to publish or subscribe to")

	// select strategy
	selector = flag.String("select", "all", "Server select strategy. Supports all, shard")
	// resolver for discovery
	resolver = flag.String("resolver", "ip", "Server resolver for discovery. Supports ip, dns")
	// transport http or grpc
	transport = flag.String("transport", "http", "Transport for communication. Support http, grpc")
)

func init() {
	flag.Parse()

	if *proxy && *client {
		logger.Fatal("Client and proxy flags cannot be specified together")
	}

	if *proxy && len(*servers) == 0 {
		logger.Fatal("Proxy enabled without MQ server list")
	}

	if *client && len(*topic) == 0 {
		logger.Fatal("Topic not specified")
	}

	if *client && !*publish && !*subscribe {
		logger.Fatal("Specify whether to publish or subscribe")
	}

	if (*client || *interactive) && len(*servers) == 0 {
		*servers = "localhost:8081"
	}

	var bclient mqclient.Client
	var selecter mqclient.Selector
	var resolvor mqclient.Resolver

	switch *selector {
	case "shard":
		selecter = new(mqselector.Shard)
	default:
		selecter = new(mqselector.All)
	}

	switch *resolver {
	case "dns":
		resolvor = new(mqresolver.DNS)
	default:
	}

	options := []mqclient.Option{
		mqclient.WithResolver(resolvor),
		mqclient.WithSelector(selecter),
		mqclient.WithServers(strings.Split(*servers, ",")...),
		mqclient.WithRetries(*retries),
	}

	switch *transport {
	case "grpc":
		bclient = mqgrpc.New(options...)
	default:
		bclient = mqclient.New(options...)
	}

	broker.Default = broker.New(
		broker.Client(bclient),
		broker.Persist(*persist),
		broker.Proxy(*client || *proxy || *interactive),
	)
}

func cli() {
	wg := sync.WaitGroup{}
	p := make(chan []byte, 1000)
	d := map[string]time.Time{}
	ttl := time.Millisecond * 10
	tick := time.NewTicker(time.Second * 5)

	// process publish
	if *publish || *interactive {
		wg.Add(1)
		go func() {
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				if *interactive {
					p <- scanner.Bytes()
				}
				broker.Publish(*topic, scanner.Bytes())
			}
			wg.Done()
		}()
	}

	// subscribe?
	if !(*subscribe || *interactive) {
		wg.Wait()
		return
	}

	// process subscribe
	ch, err := broker.Subscribe(*topic)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer broker.Unsubscribe(*topic, ch)

	for {
		select {
		// process sub event
		case b := <-ch:
			// skip if deduped
			if t, ok := d[string(b)]; ok && time.Since(t) < ttl {
				continue
			}
			d[string(b)] = time.Now()
			fmt.Println(string(b))
		// add dedupe entry
		case b := <-p:
			d[string(b)] = time.Now()
		// flush deduper
		case <-tick.C:
			d = map[string]time.Time{}
		}
	}
}

func StartMain() bool {
	// handle client
	if *client || *interactive {
		cli()
		return true
	}
	// cleanup broker
	defer broker.Default.Close()
	options := []server.Option{
		server.WithAddress(*address),
	}
	// proxy enabled
	if *proxy {
		logger.Info("Proxy enabled")
	}
	// tls enabled
	if len(*cert) > 0 && len(*key) > 0 {
		logger.Info("TLS Enabled")
		options = append(options, server.WithTLS(*cert, *key))
	}
	var server server.Server
	// now serve the transport
	switch *transport {
	case "grpc":
		logger.Info("GRPC transport enabled")
		server = grpcsrv.New(options...)
	default:
		logger.Info("HTTP transport enabled")
		server = httpsrv.New(options...)
	}
	logger.Info("MQ listening on", *address)
	go func() {
		if err := server.Run(); err != nil {
			logger.Fatal(err)
		}
	}()
	return false
}
