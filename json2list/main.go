package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
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

	var useOnlyLowerCase bool
	flag.BoolVar(&useOnlyLowerCase, "lower", false, "")
	flag.BoolVar(&useOnlyLowerCase, "l", false,"")

	flag.Parse()

	//fmt.Println("All options parsed")

	const maxCapacity = 512 * 1024
	buf := make([]byte, maxCapacity)

	if isFlagPassed("i") || isFlagPassed("input") {
		buf = readJsonFileToByte(inputFile)
	} else {
		// fetch for all domains from stdin
		sc := bufio.NewScanner(os.Stdin)

		sc.Buffer(buf, maxCapacity)
	}

	parseJsonToWordList(buf, outputFile)
}

func parseJsonToWordList(buffer []byte, outputFile string) {
	file, err := os.Create(outputFile)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	var result map[string]interface{}
	e := json.Unmarshal(buffer, &result)

	// panic on error
	if e != nil {
		panic(e)
	}

	var entries []string
	var uniqueMap = make(map[string]bool)

	parseMap(result, &entries, uniqueMap)

	w := bufio.NewWriter(file)
	for _, line := range entries {
		fmt.Fprintln(w, line)
	}
	w.Flush()
}

func parseMap(aMap map[string]interface{}, entries *[]string, uniqueMap map[string]bool) {
	for key, val := range aMap {
		switch concreteVal := val.(type) {
		case map[string]interface{}:
			if checkForInclusion(key) {
				add(key, entries, uniqueMap)
			}
			parseMap(val.(map[string]interface{}), entries, uniqueMap)

		case []interface{}:
			if checkForInclusion(key) {
				add(key, entries, uniqueMap)
			}
			parseArray(val.([]interface{}), entries, uniqueMap)

		case nil:
			continue

		case string:
			if checkForInclusion(key) {
				add(key, entries, uniqueMap)
			}

			if checkForInclusion(concreteVal) {
				add(concreteVal, entries, uniqueMap)
			}

		default:
			continue
		}
	}
}

func parseArray(anArray []interface{}, entries *[]string, uniqueMap map[string]bool) {
	for _, val := range anArray {
		switch concreteVal := val.(type) {
		case map[string]interface{}:
			parseMap(val.(map[string]interface{}), entries, uniqueMap)

		case []interface{}:
			parseArray(val.([]interface{}), entries, uniqueMap)

		case nil:
			continue

		case string:
			if checkForInclusion(concreteVal) {
				//fmt.Println("Adding value from array", concreteVal)
				add(concreteVal, entries, uniqueMap)
			}

		default:
			continue
		}
	}
}

func add(entry string, entries *[]string, uniqueMap map[string]bool) {
	if strings.ToUpper(entry) != entry {
		entry = strings.ToLower(entry)
	}
	if uniqueMap[entry] {
		return // Already in the map
	}
	*entries = append(*entries, entry)
	uniqueMap[entry] = true
}

func checkForInclusion(content string) bool {

	// If it is empty it is not used as keyword
	if len(content) < 2 {
		return false
	}
	// If is just a number we don't use it as keyword
	if isNumeric(content) {
		return false
	}
	// If a white space is contained we don't use it as keyword
	if strings.Contains(content, " ") {
		return false
	}

	if strings.Contains(content, "/") || strings.Contains(content, ",") ||
		strings.Contains(content, "{") || strings.Contains(content, "}") ||
		strings.Contains(content, ":") || strings.Contains(content, "%") ||
		strings.Contains(content, "."){
		return false
	}

	if strings.Contains(content, "{") || strings.Contains(content, "}") {
		return false
	}

	if strings.HasPrefix(content, "#") {
		return false
	}

	if strings.Contains(content, "\u00e4") || strings.Contains(content, "\u00f6") ||
		strings.Contains(content, "\u00fc") || strings.Contains(content, "\u00D6") ||
		strings.Contains(content, "\u00DC") || strings.Contains(content, "\u00C4") ||
		strings.Contains(content, "\u00DF") {
		return false
	}

	if strings.Contains(content, "-") {
		found := false
		parts := strings.Split(content, "-")
		for _, part := range parts {
			if !isNumeric(part) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

func isNumeric(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

func readJsonFileToByte(jsonFile string) []byte {
	file, err := os.Open(jsonFile)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Successfully opened", jsonFile)
	// defer the closing of our jsonFile so that we can parse it later on
	defer file.Close()

	byteValue, _ := ioutil.ReadAll(file)

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
