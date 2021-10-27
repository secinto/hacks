package json_to_wordlist

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func main() {
	var onlyKeys bool
	flag.BoolVar(&onlyKeys, "k", false, "")
	flag.BoolVar(&onlyKeys, "keys", false, "")

	var onlyValues bool
	flag.BoolVar(&onlyValues, "v", false, "")
	flag.BoolVar(&onlyValues, "values", false, "")

	var outputFile string
	flag.StringVar(&outputFile, "output", "wordlist.txt", "")
	flag.StringVar(&outputFile, "o", "wordlist.txt", "")

	var inputFile string
	flag.StringVar(&inputFile, "input", "", "")
	flag.StringVar(&inputFile, "i", "", "")

	flag.Parse()

	const maxCapacity = 512*1024
	buf := make([]byte, maxCapacity)

	if isFlagPassed("input") {
		buf := readJsonFileToByte(inputFile)
	} else {
		// fetch for all domains from stdin
		sc := bufio.NewScanner(os.Stdin)

		sc.Buffer(buf, maxCapacity)
	}
}

func parseJsonToWordList(buffer []byte, outputFile string) {
	file, err := os.Create(outputFile)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	w := bufio.NewWriter(file)

	writtenBytes, err := w.Write(buffer)

	fmt.Printf("Wrote %d byte to file %s", writtenBytes, outputFile)

	w.Flush()

}

func readJsonFileToByte(jsonFile string) []byte {
	file, err := os.Open(jsonFile)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Successfully Opened users.json")
	// defer the closing of our jsonFile so that we can parse it later on
	defer file.Close()

	byteValue, _ := ioutil.ReadAll(file)

	var result map[string]interface{}
	json.Unmarshal([]byte(byteValue), &result)
	return byteValue
}

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func init() {
	flag.Usage = func() {
		h := []string{
			"Create a wordlist from the provided JSON file. Per default all keys and values are used.",
			"",
			"Options:",
			"  -i, --input <file>        JSON input file to use",
			"  -k, --keys                Use only keys for the wordlist",
			"  -v, --values              Use only the values for the wordlist",
			"  -o, --output <file        File to store the created wordlist (will be created)",
			"",
		}

		fmt.Fprintf(os.Stderr, strings.Join(h, "\n"))
	}
}