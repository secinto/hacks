# json2list

<h4 align="center">Create wordlists from JSON data.</h4>
     
<p align="center">
  <a href="#usage">Usage</a> •
  <a href="#Installation">Installation</a> •
</p>

---


json2list is a simple tool for creating target specific wordlists from JSON input data. json2list is written in Go.
Parses JSON data and uses keys and values which seem to be variables or special names as entries for a wordlist.  

# Usage

```sh
json2tool -h
```
This will display help for the tool. It is also possible to provide the input as piped input via stdin. 

For example.
```sh
cat input.json | json2tool -o wordlist.txt
```
Here are all the switches it supports.
| Flag             | Description                                                | Example                                        |
| ---------------- | ---------------------------------------------------------- | -----------------------------------------------|
| -input / -i      | IP address to be used as local bind                        | json2list -i input.json                        |
| -keys            | Use only key from the JSON as input                        | json2list -keys -i input.json                  |
| -values          | Use only values from the JSON as input                     | json2list -values -i input.json                |
| -output / -o     | Write output to specified file. Will be created            | json2list -output wordlist.txt -i input.json   |
| -lower / -l      | Use only lower case entries                                | json2list -lower -i input.json                 |
| -v               | Show Verbose output                                        | json2list -v                                   |
| -version         | Show current program version                               | json2list -vers   ion                          |


# Installation

json2list requires **go1.17** to install successfully. Run the following command to get the repo -

```sh
go install -v github.com/secinto/hacks/json2list@latest
```
