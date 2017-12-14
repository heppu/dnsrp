package main

import (
	"flag"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/BurntSushi/toml"
	"github.com/coreos/go-log/log"
	"github.com/miekg/dns"
)

// Config contains name server rules
type Config struct {
	DefaultServer string
	Servers       map[string][]string
}

var (
	defaultServer string
	rules         map[string]string
)

func init() {
	cfgFile := flag.String("c", "", "Config file")
	flag.Parse()
	config := Config{}
	if _, err := toml.DecodeFile(*cfgFile, &config); err != nil {
		log.Fatalln(err)
		return
	}
	if config.DefaultServer == "" {
		log.Fatalln("Empty defaultServer in", *cfgFile)
	}

	defaultServer = config.DefaultServer
	rules = map[string]string{}
	for k, v := range config.Servers {
		for _, domain := range v {
			rules[domain+"."] = k
		}
	}
}

func main() {
	log.Println("Starting")
	tcpSrv := &dns.Server{
		Addr: "127.0.0.1:53",
		Net:  "tcp",
		Handler: dns.HandlerFunc(
			func(w dns.ResponseWriter, m *dns.Msg) {
				handler(w, m, &dns.Client{Net: "tcp"})
			},
		),
	}

	udpSrv := &dns.Server{
		Addr: "127.0.0.1:53",
		Net:  "udp",
		Handler: dns.HandlerFunc(
			func(w dns.ResponseWriter, m *dns.Msg) {
				handler(w, m, &dns.Client{Net: "udp"})
			},
		),
	}

	go func() {
		if err := tcpSrv.ListenAndServe(); err != nil {
			log.Fatalln(err)
		}
	}()

	go func() {
		if err := udpSrv.ListenAndServe(); err != nil {
			log.Fatalln(err)
		}
	}()

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	log.Println("\nStopping...")
	if err := tcpSrv.Shutdown(); err != nil {
		log.Error("Error while stopping tcp server:", err)
	}
	if err := udpSrv.Shutdown(); err != nil {
		log.Error("Error while stopping udp server:", err)
	}
}

func handler(w dns.ResponseWriter, m *dns.Msg, c *dns.Client) {
	ip := defaultServer
	for domain, newIP := range rules {
		if strings.HasSuffix(m.Question[0].Name, domain) {
			ip = newIP
			break
		}
	}

	log.Debugf("Lookup %s from %s\n", m.Question[0].Name, ip)

	r, _, err := c.Exchange(m, ip+":53")
	if err != nil {
		log.Error(err)
		return
	}

	w.WriteMsg(r)
}
