/*
Usage:
> go build -o byok main.go
> ./byok --in config-in.yaml --addrs="0.0.0.0,0.0.0.1,0.0.0.2" --out config-out
Generate 3 files with broker address configs as addrs[0], addrs[1], addrs[2]
*/
package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	goyaml "github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	flag "github.com/spf13/pflag"
)

// Replaces the node in place with a string value.
func replaceStringAt(f *ast.File, path, value string) {
	yPath, _ := goyaml.PathString(path)
	oldNode, _ := yPath.FilterFile(f)
	newNode := &ast.StringNode{
		BaseNode: &ast.BaseNode{},
		Token:    oldNode.GetToken(),
		Value:    value,
	}

	if c := oldNode.GetComment(); c != nil {
		newNode.SetComment(c)
	}

	yPath.ReplaceWithNode(f, newNode)
}

// Replaces array at the place to array node.
func replaceArrayAt(f *ast.File, path string, values []string) {
	yPath, _ := goyaml.PathString(path)
	var buf bytes.Buffer
	node, _ := goyaml.NewEncoder(&buf, []goyaml.EncodeOption{
		goyaml.UseSingleQuote(false),
	}...).EncodeToNode(values)
	yPath.ReplaceWithNode(f, node)
}

// generate config after formatting server address and other broker addresses into ansible yaml configuration
func generateConfig(b []byte, brokerID int, addr string, addrs []string) []byte {
	f, err := parser.ParseBytes(b, parser.ParseComments)
	if err != nil {
		log.Fatalf("error unmarshalling configuration: %v", err)
	}

	replaceStringAt(f, "$.kafka_broker_id", strconv.Itoa(brokerID))
	replaceStringAt(f, "$.kafka_advertised_listeners[0]", fmt.Sprintf("INTERNAL://%s:9092", addr))
	replaceStringAt(f, "$.kafka_advertised_listeners[1]", fmt.Sprintf("BROKER://%s:9091", addr))

	zookeeperAddrs := make([]string, len(addrs))
	copy(zookeeperAddrs, addrs)
	for i := 0; i < len(zookeeperAddrs); i++ {
		zookeeperAddrs[i] += ":2888:3888"
	}
	a := strings.Join(zookeeperAddrs, ",")
	replaceStringAt(f, "$.kafka_zookeeper_connect", a)

	replaceArrayAt(f, "$.kafka_zookeeper_servers", zookeeperAddrs)

	kafkaAddrs := make([]string, len(addrs))
	copy(kafkaAddrs, addrs)
	for i := 0; i < len(kafkaAddrs); i++ {
		kafkaAddrs[i] += ":9092"
	}
	replaceArrayAt(f, "$.kaf_kakfa_servers", kafkaAddrs)

	// Go through documents and dump to []byte
	var docs [][]byte
	for _, doc := range f.Docs {
		docs = append(docs, []byte(doc.String()))
	}
	if len(docs) > 0 {
		return append(bytes.Join(docs, []byte("\n")), byte('\n'))
	}
	return []byte{}
}

func main() {
	// Parse command line flags
	flagSet := flag.NewFlagSet("config", flag.ContinueOnError)
	flagSet.Usage = func() {
		fmt.Println(flagSet.FlagUsages())
		os.Exit(0)
	}

	var (
		inFile, outFile string
		addrs           []string
	)
	flagSet.StringVar(&inFile, "in", "test.yaml", "Ansible kafka nodes configurations input")
	flagSet.StringSliceVar(&addrs, "addrs", []string{}, "Broker addresses")
	flagSet.StringVar(&outFile, "out", "", "Ansible kafka nodes configuration output files")

	err := flagSet.Parse(os.Args[1:])
	if err != nil {
		log.Fatalf("error parsing flags: %v", err)
	}

	b, err := os.ReadFile(inFile)
	if err != nil {
		log.Fatalf("error reading configuration file: %v", err)
	}

	// Validate flags
	if len(addrs) == 0 {
		log.Fatalf("--addrs cannot be left empty")
	}

	if outFile == "" {
		log.Fatalf("--out cannot be left empty")
	}

	var out []byte
	// Generate and write back to file.
	for i := 0; i < len(addrs); i++ {
		out = generateConfig(b, i+1, addrs[i], addrs)
		if err := os.WriteFile(fmt.Sprintf("%s.%d", outFile, i), out, 0644); err != nil {
			log.Fatalf("error writing to %s: %v", outFile, err)
		}

		// reset
		out = out[:0]
	}
}
