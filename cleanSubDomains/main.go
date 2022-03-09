package main

import (
	"flag"
	log "github.com/sirupsen/logrus"
	"github.com/thedevsaddam/gojsonq/v2"
)

func main() {
	log.SetFormatter(&log.JSONFormatter{})

	var dupCount int
	flag.IntVar(&dupCount, "dc", 10, "Duplicate count. Maximum amount of duplicates allowed per input")

	var inFile string
	flag.StringVar(&inFile, "i", "httpx_output.json", "Path and name of the input JSON file as created from (pd) httpx")

	var outFile string
	flag.StringVar(&outFile, "o", "domains_purified.txt", "Path and name of the output file to write")

	var host string
	flag.StringVar(&host, "h", "", "Host name or IP which should be filtered for")

	flag.Parse()

	parseJsonToWordList(inFile, outFile)
}

func parseJsonToWordList(inputFile string, outputFile string) {

	jq := gojsonq.New().File(inputFile)
	log.Debug(jq)

	jq.From("hosts").Where("length")

}
