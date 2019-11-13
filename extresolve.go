package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/InVisionApp/tabular"
	"github.com/swaggo/cli"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

var app = cli.NewApp()
var tab tabular.Table

type DNS []struct {
	IP          string      `json:"ip"`
	Name        string      `json:"name"`
	CountryID   string      `json:"country_id"`
	City        string      `json:"city"`
	Version     string      `json:"version"`
	Error       interface{} `json:"error"`
	Dnssec      bool        `json:"dnssec"`
	Reliability float64     `json:"reliability"`
	CheckedAt   time.Time   `json:"checked_at"`
	CreatedAt   time.Time   `json:"created_at"`
}

func init() {
	// Cli settings
	app.Name = "Dns check"
	app.Version = "1.0"
	app.Usage = "Checking current host in all public dns in current geo area.\r\n" +
		"\t Uses public-dns.info for getting dns servers in selected geo area."
	app.UsageText = "check_dns -n {host} -r {geo_area}"
	app.HideHelp = false
	app.Compiled = time.Now()
	app.Authors = []cli.Author{
		{
			Name:  "Dmitriy Vlassov",
			Email: "dmitriy@vlassov.kz",
		},
	}
	app.Flags = []cli.Flag{
		cli.Int64Flag{
			Name:  "count, c",
			Value: 10,
			Usage: "quantity checks",
		},
		cli.StringFlag{
			Name:  "protocol",
			Value: "udp",
			Usage: "connect protocol (udp or tcp)",
		},
		cli.StringFlag{
			Name:     "host, n",
			Value:    "",
			Required: true,
			Usage:    "host for checks",
		},
		cli.StringFlag{
			Name:     "region, r",
			Value:    "",
			Required: true,
			Usage:    "region for checks",
		},
		cli.Int64Flag{
			Name:  "timeout, t",
			Value: 3,
			Usage: "timeout for check",
		},
		cli.Int64Flag{
			Name:  "timeout-get-dns-list",
			Value: 30,
			Usage: "timeout for get dns from public-dns.info service",
		},
		cli.BoolFlag{
			Name:  "detail, d",
			Usage: "more info about dns servers",
		},
	}
	// Tabular settings
	tab = tabular.New()
	tab.Col("DNS", "DNS server", 20)
	tab.Col("Name", "DNS name", 60)
	tab.Col("City", "DNS city", 20)
	tab.Col("Dnssec", "DNS security", 12)
	tab.Col("Reliability", "DNS Reliability", 15)
	tab.Col("Result", "DNS responce", 0)
}

func chechHost(c *cli.Context) {
	count := c.Int("count")

	protocol := c.String("protocol")

	host := c.String("host")

	geo := c.String("region")

	timeout := c.Int("timeout")

	timeoutGetDns := c.Int("timeout-get-dns-list")

	detail := c.Bool("detail")

	var httpClient = &http.Client{
		Timeout: time.Duration(timeoutGetDns) * time.Second,
	}
	resp, err := httpClient.Get("https://public-dns.info/nameserver/" + geo + ".json")
	if err != nil {
		log.Fatal(err)
	}
	f := func(bool) string {
		if detail {
			return tab.Print("DNS", "Name", "City", "Dnssec", "Reliability", "Result")
		} else {
			return tab.Print("DNS", "Name", "City", "Result")
		}
	}
	format := f(detail)

	if body, err := ioutil.ReadAll(resp.Body); err == nil {
		var dnsHosts = DNS{}
		if err = json.Unmarshal(body, &dnsHosts); err == nil {
			for id, row := range dnsHosts {
				var resolver *net.Resolver
				resolver = &net.Resolver{
					PreferGo: true,
					Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
						d := net.Dialer{
							Timeout: time.Duration(timeout) * time.Second,
						}
						return d.DialContext(ctx, protocol, net.JoinHostPort(row.IP, "53"))
					},
				}
				ips, _ := resolver.LookupIPAddr(context.Background(), host)
				var resolveHost = ""
				for _, ip := range ips {
					resolveHost = resolveHost + " " + ip.String()
				}
				if detail {
					fmt.Printf(format, row.IP, row.Name, row.City, row.Dnssec, row.Reliability, resolveHost)
				} else {
					fmt.Printf(format, row.IP, row.Name, row.City, resolveHost)
				}

				if (id + 1) == count {
					break
				}
			}
		} else {
			log.Fatal(err)
		}
	} else {
		log.Fatal(err)
	}
	defer resp.Body.Close()
}

func main() {
	app.Action = func(c *cli.Context) {
		chechHost(c)
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
