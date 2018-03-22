package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type Strd struct {
	XMLName xml.Name `xml:"Strd"`
	Cd      string   `xml:"RfrdDocInf>Tp>CdOrPrtry>Cd"`
	Nb      string   `xml:"RfrdDocInf>Nb"`
	RltdDt  string   `xml:"RfrdDocInf>RltdDt"`
	RmtdAmt RmtdAmt  `xml:"RmtdAmt"`
}

type RmtdAmt struct {
	RmtdAmt string `xml:",chardata"`
	RAName  string `xml:"name,attr"`
}

type Document struct {
	XMLName  xml.Name `xml:"Document"`
	Chequeno string   `xml:"CstmrCdtTrfInitn>PmtInf>CdtTrfTxInf>ChqInstr>ChqNb"`
}

func main() {
	// Only one XML files in the directory find it and parse it
	xmlfiles, err := filepath.Glob("*.xml")
	if err != nil {
		log.Fatal(err)
	}
	xmlFile, err := os.Open(xmlfiles[0])
	if err != nil {
		fmt.Println(err)
	}
	defer xmlFile.Close()

	// read our opened xmlFile as a byte array.
	byteValue, _ := ioutil.ReadAll(xmlFile)

	// we initialize our Document struct
	var doc Document

	// we unmarshal our byteArray which contains our
	// xmlFiles content into 'doc' to get the cheque number
	xml.Unmarshal(byteValue, &doc)

	// Now we find the only csv file in the folder
	csvfiles, err := filepath.Glob("*.csv")
	if err != nil {
		log.Fatal(err)
	}

	// Open the file and parse each line
	csvFile, _ := os.Open(csvfiles[0])
	reader := csv.NewReader(bufio.NewReader(csvFile))
	var x []Strd
	var st Strd
	for {
		line, error := reader.Read()
		if error == io.EOF {
			break
		} else if error != nil {
			log.Fatal(error)
		}
		// if the cheque number in the csv matches the cheque number we want
		if line[0] == doc.Chequeno {
			// Create the xml we want to insert and add it to the x array
			st = Strd{
				Nb:     line[0],
				RltdDt: line[1],
				Cd:     "CINV",
			}
			st.RmtdAmt = RmtdAmt{RAName: "Ccy", RmtdAmt: line[2]}
			x = append(x, st)
		}
	}

	// Open the xml file in append mode
	f, err := os.OpenFile(xmlfiles[0], os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	defer f.Close()

	// Space out the xml properly
	output, err := xml.MarshalIndent(x, "", "\t")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	// Convert our string to bytes and split it on every new line
	y := []byte(output)
	z := bytes.SplitAfter(y, []byte("\n"))

	// for eaech new like add 4 tabs
	var c []byte
	for i := 0; i < len(z); i++ {
		c = append(c, z[i]...)
		c = append(c, []byte("\t\t\t\t")...)
	}

	// find where we want to insert the new xml
	findval := []byte("</Cdtr>")
	splitstr := bytes.SplitAfter(byteValue, findval)
	newline := []byte("\n\t\t\t\t")

	// now concat the bytes back together with our bit added
	byteValue = append(splitstr[0], newline...)
	byteValue = append(byteValue, c[:len(c)-4]...)
	byteValue = append(byteValue, splitstr[1]...)

	// Save it back down to a new file
	err = ioutil.WriteFile("Result.xml", byteValue, 0644)

}
